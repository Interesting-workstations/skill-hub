package skill

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/Interesting-workstations/skill-hub/server/internal/orglogo"
	"github.com/Interesting-workstations/skill-hub/server/internal/response"
)

// Handler 负责 HTTP 层：参数解析、调用 Service、返回统一响应。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册技能资源库相关路由。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/health", h.health)
	mux.HandleFunc("GET /api/v1/stats", h.stats)
	mux.HandleFunc("GET /api/v1/skills", h.listSkills)
	mux.HandleFunc("GET /api/v1/skills/{id}", h.getSkill)
	mux.HandleFunc("GET /api/v1/skills/{id}/download", h.downloadSkill)
	mux.HandleFunc("POST /api/v1/skills/submit", h.submitSkill)
	mux.HandleFunc("GET /api/v1/authors", h.listAuthors)
	mux.HandleFunc("GET /api/v1/authors/{slug}", h.getAuthor)
	mux.HandleFunc("GET /api/v1/categories", h.listCategories)
	mux.HandleFunc("GET /api/v1/categories/{slug}", h.getCategory)
	// 官方组织概览（官方技能 / 官方组织统一数据源）
	mux.HandleFunc("GET /api/v1/official-orgs", h.listOfficialOrgs)
	// 官方组织 logo 图片代理（绕开浏览器访问 GitHub 的防盗链）
	mux.HandleFunc("GET /api/v1/org-logo/{owner}", h.orgLogo)
	// 官方组织 logo 通用代理（白名单域名，支持官网 SVG 等非 GitHub 来源，绕 Cloudflare）
	mux.HandleFunc("GET /api/v1/img-proxy", h.imgProxy)
	// 公开内容：文章 / 站点配置 / SEO（admin 管理，官网读取）
	mux.HandleFunc("GET /api/v1/articles", h.listArticles)
	mux.HandleFunc("GET /api/v1/articles/{id}", h.getArticle)
	mux.HandleFunc("GET /api/v1/site-config", h.getSiteConfig)
	mux.HandleFunc("GET /api/v1/seo", h.getSeo)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"status": "ok"})
}

// GET /api/v1/stats
func (h *Handler) stats(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.Stats())
}

// GET /api/v1/skills?category=&author=&official=&featured=&q=&limit=&offset=
func (h *Handler) listSkills(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	official, err1 := strconv.ParseBool(q.Get("official"))
	featured, err2 := strconv.ParseBool(q.Get("featured"))
	filter := SkillFilter{
		Category: q.Get("category"),
		Author:   q.Get("author"),
		Query:    q.Get("q"),
	}
	if err1 == nil {
		filter.Official = official
	}
	if err2 == nil {
		filter.Featured = featured
	}
	// 分页参数（limit<=0 返回全部）
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Offset = n
		}
	}
	response.OK(w, applyLang(h.svc.ListSkills(filter), q.Get("lang")))
}

// GET /api/v1/skills/{id}
func (h *Handler) getSkill(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	skill, ok := h.svc.GetSkill(id)
	if !ok {
		response.NotFound(w, "技能不存在")
		return
	}
	applyLangOne(&skill, r.URL.Query().Get("lang"))
	response.OK(w, skill)
}

// applyLang 按请求语言覆盖技能标题/描述：lang=zh 且有中文翻译时用中文，否则保留英文。
func applyLang(skills []domain.Skill, lang string) []domain.Skill {
	if lang != "zh" {
		return skills
	}
	for i := range skills {
		applyLangOne(&skills[i], lang)
	}
	return skills
}

func applyLangOne(s *domain.Skill, lang string) {
	if lang != "zh" {
		return
	}
	if s.NameZh != "" {
		s.Name = s.NameZh
	}
	if s.DescriptionZh != "" {
		s.Description = s.DescriptionZh
	}
}

