// Package orglogo 官方组织 logo 本地图片缓存。
//
// GitHub / 官网头像不稳定（防盗链、被墙、CDN 抖动），此包把组织 logo
// 提前下载到本地磁盘目录（data/org-logos/{owner}.{ext}），
// 前端统一访问本地缓存，不再每次实时回源 GitHub。
//
// 能力：
//   - Fetch：从来源（GitHub 头像或官网白名单）下载并保存到本地
//   - Serve：优先读本地缓存，无缓存时自动下载并缓存（首次访问兜底）
//   - 后台「重新拉取图片」可强制刷新所有组织图片
package orglogo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Dir 本地图片缓存目录（相对程序运行目录）。
// Docker 中 server 工作目录为 /app，即 /app/data/org-logos（compose 挂载卷持久化）。
var Dir = "data/org-logos"

// maxSize 单张图片最大体积（2MB），防止异常来源撑爆磁盘。
const maxSize = 2 << 20

// client 下载用 HTTP 客户端（带浏览器 UA，绕部分站点 UA 校验）。
var client = &http.Client{Timeout: 15 * time.Second}

// Source 解析组织 logo 的下载来源 URL。
//   - logoURL 为空或指向 github.com → 回退 GitHub 头像 https://github.com/{owner}.png
//   - 否则使用 logoURL 原样（官网等来源，后台显式锁定）
func Source(owner, logoURL string) string {
	logoURL = strings.TrimSpace(logoURL)
	if logoURL == "" || strings.HasPrefix(logoURL, "https://github.com/") {
		name := strings.Trim(owner, "/")
		if logoURL != "" {
			// https://github.com/{name}.png 或 https://github.com/{name}
			name = strings.TrimSuffix(strings.TrimPrefix(logoURL, "https://github.com/"), ".png")
			name = strings.Trim(name, "/")
		}
		return "https://github.com/" + url.PathEscape(name) + ".png"
	}
	return logoURL
}

// LocalPath 返回 owner 对应的本地缓存文件路径。
// 文件不存在时返回 os.ErrNotExist。
func LocalPath(owner string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(Dir, sanitize(owner)+".*"))
	if err != nil || len(matches) == 0 {
		return "", os.ErrNotExist
	}
	return matches[0], nil
}

// Has 是否已有本地缓存。
func Has(owner string) bool {
	_, err := LocalPath(owner)
	return err == nil
}

// Fetch 从来源下载组织 logo 并保存到本地缓存，返回本地文件路径。
// 会清理该组织旧的扩展名文件（如 .png → .svg）。
func Fetch(owner, logoURL string) (string, error) {
	src := Source(owner, logoURL)
	req, err := http.NewRequest(http.MethodGet, src, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败 status=%d (%s)", resp.StatusCode, src)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxSize {
		return "", fmt.Errorf("图片过大 (>%dMB): %s", maxSize>>20, src)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("图片内容为空: %s", src)
	}

	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return "", err
	}
	name := sanitize(owner) + extFromContentType(resp.Header.Get("Content-Type"))
	path := filepath.Join(Dir, name)

	// 清理同组织的旧扩展名文件（避免残留 .png 后又生成 .svg）
	if olds, _ := filepath.Glob(filepath.Join(Dir, sanitize(owner)+".*")); len(olds) > 0 {
		for _, f := range olds {
			if f != path {
				_ = os.Remove(f)
			}
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Serve 写响应：优先读本地缓存；无缓存时自动下载并缓存后返回。
// 下载失败时返回 404（不阻断页面渲染，前端可回退占位图）。
func Serve(w http.ResponseWriter, r *http.Request, owner, logoURL string) {
	if p, err := LocalPath(owner); err == nil {
		serveFile(w, r, p)
		return
	}
	if p, err := Fetch(owner, logoURL); err == nil {
		serveFile(w, r, p)
		return
	}
	http.NotFound(w, r)
}

// serveFile 输出本地图片文件，带长缓存。
func serveFile(w http.ResponseWriter, r *http.Request, path string) {
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, path)
}

// sanitize 清理 owner 中的不安全字符，防止路径穿越。
func sanitize(owner string) string {
	owner = strings.TrimSpace(owner)
	owner = strings.ReplaceAll(owner, "/", "_")
	owner = strings.ReplaceAll(owner, "\\", "_")
	owner = strings.ReplaceAll(owner, "..", "_")
	return owner
}

// extFromContentType 根据响应 Content-Type 推断扩展名（默认 .png）。
func extFromContentType(ct string) string {
	switch strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0])) {
	case "image/jpeg":
		return ".jpg"
	case "image/svg+xml":
		return ".svg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	case "image/avif":
		return ".avif"
	default: // image/png 及未知类型
		return ".png"
	}
}

// ErrNotExist 与 os.ErrNotExist 一致（供外部判断）。
var ErrNotExist = errors.New("本地缓存不存在")
