package crawler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // 注册 JPEG 解码器（供头像颜色分析）
	_ "image/png"  // 注册 PNG 解码器
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// apiBase GitHub API 基地址（var 以便测试注入 mock server）。
var apiBase = "https://api.github.com"

// OrgInfo 官方组织的展示信息。
type OrgInfo struct {
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

// OfficialOrg 官方组织（owner → 展示信息），用于从数据库动态注入爬虫。
type OfficialOrg struct {
	Owner       string `json:"owner"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

// tokenEntry token 池中单个 token 及其健康状态。
type tokenEntry struct {
	token    string
	broken   bool      // 熔断标记：请求被拒绝/限流（401/403/429）
	brokenAt time.Time // 熔断时间：冷却期后自动恢复重试
}

// tokenCooldown 熔断冷却期：被标记坏的 token 冷却到期后自动重新尝试。
const tokenCooldown = 5 * time.Minute

// TokenPool GitHub Token 池：管理多个 token，支持故障切换与健康检查。
// 线程安全：可被多个 Client/goroutine 共享。
type TokenPool struct {
	mu      sync.Mutex
	entries []*tokenEntry
	next    int // 轮询游标（Round-Robin）
}

// NewTokenPool 创建 token 池；tokens 为要管理的 GitHub Token 列表。
func NewTokenPool(tokens []string) *TokenPool {
	p := &TokenPool{}
	for _, t := range tokens {
		if strings.TrimSpace(t) != "" {
			p.entries = append(p.entries, &tokenEntry{token: strings.TrimSpace(t)})
		}
	}
	return p
}

// Tokens 返回池中全部 token 字符串（含熔断中的）。
func (p *TokenPool) Tokens() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, 0, len(p.entries))
	for _, e := range p.entries {
		out = append(out, e.token)
	}
	return out
}

// Pick 轮询选取一个当前可用（未熔断或已过冷却期）的 token；无可用时返回空串。
func (p *TokenPool) Pick() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.entries) == 0 {
		return ""
	}
	for i := 0; i < len(p.entries); i++ {
		idx := (p.next + i) % len(p.entries)
		e := p.entries[idx]
		if !e.broken || time.Since(e.brokenAt) > tokenCooldown {
			// 冷却到期自动恢复
			if e.broken {
				e.broken = false
				e.brokenAt = time.Time{}
			}
			p.next = (idx + 1) % len(p.entries)
			return e.token
		}
	}
	return ""
}

// HasAvailable 池中是否至少有一个可用 token。
func (p *TokenPool) HasAvailable() bool {
	return p.Pick() != ""
}

// MarkBroken 标记指定 token 熔断（被 GitHub 拒绝/限流），冷却期后自动恢复。
func (p *TokenPool) MarkBroken(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, e := range p.entries {
		if e.token == token {
			e.broken = true
			e.brokenAt = time.Now()
			return
		}
	}
}

// MarkOK 标记指定 token 恢复可用（请求成功 / 健康检查通过）。
func (p *TokenPool) MarkOK(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, e := range p.entries {
		if e.token == token {
			e.broken = false
			e.brokenAt = time.Time{}
			return
		}
	}
}

// CheckHealth 健康检查：对池中每个 token 调用 GitHub rate_limit 接口，
// 判断其是否有效可用；同时把失效/被限流的 token 标记为熔断。
// 返回 map[token]是否可用（用于后台展示检测结果）。
func (p *TokenPool) CheckHealth(httpClient *http.Client) map[string]bool {
	result := make(map[string]bool)
	for _, t := range p.Tokens() {
		ok := checkTokenHealth(httpClient, t)
		result[t] = ok
		if ok {
			p.MarkOK(t)
		} else {
			p.MarkBroken(t)
		}
	}
	return result
}

// TokenHealth 单个 GitHub Token 的健康检查结果（供后台展示，脱敏）。
type TokenHealth struct {
	Masked string `json:"masked"` // 脱敏显示（前 8 后 4）
	OK     bool   `json:"ok"`     // 是否可用
	Detail string `json:"detail"` // 说明（有效 / 具体错误）
}

// CheckTokenHealth 检测单个 token 是否可用（GET /rate_limit），返回脱敏结果。
func CheckTokenHealth(httpClient *http.Client, token string) TokenHealth {
	req, err := http.NewRequest(http.MethodGet, apiBase+"/rate_limit", nil)
	if err != nil {
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: "请求构造失败"}
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: "网络错误: " + err.Error()}
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return TokenHealth{Masked: maskToken(token), OK: true, Detail: "有效"}
	case http.StatusUnauthorized:
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: "无效（401 未授权）"}
	case http.StatusForbidden:
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: "被拒绝（403 无权限）"}
	case http.StatusTooManyRequests:
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: "已限流（429 超配额）"}
	default:
		return TokenHealth{Masked: maskToken(token), OK: false, Detail: fmt.Sprintf("异常状态 %d", resp.StatusCode)}
	}
}

// maskToken 脱敏 token：保留前 8 位与后 4 位，中间用 * 省略。
func maskToken(t string) string {
	if len(t) <= 12 {
		return "****"
	}
	return t[:8] + "****" + t[len(t)-4:]
}

// checkTokenHealth 用 GET /rate_limit 探测单个 token 是否有效。
func checkTokenHealth(httpClient *http.Client, token string) bool {
	return CheckTokenHealth(httpClient, token).OK
}

// TokensFromEnv 从环境变量读取 GitHub Token 列表：
// 优先 GITHUB_TOKENS（逗号分隔多个），兼容旧 GITHUB_TOKEN（单个）。
func TokensFromEnv() []string {
	raw := os.Getenv("GITHUB_TOKENS")
	if strings.TrimSpace(raw) == "" {
		if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
			return []string{t}
		}
		return nil
	}
	var out []string
	for _, t := range strings.Split(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// Client GitHub API 客户端（认证可显著提升速率限制）。
type Client struct {
	token string     // 单个 token（兼容旧逻辑；使用 pool 时忽略）
	pool  *TokenPool // token 池（多 token 故障切换；nil 表示单 token/匿名模式）
	http  *http.Client
	// officialOrgs 官方组织表（owner → 展示信息）；默认内置，可用 SetOfficialOrgs 动态覆盖。
	officialOrgs map[string]OrgInfo
	// stopCh 停止信号：调用 Cancel() 关闭后，进行中的请求与循环尽快退出（后台手动停止任务用）。
	stopCh     chan struct{}
	cancelOnce sync.Once
	// tokenBroken 熔断标记：单 token 模式被 GitHub 限流/拒绝（401/403/429）后置位，
	// 后续请求直接以匿名方式发起，避免每个请求都先被 token 拒绝。
	tokenBroken bool
}

// NewClient 创建客户端；token 为空时以匿名方式请求（速率受限）。
// 若要使用多 token 池，请用 NewClientWithTokens。
func NewClient(token string) *Client {
	return &Client{
		token:        token,
		http:         &http.Client{Timeout: 20 * time.Second},
		officialOrgs: defaultOfficialOrgs,
		stopCh:       make(chan struct{}),
	}
}

// NewClientWithTokens 创建使用 token 池的客户端：多个 token 自动故障切换。
func NewClientWithTokens(tokens []string) *Client {
	c := NewClient("")
	if len(tokens) > 0 {
		c.pool = NewTokenPool(tokens)
	}
	return c
}

// NewClientFromEnv 从环境变量（GITHUB_TOKENS / GITHUB_TOKEN）创建带 token 池的客户端。
func NewClientFromEnv() *Client {
	return NewClientWithTokens(TokensFromEnv())
}

// SetTokenPool 动态设置/替换 token 池（后台配置变更后调用，立即生效）。
func (c *Client) SetTokenPool(pool *TokenPool) {
	c.pool = pool
}

// HasToken 是否配置且仍可用的 GitHub Token。匿名模式（无 Token / Token 已熔断）
// GitHub 限流仅 60 次/小时，调用方应据此跳过高耗配额的自动发现并限制单次抓取量。
func (c *Client) HasToken() bool {
	if c.pool != nil {
		return c.pool.HasAvailable()
	}
	return c.token != "" && !c.tokenBroken
}

// currentToken 返回当前可用的 token（池优先，其次单 token）；无可用返回空串。
// 供不经过 get() 的辅助请求（如头像下载）使用。
func (c *Client) currentToken() string {
	if c.pool != nil {
		return c.pool.Pick()
	}
	if !c.tokenBroken {
		return c.token
	}
	return ""
}

// Cancel 请求停止客户端：关闭停止信号，进行中的 HTTP 请求立即取消。
// 停止后再次调用无副作用（幂等）。
func (c *Client) Cancel() {
	c.cancelOnce.Do(func() { close(c.stopCh) })
}

// IsCancelled 客户端是否已被停止。
func (c *Client) IsCancelled() bool {
	select {
	case <-c.stopCh:
		return true
	default:
		return false
	}
}

// get 请求 GitHub API。
// 池模式：轮询选取一个可用 token；若被限流/拒绝（401/403/429），自动标记该 token 熔断并切换下一个；
// 所有 token 都失败后降级为匿名请求重试一次。
// 单 token 模式：token 被限流/拒绝后熔断，降级匿名重试。
func (c *Client) get(path string, out any) error {
	if c.pool != nil {
		return c.getWithPool(path, out)
	}
	withToken := c.token != "" && !c.tokenBroken
	status, err := c.doGet(path, c.token, out)
	if err != nil && withToken &&
		(status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusTooManyRequests) {
		c.tokenBroken = true
		_, err = c.doGet(path, "", out)
	}
	return err
}

// getWithPool 池模式 GET：轮询尝试可用 token，失败自动切换，全部失败降级匿名。
func (c *Client) getWithPool(path string, out any) error {
	var lastErr error
	triedAny := false
	for {
		tok := c.pool.Pick()
		if tok == "" {
			break // 无可用 token
		}
		triedAny = true
		status, err := c.doGet(path, tok, out)
		if err == nil {
			c.pool.MarkOK(tok)
			return nil
		}
		lastErr = err
		if status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusTooManyRequests {
			c.pool.MarkBroken(tok)
			continue // 换下一个 token
		}
		return err // 非限流类错误（如 404 仓库不存在）直接返回
	}
	// 池中 token 全部失败 → 匿名重试一次（避免因 token 全失效导致整批失败）
	if triedAny {
		if _, err := c.doGet(path, "", out); err == nil {
			return nil
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("GitHub API %s: 无可用 token", path)
}

// doGet 执行一次 GET 请求（token 为空表示匿名）。返回 HTTP 状态码。
func (c *Client) doGet(path, token string, out any) (int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 停止信号触发时取消请求上下文（手动停止任务时立即中断正在进行的请求）
	go func() {
		select {
		case <-c.stopCh:
			cancel()
		case <-ctx.Done():
		}
	}()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+path, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("GitHub API %s: %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

// SetOfficialOrgs 注入动态官方组织列表（来自数据库，替换内置默认）。
// 传入空列表时不生效（保留当前列表），避免误清空。
func (c *Client) SetOfficialOrgs(orgs []OfficialOrg) {
	if len(orgs) == 0 {
		return
	}
	m := make(map[string]OrgInfo, len(orgs))
	for _, o := range orgs {
		owner := strings.ToLower(strings.TrimSpace(o.Owner))
		if owner != "" {
			m[owner] = OrgInfo{DisplayName: o.DisplayName, Avatar: o.Avatar}
		}
	}
	if len(m) > 0 {
		c.officialOrgs = m
	}
}

// IsOfficial 判断仓库 owner 是否为官方组织。
func (c *Client) IsOfficial(owner string) bool {
	_, ok := c.officialOrgs[strings.ToLower(owner)]
	return ok
}

// OfficialDisplayName 返回官方组织的展示名；非官方组织返回 GitHub 用户名本身。
func (c *Client) OfficialDisplayName(owner string) string {
	if info, ok := c.officialOrgs[strings.ToLower(owner)]; ok && info.DisplayName != "" {
		return info.DisplayName
	}
	return owner
}

// OfficialAvatar 返回官方组织的头像 emoji；非官方组织返回空字符串。
// 兼容 GitHub owner 与展示名（如 "Hugging Face"）两种输入。
func (c *Client) OfficialAvatar(owner string) string {
	lower := strings.ToLower(owner)
	if info, ok := c.officialOrgs[lower]; ok {
		return info.Avatar
	}
	for _, info := range c.officialOrgs {
		if info.DisplayName != "" && strings.EqualFold(info.DisplayName, owner) {
			return info.Avatar
		}
	}
	return ""
}

// GetUserType 查询 GitHub 用户/组织类型（Organization / User / NotFound）。
// 用于官方组织校验：owner 必须是真正的 GitHub 组织（type=Organization），
// 否则爬虫会把个人账号的仓库误标为官方，logo 也会显示个人头像。
func (c *Client) GetUserType(owner string) (string, error) {
	var u struct {
		Type string `json:"type"`
	}
	if err := c.get("/users/"+url.PathEscape(owner), &u); err != nil {
		if strings.Contains(err.Error(), "404") {
			return "NotFound", nil
		}
		return "", err
	}
	if u.Type == "" {
		return "NotFound", nil
	}
	return u.Type, nil
}

// OrgCheck GitHub 组织校验信息（类型 + 头像是否有效）。
type OrgCheck struct {
	Type     string // Organization / User / NotFound
	Avatar   string // GitHub 头像 URL
	AvatarOK bool   // 头像是否为有效品牌 logo（非默认 identicon / 纯色块）
}

// CheckOrg 校验 owner 是否为真正的 GitHub 组织，并检测头像是否有效。
// 部分组织虽 type=Organization 但从未设置头像（GitHub 默认 identicon 或纯色块），
// 显示出来是乱码几何图，需要提示管理员改用官网 logo。
func (c *Client) CheckOrg(owner string) (OrgCheck, error) {
	var u struct {
		Type   string `json:"type"`
		Avatar string `json:"avatar_url"`
	}
	if err := c.get("/users/"+url.PathEscape(owner), &u); err != nil {
		if strings.Contains(err.Error(), "404") {
			return OrgCheck{Type: "NotFound"}, nil
		}
		return OrgCheck{}, err
	}
	if u.Type == "" {
		return OrgCheck{Type: "NotFound"}, nil
	}
	if u.Type != "Organization" {
		return OrgCheck{Type: u.Type, Avatar: u.Avatar}, nil
	}
	return OrgCheck{Type: u.Type, Avatar: u.Avatar, AvatarOK: c.avatarLooksReal(u.Avatar)}, nil
}

// avatarLooksReal 下载头像（64px）并判断是否有效品牌 logo。
// 启发式：GitHub 默认 identicon / 纯色块的颜色极少，有效 logo 通常色彩丰富。
// 只要主要色占比 < 92% 即视为有效（避免误杀单色但清晰的 logo）。
func (c *Client) avatarLooksReal(avatarURL string) bool {
	if avatarURL == "" {
		return false
	}
	u, err := url.Parse(avatarURL)
	if err != nil {
		return false
	}
	q := u.Query()
	q.Set("s", "64")
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if tok := c.currentToken(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil || len(data) == 0 {
		return false
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}
	b := img.Bounds()
	sx := max(1, b.Dx()/24)
	sy := max(1, b.Dy()/24)
	total := 0
	counts := make(map[uint32]int)
	for y := b.Min.Y; y < b.Max.Y; y += sy {
		for x := b.Min.X; x < b.Max.X; x += sx {
			r, g, bl, _ := img.At(x, y).RGBA()
			counts[(r>>12)<<8|(g>>12)<<4|(bl>>12)]++
			total++
		}
	}
	if total == 0 {
		return false
	}
	best := 0
	second := 0
	for _, n := range counts {
		if n > best {
			second = best
			best = n
		} else if n > second {
			second = n
		}
	}
	// 主色 ≥ 95% 且第二主色 < 1%：几乎纯色（GitHub 默认 identicon / 未设置头像）。
	// 黑底品牌 logo（如 Vercel、Anthropic）通常有明显亮色内容（第二色 ≥ 2%），不会误报。
	return best*100/total < 95 || second*100/total >= 1
}

// SearchRepos 按关键词搜索仓库。
// sort 可为 stars/updated/created（空则默认 stars 降序）；
// 其中 updated 用于“每日最新”场景——按最近更新时间排序抓取最新仓库。
func (c *Client) SearchRepos(query string, sort string, perPage, page int) ([]Repo, error) {
	var result struct {
		Items []Repo `json:"items"`
	}
	if sort == "" {
		sort = "stars"
	}
	path := fmt.Sprintf("/search/repositories?q=%s&sort=%s&order=desc&per_page=%d&page=%d",
		url.QueryEscape(query), sort, perPage, page)
	if err := c.get(path, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

// GetRepo 获取单个仓库的详细信息（含默认分支、star、license）。
func (c *Client) GetRepo(fullName string) (Repo, error) {
	var repo Repo
	if err := c.get("/repos/"+fullName, &repo); err != nil {
		return Repo{}, err
	}
	return repo, nil
}

// ListOrgRepos 获取 GitHub 组织的公开仓库（按 star 降序，取前 perPage 个）。
// 用于爬虫执行时自动发现官方组织的技能仓库，使官方技能与官方组织挂钩。
func (c *Client) ListOrgRepos(org string, perPage int) ([]Repo, error) {
	if perPage <= 0 {
		perPage = 5
	}
	var out []Repo
	if err := c.get(fmt.Sprintf("/orgs/%s/repos?per_page=%d&sort=stars&type=public", url.PathEscape(org), perPage), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// HasSkillStructure 判断仓库是否具备 skill 结构（根目录有 SKILL.md，或有 skills/skillsets 目录）。
// 用于官方仓库发现：只保留真正能产出技能的结构，避免爬取 agent/mcp 代码仓库浪费资源。
func (c *Client) HasSkillStructure(fullName string) bool {
	entries, err := c.ListContents(fullName, "", "")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Type == "file" && strings.EqualFold(e.Name, "SKILL.md") {
			return true
		}
		if e.Type == "dir" && (strings.EqualFold(e.Name, "skills") || strings.EqualFold(e.Name, "skillsets")) {
			return true
		}
	}
	return false
}

// ListContents 列出仓库某路径下的内容。
func (c *Client) ListContents(fullName, path, ref string) ([]ContentEntry, error) {
	p := "/repos/" + fullName + "/contents/" + path
	if ref != "" {
		p += "?ref=" + url.QueryEscape(ref)
	}
	var entries []ContentEntry
	if err := c.get(p, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// GetFile 获取仓库文件内容（自动 base64 解码）。
func (c *Client) GetFile(fullName, path, ref string) ([]byte, error) {
	p := "/repos/" + fullName + "/contents/" + path
	if ref != "" {
		p += "?ref=" + url.QueryEscape(ref)
	}
	var f struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.get(p, &f); err != nil {
		return nil, err
	}
	if f.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(f.Content)
	}
	return []byte(f.Content), nil
}

// GetReadme 获取仓库某目录下的 README 内容（自动 base64 解码）。
// dir 为空表示仓库根目录；GitHub 自动匹配目录内的 README 文件（大小写、扩展名均支持）。
// 目录下不存在 README 时返回错误。
func (c *Client) GetReadme(fullName, dir, ref string) ([]byte, error) {
	p := "/repos/" + fullName + "/readme"
	if dir != "" {
		p += "/" + dir
	}
	if ref != "" {
		p += "?ref=" + url.QueryEscape(ref)
	}
	var f struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.get(p, &f); err != nil {
		return nil, err
	}
	if f.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(f.Content)
	}
	return []byte(f.Content), nil
}

// TreeEntry 递归文件树（git trees API）中的一个条目。
type TreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"` // blob / tree
	Size int    `json:"size"`
	SHA  string `json:"sha"`
}

// GetTree 获取仓库指定 ref（分支/标签/SHA）的完整递归文件树。
// 用于按技能目录（SkillPath）下载其下所有文件。
func (c *Client) GetTree(fullName, ref string) ([]TreeEntry, error) {
	var result struct {
		Truncated bool        `json:"truncated"`
		Tree      []TreeEntry `json:"tree"`
	}
	p := "/repos/" + fullName + "/git/trees/" + ref + "?recursive=1"
	if err := c.get(p, &result); err != nil {
		return nil, err
	}
	if result.Truncated {
		return nil, fmt.Errorf("GitHub 返回的文件树被截断（仓库过大）")
	}
	return result.Tree, nil
}

// GetBlob 按 blob SHA 获取文件原始内容（自动 base64 解码）。
func (c *Client) GetBlob(fullName, sha string) ([]byte, error) {
	var f struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err := c.get("/repos/"+fullName+"/git/blobs/"+sha, &f); err != nil {
		return nil, err
	}
	if f.Encoding == "base64" {
		return base64.StdEncoding.DecodeString(f.Content)
	}
	return []byte(f.Content), nil
}

// DownloadSkillFolder 下载技能目录（SkillPath 指定）下的所有文件，返回 相对路径 → 内容。
// 仓库根技能（path 为空）会返回整个仓库文件。失败时返回 error。
func (c *Client) DownloadSkillFolder(fullName, ref, skillPath string) (map[string][]byte, error) {
	tree, err := c.GetTree(fullName, ref)
	if err != nil {
		return nil, err
	}
	prefix := strings.TrimPrefix(skillPath, "/")
	files := make(map[string][]byte)
	for _, e := range tree {
		if e.Type != "blob" {
			continue
		}
		rel := e.Path
		if prefix != "" {
			if !strings.HasPrefix(e.Path, prefix+"/") {
				continue
			}
			rel = strings.TrimPrefix(e.Path, prefix+"/")
		}
		content, err := c.GetBlob(fullName, e.SHA)
		if err != nil {
			return nil, err
		}
		files[rel] = content
	}
	return files, nil
}