// GET /api/v1/skills/{id}/download —— 下载该技能的 ZIP。
// 优先按技能目录（skillPath）从 GitHub 拉取实际文件打包（与 mcpservers.org 的
// 「下载 ZIP」一致：下载的是技能子文件夹而非整个仓库，文件保持原始内容）；
// 拉取失败（无 skillPath / 网络不可用 / 仓库不存在）时回退为
// 由数据库内容动态生成的 SKILL.md + README.md + meta.json。
func (h *Handler) downloadSkill(w http.ResponseWriter, r *http.Request) {
	skill, ok := h.svc.GetSkill(r.PathValue("id"))
	if !ok {
		response.NotFound(w, "技能不存在")
		return
	}

	// 优先：从 GitHub 拉取技能目录的真实文件并打包
	if files, ok := h.fetchSkillFolder(skill); ok && len(files) > 0 {
		var buf bytes.Buffer
		zw := zip.NewWriter(&buf)
		rels := make([]string, 0, len(files))
		for rel := range files {
			rels = append(rels, rel)
		}
		sort.Strings(rels)
		for _, rel := range rels {
			if f, err := zw.Create(rel); err == nil {
				_, _ = f.Write(files[rel])
			}
		}
		if err := zw.Close(); err == nil {
			h.writeZip(w, buf.Bytes(), skill.ID)
			return
		}
	}

	// 回退：SKILL.md（数据库内容转 Markdown）+ README.md + meta.json
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// SKILL.md：技能主体内容（content 区块转 Markdown）
	if f, err := zw.Create("SKILL.md"); err == nil {
		_, _ = f.Write([]byte(skillMarkdown(skill)))
	}
	// README.md：技能说明与使用方式
	if f, err := zw.Create("README.md"); err == nil {
		_, _ = f.Write([]byte(skillReadme(skill)))
	}
	// meta.json：结构化元信息
	if f, err := zw.Create("meta.json"); err == nil {
		meta, _ := json.MarshalIndent(map[string]any{
			"name":           skill.Name,
			"author":         skill.Author,
			"description":    skill.Description,
			"category":       skill.Category,
			"tags":           skill.Tags,
			"isOfficial":     skill.IsOfficial,
			"githubUrl":      skill.GithubURL,
			"license":        skill.License,
			"installCommand": skill.InstallCommand,
			"skillPath":      skill.SkillPath,
		}, "", "  ")
		_, _ = f.Write(meta)
	}
	if err := zw.Close(); err != nil {
		response.Internal(w)
		return
	}

	h.writeZip(w, buf.Bytes(), skill.ID)
}

// writeZip 以附件形式输出 ZIP 响应。
func (h *Handler) writeZip(w http.ResponseWriter, data []byte, id string) {
	filename := id
	if filename == "" {
		filename = "skill"
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	_, _ = w.Write(data)
}

// fetchSkillFolder 从 GitHub 拉取技能目录（SkillPath）下的实际文件，返回 相对路径 → 内容。
// 需要技能有 githubUrl 与 skillPath；仓库/网络/接口出错时返回 ok=false 由调用方回退。
func (h *Handler) fetchSkillFolder(s domain.Skill) (map[string][]byte, bool) {
	if s.GithubURL == "" || s.SkillPath == "" {
		return nil, false
	}
	u, err := url.Parse(s.GithubURL)
	if err != nil {
		return nil, false
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, false
	}
	fullName := parts[0] + "/" + parts[1]

	client := crawler.NewClientFromEnv()
	repo, err := client.GetRepo(fullName)
	if err != nil {
		return nil, false
	}
	files, err := client.DownloadSkillFolder(fullName, repo.DefaultBranch, s.SkillPath)
	if err != nil || len(files) == 0 {
		return nil, false
	}
	return files, true
}

// skillMarkdown 将技能内容区块转为 SKILL.md 主体。
func skillMarkdown(s domain.Skill) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", s.Name))
	if s.Description != "" {
		b.WriteString(s.Description + "\n\n")
	}
	for _, sec := range s.Content {
		if sec.Heading != "" {
			b.WriteString("## " + sec.Heading + "\n\n")
		}
		for _, line := range sec.Body {
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// skillReadme 生成技能的 README.md。
func skillReadme(s domain.Skill) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s\n\n", s.Name))
	b.WriteString(fmt.Sprintf("> 作者：%s\n\n", s.Author))
	b.WriteString(s.Description + "\n\n")
	if s.InstallCommand != "" {
		b.WriteString("## 安装\n\n```bash\n" + s.InstallCommand + "\n```\n\n")
	}
	if s.GithubURL != "" {
		b.WriteString("## 仓库\n\n" + s.GithubURL + "\n\n")
	}
	if len(s.Tags) > 0 {
		b.WriteString("## 标签\n\n")
		for _, t := range s.Tags {
			b.WriteString("- " + t + "\n")
		}
		b.WriteString("\n")
	}
	if s.License != "" {
		b.WriteString("## 许可\n\n" + s.License + "\n")
	}
	return b.String()
}

// GET /api/v1/authors
func (h *Handler) listAuthors(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.ListAuthors())
}

// GET /api/v1/authors/{slug}
func (h *Handler) getAuthor(w http.ResponseWriter, r *http.Request) {
	detail, ok := h.svc.GetAuthor(r.PathValue("slug"))
	if !ok {
		response.NotFound(w, "作者不存在")
		return
	}
	detail.Skills = applyLang(detail.Skills, r.URL.Query().Get("lang"))
	response.OK(w, detail)
}

// GET /api/v1/categories
func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats := h.svc.ListCategories()
	lang := r.URL.Query().Get("lang")
	for i := range cats {
		cats[i].Skills = applyLang(cats[i].Skills, lang)
	}
	response.OK(w, cats)
}

