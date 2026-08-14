package crawler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// apiBase GitHub API 基地址（var 以便测试注入 mock server）。
var apiBase = "https://api.github.com"

// Client GitHub API 客户端（认证可显著提升速率限制）。
type Client struct {
	token string
	http  *http.Client
}

// NewClient 创建客户端；token 为空时以匿名方式请求（速率受限）。
func NewClient(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 20 * time.Second},
	}
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
