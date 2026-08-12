package crawler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://api.github.com"

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
