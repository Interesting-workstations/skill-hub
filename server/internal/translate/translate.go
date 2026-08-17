// Package translate 提供技能标题/描述的中文翻译能力（在线翻译 API）。
//
// Provider 支持：
//   - google（默认）：无需 Key，调用 translate.googleapis.com 免费接口（质量较好）
//   - baidu：需要 BAIDU_APPID + BAIDU_KEY（有免费额度）
//   - deepl：需要 DEEPL_KEY
//
// 翻译结果带进程内缓存，避免重复翻译同一文本。
package translate

import (
	"bytes"
	"crypto/md5"
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
	provider string // google / baidu / deepl / off
	baiduApp string
	baiduKey string
	deeplKey string
	http     *http.Client

	broken bool // 熔断标记：翻译服务网络不可达时置位，后续跳过翻译直接返回原文

	mu    sync.Mutex
	cache map[string]string // key: toLang + "|" + text
}

// New 创建翻译器，配置从环境变量读取：
//   - TRANSLATE_PROVIDER=google|baidu|deepl（默认 google）
//   - BAIDU_APPID / BAIDU_KEY（baidu 需要）
//   - DEEPL_KEY（deepl 需要）
func New() *Translator {
	return &Translator{
		provider: strings.ToLower(os.Getenv("TRANSLATE_PROVIDER")),
		baiduApp: os.Getenv("BAIDU_APPID"),
		baiduKey: os.Getenv("BAIDU_KEY"),
		deeplKey: os.Getenv("DEEPL_KEY"),
		http:     &http.Client{Timeout: 12 * time.Second},
		cache:    make(map[string]string),
	}
}

// Enabled 是否启用翻译（off/none 禁用；google 默认启用；baidu/deepl 需配置 Key）。
func (t *Translator) Enabled() bool {
	switch t.provider {
	case "off", "none":
		return false
	case "baidu":
		return t.baiduApp != "" && t.baiduKey != ""
	case "deepl":
		return t.deeplKey != ""
	default:
		return true
	}
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

// TranslateBatch 批量翻译（串行带缓存，避免并发打爆免费接口）。
func (t *Translator) TranslateBatch(texts []string, toLang string) []string {
	out := make([]string, len(texts))
	for i, s := range texts {
		out[i] = t.Translate(s, toLang)
	}
	return out
}

func (t *Translator) translate(text, toLang string) (string, error) {
	switch t.provider {
	case "baidu":
		return t.baidu(text, toLang)
	case "deepl":
		return t.deepl(text, toLang)
	default:
		return t.google(text, toLang)
	}
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
func (t *Translator) baidu(text, toLang string) (string, error) {
	salt := fmt.Sprintf("%d", time.Now().UnixNano())
	sign := md5.Sum([]byte(t.baiduApp + text + salt + t.baiduKey))
	form := url.Values{}
	form.Set("q", text)
	form.Set("from", "auto")
	form.Set("to", toLang)
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
