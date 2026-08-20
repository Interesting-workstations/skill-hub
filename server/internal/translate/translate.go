// Package translate 提供技能标题/描述的中文翻译能力（在线翻译 API）。
//
// Provider 支持：
//   - tencent（腾讯云机器翻译 TMT）：需要 TX_SECRETID + TX_SECRETKEY（每月免费额度）
//   - baidu：需要 BAIDU_APPID + BAIDU_KEY（有免费额度）
//   - google（默认）：无需 Key，调用 translate.googleapis.com 免费接口（质量较好）
//   - deepl：需要 DEEPL_KEY
//
// 多个 provider 可链式降级：当前 provider 失败时自动按顺序尝试下一个，哪个能用用哪个。
// 翻译结果带进程内缓存，避免重复翻译同一文本。
package translate

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// Translator 文本翻译器。
type Translator struct {
	// providers 按优先级排列的翻译通道（tencent/baidu/deepl/google），
	// 依次尝试，当前失败自动降级下一个。
	providers []string

	baiduApp    string
	baiduKey    string
	deeplKey    string
	txSecretID  string
	txSecretKey string
	http        *http.Client

	broken       bool // 熔断标记：翻译服务网络不可达时置位，后续跳过翻译直接返回原文
	googleBroken bool // Google 不可达标记（中国大陆被墙），避免每次降级都等 12s 超时

	last *lastSuccess // 最近一次翻译成功的通道（后台状态展示）

	mu    sync.Mutex
	cache map[string]string // key: toLang + "|" + text
}

// providerOrder 默认 provider 优先级（可被 TRANSLATE_PROVIDER 覆盖）。
// 腾讯云免费额度优先，其次百度，再 Google，最后 DeepL。
var providerOrder = []string{"tencent", "baidu", "google", "deepl"}

// New 创建翻译器，配置从环境变量读取：
//   - TRANSLATE_PROVIDER=tencent|baidu|deepl|google|off（默认按 tencent>baidu>google>deepl 自动降级）
//   - TX_SECRETID / TX_SECRETKEY（tencent 需要）
//   - BAIDU_APPID / BAIDU_KEY（baidu 需要）
//   - DEEPL_KEY（deepl 需要）
func New() *Translator {
	configured := strings.ToLower(os.Getenv("TRANSLATE_PROVIDER"))
	providers := make([]string, 0, len(providerOrder))
	if configured != "" && configured != "auto" && configured != "off" && configured != "none" {
		// 用户显式指定：单通道（保持向后兼容），失败仍可降级后续通道
		providers = append(providers, configured)
		for _, p := range providerOrder {
			if p != configured {
				providers = append(providers, p)
			}
		}
	} else {
		providers = append(providers, providerOrder...)
	}
	// 过滤：未配置对应密钥的通道移除（google 始终可用）
	filtered := make([]string, 0, len(providers))
	for _, p := range providers {
		switch p {
		case "tencent":
			if os.Getenv("TX_SECRETID") != "" && os.Getenv("TX_SECRETKEY") != "" {
				filtered = append(filtered, p)
			}
		case "baidu":
			if os.Getenv("BAIDU_APPID") != "" && os.Getenv("BAIDU_KEY") != "" {
				filtered = append(filtered, p)
			}
		case "deepl":
			if os.Getenv("DEEPL_KEY") != "" {
				filtered = append(filtered, p)
			}
		default:
			filtered = append(filtered, p)
		}
	}
	if len(filtered) == 0 {
		filtered = []string{"google"}
	}
	return &Translator{
		providers:   filtered,
		baiduApp:    os.Getenv("BAIDU_APPID"),
		baiduKey:    os.Getenv("BAIDU_KEY"),
		deeplKey:    os.Getenv("DEEPL_KEY"),
		txSecretID:  os.Getenv("TX_SECRETID"),
		txSecretKey: os.Getenv("TX_SECRETKEY"),
		http:        &http.Client{Timeout: 12 * time.Second},
		cache:       make(map[string]string),
	}
}

// Providers 返回当前生效的 provider 通道列表（按优先级）。
func (t *Translator) Providers() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]string(nil), t.providers...)
}

