package crawler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // 注册 JPEG 解码器（供头像颜色分析）
	_ "image/png"  // 注册 PNG 解码器
	"io"
	"net/http"
	"net/url"
	"strings"
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

// Client GitHub API 客户端（认证可显著提升速率限制）。
type Client struct {
	token string
	http  *http.Client
	// officialOrgs 官方组织表（owner → 展示信息）；默认内置，可用 SetOfficialOrgs 动态覆盖。
	officialOrgs map[string]OrgInfo
}

// NewClient 创建客户端；token 为空时以匿名方式请求（速率受限）。
func NewClient(token string) *Client {
	return &Client{
		token:        token,
		http:         &http.Client{Timeout: 20 * time.Second},
		officialOrgs: defaultOfficialOrgs,
	}
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

func (c *Client) get(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API %s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
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
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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

// SearchRepos 按关键词搜索仓库（按 star 排序）。
func (c *Client) SearchRepos(query string, perPage, page int) ([]Repo, error) {
	var result struct {
		Items []Repo `json:"items"`
	}
	path := fmt.Sprintf("/search/repositories?q=%s&sort=stars&order=desc&per_page=%d&page=%d",
		url.QueryEscape(query), perPage, page)
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