// GET /api/v1/categories/{slug}
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	cat, ok := h.svc.GetCategory(r.PathValue("slug"))
	if !ok {
		response.NotFound(w, "分类不存在")
		return
	}
	cat.Skills = applyLang(cat.Skills, r.URL.Query().Get("lang"))
	response.OK(w, cat)
}

// GET /api/v1/official-orgs —— 官方组织概览（含各组织官方技能数）。
func (h *Handler) listOfficialOrgs(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.ListOfficialOrgs())
}

// GET /api/v1/org-logo/{owner} —— 官方组织 logo 图片服务。
// 图片已提前下载到服务器本地（data/org-logos），优先读本地文件（不再实时回源 GitHub，
// 解决 GitHub 头像/防盗链不稳定问题）；首次未缓存时自动下载并缓存兜底。
func (h *Handler) orgLogo(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	if owner == "" {
		response.NotFound(w, "组织不存在")
		return
	}
	logoURL, _ := h.svc.OrgLogoURL(owner)
	orglogo.Serve(w, r, owner, logoURL)
}

// logoProxyAllowedHosts 官方组织 logo 代理白名单（防 SSRF，仅代理知名品牌来源）。
var logoProxyAllowedHosts = []string{
	"sst.dev",
	"perplexity.ai",
	"zhipuai.cn",
	"www.zhipuai.cn",
	"bigmodel.cn",
	"nomic.ai",
	"www.nomic.ai",
	"avatars.githubusercontent.com",
	"github.com",
	"upload.wikimedia.org",
}

// logoProxyAllowed 判断 host 是否在白名单（含子域名）。
func logoProxyAllowed(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	for _, a := range logoProxyAllowedHosts {
		if h == a || strings.HasSuffix(h, "."+a) {
			return true
		}
	}
	return false
}

// GET /api/v1/img-proxy?url=... —— 通用图片代理（白名单域名）。
// 部分官网（如 Perplexity 的 Cloudflare 返回 NotSameOrigin 头）会拦截浏览器直接加载图片，
// 改由后端无 Referer 拉取并转发，同时提供缓存。
func (h *Handler) imgProxy(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	if raw == "" {
		response.InvalidParam(w, "缺少 url 参数")
		return
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		response.InvalidParam(w, "url 不合法")
		return
	}
	if !logoProxyAllowed(u.Host) {
		response.NotFound(w, "logo 不存在")
		return
	}
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		response.NotFound(w, "logo 获取失败")
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	resp, err := client.Do(req)
	if err != nil {
		response.NotFound(w, "logo 获取失败")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		response.NotFound(w, "logo 不存在")
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := io.Copy(w, resp.Body); err != nil {
		return
	}
}

// POST /api/v1/skills/submit —— 用户提交技能（进入待审核）。
func (h *Handler) submitSkill(w http.ResponseWriter, r *http.Request) {
	var in SubmitSkillInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		response.InvalidParam(w, "请求参数错误")
		return
	}
	created, err := h.svc.SubmitSkill(in)
	if err != nil {
		response.InvalidParam(w, err.Error())
		return
	}
	response.OK(w, created)
}

// GET /api/v1/articles —— 已发布文章列表。
func (h *Handler) listArticles(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.ListArticles())
}

// GET /api/v1/articles/{id} —— 文章详情（浏览量 +1，同一 IP 当天去重；?incr=0 不计数）。
func (h *Handler) getArticle(w http.ResponseWriter, r *http.Request) {
	// incr=0：前端已读过的文章跳过计数（localStorage 去重）
	a, ok := h.svc.GetArticle(r.PathValue("id"), clientIP(r), r.URL.Query().Get("incr") != "0")
	if !ok {
		response.NotFound(w, "文章不存在")
		return
	}
	response.OK(w, a)
}

// clientIP 从请求中提取客户端 IP（优先代理头，回退 RemoteAddr）。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// GET /api/v1/site-config —— 站点配置（站点名 / 标语 / ICP 等）。
func (h *Handler) getSiteConfig(w http.ResponseWriter, _ *http.Request) {
	if cfg, ok := h.svc.GetSiteConfig(); ok {
		response.OK(w, cfg)
		return
	}
	response.OK(w, map[string]string{"siteName": "Agent Skills 资源库", "slogan": "AI 编程助手的可复用技能"})
}

// GET /api/v1/seo —— SEO 配置（默认标题 / 描述 / 关键词）。
func (h *Handler) getSeo(w http.ResponseWriter, _ *http.Request) {
	if cfg, ok := h.svc.GetSeo(); ok {
		response.OK(w, cfg)
		return
	}
	response.OK(w, map[string]string{})
}