// SetProviders 动态设置 provider 链（后台管理切换用，运行时立即生效）。
func (t *Translator) SetProviders(providers []string) {
	ps := make([]string, 0, len(providers))
	for _, p := range providers {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		// 过滤未配置密钥的通道
		switch p {
		case "tencent":
			if t.txSecretID == "" || t.txSecretKey == "" {
				continue
			}
		case "baidu":
			if t.baiduApp == "" || t.baiduKey == "" {
				continue
			}
		case "deepl":
			if t.deeplKey == "" {
				continue
			}
		}
		ps = append(ps, p)
	}
	if len(ps) == 0 {
		ps = []string{"google"}
	}
	t.mu.Lock()
	t.providers = ps
	t.broken = false // 配置变更后重置熔断
	t.mu.Unlock()
}

// lastSuccess 最近一次翻译成功的通道（后台状态展示用）。
type lastSuccess struct {
	provider string
	at       time.Time
}

// TestProvider 用单个指定通道翻译一段文本，验证该通道是否可用。
// 不修改熔断状态、不写缓存，供后台「测试连通性」使用。
func (t *Translator) TestProvider(provider, text, toLang string) (out string, elapsed time.Duration, err error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	start := time.Now()
	switch provider {
	case "tencent":
		if t.txSecretID == "" || t.txSecretKey == "" {
			return "", time.Since(start), fmt.Errorf("未配置 TX_SECRETID / TX_SECRETKEY")
		}
		out, err = t.tencent(text, toLang)
	case "baidu":
		if t.baiduApp == "" || t.baiduKey == "" {
			return "", time.Since(start), fmt.Errorf("未配置 BAIDU_APPID / BAIDU_KEY")
		}
		out, err = t.baidu(text, toLang)
	case "deepl":
		if t.deeplKey == "" {
			return "", time.Since(start), fmt.Errorf("未配置 DEEPL_KEY")
		}
		out, err = t.deepl(text, toLang)
	case "google", "":
		out, err = t.google(text, toLang)
	default:
		return "", time.Since(start), fmt.Errorf("未知通道: %s", provider)
	}
	return out, time.Since(start), err
}

// ProviderStatus 各通道配置/可用状态（后台管理页展示）。
func (t *Translator) ProviderStatus() map[string]bool {
	return map[string]bool{
		"tencent": t.txSecretID != "" && t.txSecretKey != "",
		"baidu":   t.baiduApp != "" && t.baiduKey != "",
		"deepl":   t.deeplKey != "",
		"google":  true,
	}
}

// Enabled 是否启用翻译（off/none 禁用）。
func (t *Translator) Enabled() bool {
	for _, p := range t.providers {
		if p == "off" || p == "none" {
			return false
		}
	}
	return len(t.providers) > 0
}

// Translate 翻译单段文本到目标语言（zh / en）。失败时返回原文（降级，不阻断主流程）。
func (t *Translator) Translate(text, toLang string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return text
	}
	key := toLang + "|" + text
	t.mu.Lock()
	if t.broken {
		t.mu.Unlock()
		return text
	}
	if v, ok := t.cache[key]; ok {
		t.mu.Unlock()
		return v
	}
	t.mu.Unlock()

	out, err := t.translate(text, toLang)
	if err != nil {
		// 失败降级 + 熔断：翻译服务不可达（如中国大陆访问 Google 被墙）时标记 broken，
		// 后续批量文本直接返回原文，避免逐条等待 12s 超时导致爬虫"卡在翻译阶段"
		t.mu.Lock()
		t.broken = true
		t.mu.Unlock()
		return text
	}
	t.mu.Lock()
	t.cache[key] = out
	t.mu.Unlock()
	return out
}

// TranslateStrict 翻译单段文本，失败返回错误（不静默降级为原文）。
// 供翻译管理页使用：翻译失败时能明确报错，避免"显示成功但内容未汉化"。
// 不受 broken 熔断影响（每次真实尝试，管理员可修复配置后重试）。
func (t *Translator) TranslateStrict(text, toLang string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return text, nil
	}
	key := toLang + "|" + text
	t.mu.Lock()
	if v, ok := t.cache[key]; ok {
		t.mu.Unlock()
		return v, nil
	}
	t.mu.Unlock()

	out, err := t.translate(text, toLang)
	if err != nil {
		return text, err
	}
	t.mu.Lock()
	t.cache[key] = out
	t.mu.Unlock()
	return out, nil
}

// TranslateBatch 批量翻译（串行带缓存，避免并发打爆免费接口）。
func (t *Translator) TranslateBatch(texts []string, toLang string) []string {
	out := make([]string, len(texts))
	for i, s := range texts {
		out[i] = t.Translate(s, toLang)
	}
	return out
}

// translate 按 provider 优先级依次尝试，当前失败自动降级下一个；全部失败返回错误。
// 已熔断的通道（如 Google 被墙）会快速跳过。
func (t *Translator) translate(text, toLang string) (string, error) {
	var errs []string
	for _, p := range t.providers {
		switch p {
		case "tencent":
			out, err := t.tencent(text, toLang)
			if err == nil {
				t.recordSuccess("tencent")
				return out, nil
			}
			errs = append(errs, "tencent: "+err.Error())
		case "baidu":
			out, err := t.baidu(text, toLang)
			if err == nil {
				t.recordSuccess("baidu")
				return out, nil
			}
			errs = append(errs, "baidu: "+err.Error())
		case "deepl":
			out, err := t.deepl(text, toLang)
			if err == nil {
				t.recordSuccess("deepl")
				return out, nil
			}
			errs = append(errs, "deepl: "+err.Error())
		default: // google
			t.mu.Lock()
			gBroken := t.googleBroken
			t.mu.Unlock()
			if gBroken {
				errs = append(errs, "google: 不可达（已熔断）")
				continue
			}
			out, err := t.google(text, toLang)
			if err == nil {
				t.recordSuccess("google")
				return out, nil
			}
			t.mu.Lock()
			t.googleBroken = true
			t.mu.Unlock()
			errs = append(errs, "google: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return "", fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return "", fmt.Errorf("无可用翻译通道")
}

// recordSuccess 记录最近一次成功的通道（后台状态展示）。
func (t *Translator) recordSuccess(p string) {
	t.mu.Lock()
	t.last = &lastSuccess{provider: p, at: time.Now()}
	t.mu.Unlock()
}

// LastSuccess 返回最近一次翻译成功的通道名（无则空串）。
func (t *Translator) LastSuccess() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.last == nil {
		return ""
	}
	return t.last.provider
}

// google 免费接口（client=gtx），无需 Key。
func (t *Translator) google(text, toLang string) (string, error) {
	q := url.Values{}
	q.Set("client", "gtx")
	q.Set("sl", "auto")
	q.Set("tl", toLang)
	q.Set("dt", "t")
	q.Set("q", text)
	resp, err := t.http.Get("https://translate.googleapis.com/translate_a/single?" + q.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google translate status=%d", resp.StatusCode)
	}
	var data []any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", fmt.Errorf("google translate empty result")
	}
	segs, _ := data[0].([]any)
	var sb strings.Builder
	for _, seg := range segs {
		part, _ := seg.([]any)
		if len(part) > 0 {
			if s, ok := part[0].(string); ok {
				sb.WriteString(s)
			}
		}
	}
	if sb.Len() == 0 {
		return "", fmt.Errorf("google translate empty text")
	}
	return sb.String(), nil
}

// baidu 通用翻译 API（需 appid + key）。
// 百度 `to` 参数只接受 zh/en 等简码（不接受 zh-CN），这里做归一化。
func (t *Translator) baidu(text, toLang string) (string, error) {
	salt := fmt.Sprintf("%d", time.Now().UnixNano())
	sign := md5.Sum([]byte(t.baiduApp + text + salt + t.baiduKey))
	form := url.Values{}
	form.Set("q", text)
	form.Set("from", "auto")
	// 归一化语言码：zh-CN → zh（百度不支持带区域后缀）
	to := toLang
	if strings.HasPrefix(to, "zh") {
		to = "zh"
	} else if strings.HasPrefix(to, "en") {
		to = "en"
	}
	form.Set("to", to)
	form.Set("appid", t.baiduApp)
	form.Set("salt", salt)
	form.Set("sign", hex.EncodeToString(sign[:]))
	resp, err := t.http.PostForm("https://fanyi-api.baidu.com/api/trans/vip/translate", form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data struct {
		ErrorCode string `json:"error_code"`
		Trans     []struct {
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.ErrorCode != "" && data.ErrorCode != "0" {
		return "", fmt.Errorf("baidu error=%s", data.ErrorCode)
	}
	if len(data.Trans) == 0 {
		return "", fmt.Errorf("baidu empty result")
	}
	return data.Trans[0].Dst, nil
}

// tencent 腾讯云机器翻译 TMT（TextTranslate）。
// 使用 TC3-HMAC-SHA256 签名（腾讯云 V3 签名），POST JSON 到 tmt.tencentcloudapi.com。
// 语言码归一化：zh-CN → zh、en-* → en（TMT 用 zh/en 简码）。
func (t *Translator) tencent(text, toLang string) (string, error) {
	to := normalizeTencentLang(toLang)
	src := "auto"
	if strings.HasPrefix(strings.ToLower(text), "zh") || containsCJK(text) {
		src = "zh"
	}
	payload, _ := json.Marshal(map[string]any{
		"SourceText": text,
		"Source":     src,
		"Target":     to,
		"ProjectId":  0,
	})

	const (
		service   = "tmt"
		host      = "tmt.tencentcloudapi.com"
		action    = "TextTranslate"
		version   = "2018-03-21"
		algorithm = "TC3-HMAC-SHA256"
		region    = "ap-guangzhou"
	)
	now := time.Now()
	timestamp := now.Unix()
	date := now.UTC().Format("2006-01-02")

	// 1. CanonicalRequest
	hashedPayload := sha256hex(string(payload))
	canonicalHeaders := "content-type:application/json; charset=utf-8\nhost:" + host + "\n"
	signedHeaders := "content-type;host"
	canonicalRequest := fmt.Sprintf("POST\n/\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, hashedPayload)

	// 2. StringToSign
	credentialScope := fmt.Sprintf("%s/%s/tc3_request", date, service)
	hashedCanonical := sha256hex(canonicalRequest)
	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s", algorithm, timestamp, credentialScope, hashedCanonical)

	// 3. Signature
	secretDate := hmacSHA256([]byte("TC3"+t.txSecretKey), date)
	secretService := hmacSHA256(secretDate, service)
	secretSigning := hmacSHA256(secretService, "tc3_request")
	signature := hex.EncodeToString(hmacSHA256(secretSigning, stringToSign))

	// 4. Authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm, t.txSecretID, credentialScope, signedHeaders, signature)

	req, err := http.NewRequest(http.MethodPost, "https://"+host+"/", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", host)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Region", region)
	req.Header.Set("Authorization", authorization)

	resp, err := t.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tencent status=%d body=%s", resp.StatusCode, truncate(string(body), 200))
	}
	var data struct {
		Response struct {
			TargetText string `json:"TargetText"`
			Error      *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
			RequestID string `json:"RequestId"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.Response.Error != nil {
		return "", fmt.Errorf("tencent error=%s %s", data.Response.Error.Code, data.Response.Error.Message)
	}
	if data.Response.TargetText == "" {
		return "", fmt.Errorf("tencent empty result")
	}
	return data.Response.TargetText, nil
}

// normalizeTencentLang TMT 语言码归一化（zh-CN → zh，en-* → en）。
func normalizeTencentLang(l string) string {
	l = strings.ToLower(l)
	switch {
	case strings.HasPrefix(l, "zh"):
		return "zh"
	case strings.HasPrefix(l, "en"):
		return "en"
	case strings.HasPrefix(l, "ja"):
		return "ja"
	case strings.HasPrefix(l, "ko"):
		return "ko"
	case strings.HasPrefix(l, "fr"):
		return "fr"
	case strings.HasPrefix(l, "de"):
		return "de"
	case strings.HasPrefix(l, "es"):
		return "es"
	case strings.HasPrefix(l, "ru"):
		return "ru"
	default:
		return l
	}
}

// sha256hex SHA-256 后 hex 编码。
func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// hmacSHA256 HMAC-SHA256（key []byte, data string）。
func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(data))
	return mac.Sum(nil)
}

// truncate 截断字符串到 n 字节（用于错误信息展示，避免刷屏）。
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// containsCJK 判断字符串是否包含中文字符（TMT 源语言自动识别辅助）。
func containsCJK(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// deepl 翻译 API（需 DEEPL_KEY，免费版 api-free）。
func (t *Translator) deepl(text, toLang string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"text":        []string{text},
		"target_lang": toLang,
	})
	req, err := http.NewRequest(http.MethodPost, "https://api-free.deepl.com/v2/translate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+t.deeplKey)
	resp, err := t.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepl status=%d", resp.StatusCode)
	}
	var data struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Translations) == 0 {
		return "", fmt.Errorf("deepl empty result")
	}
	return data.Translations[0].Text, nil
}
