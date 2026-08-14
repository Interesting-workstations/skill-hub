package crawler

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// CrawlOptions 爬取选项。
type CrawlOptions struct {
	// Query 搜索关键词（可多次指定不同关键词，内部去重）。
	Query string
	// Repos 额外指定要爬取的仓库（如 "anthropics/skills"），优先于搜索结果。
	Repos []string
	// Limit 最多处理的仓库数量。
	Limit int
	// PerPage 每次搜索返回的仓库数。
	PerPage int
	// OnProgress 可选进度回调：每处理完一个仓库后调用（done 已处理数，total 总仓库数）。
	// 用于后台实时推送执行进度；nil 时忽略。
	OnProgress func(done, total int)
}

// RepoFailure 单个仓库的爬取失败信息。
type RepoFailure struct {
	Repo   string `json:"repo"`
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

// DefaultOptions 返回默认爬取选项。
func DefaultOptions() CrawlOptions {
	return CrawlOptions{Query: "agent skills", Limit: 20, PerPage: 10}
}

// Crawl 执行爬取：先爬取指定仓库，再按关键词搜索补充，最后去重返回技能列表。
func (c *Client) Crawl(opts CrawlOptions) ([]Skill, error) {
	skills, _, err := c.CrawlDetailed(opts)
	return skills, err
}

// CrawlDetailed 与 Crawl 相同，但额外返回每个仓库的失败信息（供后台监控）。
// 每处理完一个仓库会调用 opts.OnProgress（如有）以推送实时进度。
func (c *Client) CrawlDetailed(opts CrawlOptions) ([]Skill, []RepoFailure, error) {
	seen := make(map[string]bool)       // ID 指纹
	seenSource := make(map[string]bool) // name|author|githubUrl 指纹（同源同名视为同一技能）
	var skills []Skill
	var failures []RepoFailure

	// 先执行搜索拿到结果，便于统一计算总仓库数与进度
	searchRepos, err := c.searchReposPaginated(opts.Query, opts.PerPage, opts.Limit)
	if err != nil {
		return nil, failures, fmt.Errorf("搜索仓库失败: %w", err)
	}

	// 待处理仓库：指定仓库（官方）优先，其后是搜索结果
	repos := make([]string, 0, len(opts.Repos)+len(searchRepos))
	repos = append(repos, opts.Repos...)
	for _, r := range searchRepos {
		repos = append(repos, r.FullName)
	}

	total := len(repos)
	for i, fullName := range repos {
		repo, err := c.GetRepo(fullName)
		if err != nil {
			failures = append(failures, RepoFailure{Repo: fullName, Reason: "获取仓库失败", Error: err.Error()})
			if opts.OnProgress != nil {
				opts.OnProgress(i+1, total)
			}
			continue
		}
		got, err := c.crawlRepo(repo)
		if err != nil {
			failures = append(failures, RepoFailure{Repo: fullName, Reason: "解析仓库失败", Error: err.Error()})
			if opts.OnProgress != nil {
				opts.OnProgress(i+1, total)
			}
			continue
		}
		for _, s := range got {
			if seen[s.ID] {
				continue
			}
			// 同源同名（name + author + githubUrl）视为同一技能，跳过重复
			sourceKey := s.Name + "|" + s.Author + "|" + s.GithubURL
			if seenSource[sourceKey] {
				continue
			}
			seen[s.ID] = true
			seenSource[sourceKey] = true
			skills = append(skills, s)
		}
		if opts.OnProgress != nil {
			opts.OnProgress(i+1, total)
		}
	}

	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, failures, nil
}

// crawlRepo 解析单个仓库中的 skill。
// 识别规则：
//  1. 仓库根目录存在 SKILL.md → 整个仓库是一个 skill
//  2. 存在 skills/ 目录 → 其下每个含 SKILL.md 的子目录是一个 skill
func (c *Client) crawlRepo(repo Repo) ([]Skill, error) {
	entries, err := c.ListContents(repo.FullName, "", repo.DefaultBranch)
	if err != nil {
		return nil, err
	}

	var skills []Skill
	for _, e := range entries {
		switch {
		case e.Type == "file" && strings.EqualFold(e.Name, "SKILL.md"):
			s, err := c.buildSkill(repo, "", e.Path)
			if err == nil {
				skills = append(skills, s)
			}
		case e.Type == "dir" && (strings.EqualFold(e.Name, "skills") || strings.EqualFold(e.Name, "skillsets")):
			sub, err := c.ListContents(repo.FullName, e.Path, repo.DefaultBranch)
			if err != nil {
				continue
			}
			for _, subEntry := range sub {
				if subEntry.Type != "dir" {
					continue
				}
				s, err := c.buildSkill(repo, subEntry.Name, subEntry.Path+"/SKILL.md")
				if err == nil {
					skills = append(skills, s)
				}
			}
		}
	}
	return skills, nil
}

// searchReposPaginated 按关键词分页搜索仓库（按 star 降序），直到凑够 limit
// 个仓库或没有更多结果为止。GitHub 搜索 API 单页上限 100、总上限 1000 条。
func (c *Client) searchReposPaginated(query string, perPage, limit int) ([]Repo, error) {
	if perPage <= 0 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	var all []Repo
	for page := 1; page <= 10; page++ {
		batch, err := c.SearchRepos(query, perPage, page)
		if err != nil {
			return all, err
		}
		if len(batch) == 0 {
			break
		}
		all = append(all, batch...)
		if limit > 0 && len(all) >= limit {
			break
		}
	}
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// pickBody 决定技能的展示正文：优先 SKILL.md 自身（技能定义本体）；
// 仅当 SKILL.md 无实质正文时才回退到同目录 README.md。
func pickBody(skillMD, readme []byte) []byte {
	if hasSubstantiveContent(skillMD) {
		return skillMD
	}
	if len(readme) > 0 {
		return readme
	}
	return skillMD
}

// hasSubstantiveContent 判断内容去掉 frontmatter 后是否还有实质正文
// （而非仅 frontmatter 或空行）。
func hasSubstantiveContent(md []byte) bool {
	return strings.TrimSpace(stripFrontmatter(string(md))) != ""
}

// buildSkill 读取 SKILL.md 并组装 Skill 记录。
// 官方组织仓库的技能会标记为官方（IsOfficial），并根据内容推断分类。
// 展示名（Name）做标题化（frontend-design → Frontend Design），作者（Author）用官方展示名
// （anthropics → Anthropic）；同时生成安装命令（npx skills add ... --skill <目录>）、
// 技能目录下载地址与技能路径（SkillPath）。
// 正文（Content）优先取同目录 README.md，目录下没有 README 时退回 SKILL.md 自身正文。
func (c *Client) buildSkill(repo Repo, skillName, mdPath string) (Skill, error) {
	content, err := c.GetFile(repo.FullName, mdPath, repo.DefaultBranch)
	if err != nil {
		return Skill{}, err
	}
	name, desc := ParseSkillMD(content)
	if name == "" {
		name = skillName
	}
	if name == "" {
		name = repo.FullName
	}

	id := slugify(repo.FullName)
	if skillName != "" {
		id = slugify(repo.FullName) + "-" + slugify(skillName)
	}

	license := ""
	if repo.License != nil {
		license = repo.License.SPDXID
		if license == "" {
			license = repo.License.Name
		}
	}

	owner := strings.Split(repo.FullName, "/")[0]

	// 技能在仓库中的目录路径（如 skills/frontend-design）；仓库根技能为空字符串。
	// 用于生成安装命令、下载地址与按目录下载技能 ZIP。
	dir := path.Dir(mdPath)
	if dir == "." {
		dir = ""
	}

	// 正文优先 SKILL.md 自身（技能定义本体）；仅当 SKILL.md 无实质正文时，
	// 才回退到同目录 README.md（个别仓库的 SKILL.md 只有 frontmatter）。
	body := content
	if readme, err := c.GetReadme(repo.FullName, dir, repo.DefaultBranch); err == nil {
		body = pickBody(content, readme)
	}

	return Skill{
		ID:             id,
		Name:           TitleCase(name),
		Author:         c.OfficialDisplayName(owner),
		Description:    desc,
		Tags:           inferTags(name, desc),
		Category:       inferCategory(name, desc),
		DownloadURL:    skillDownloadURL(repo, dir),
		InstallCommand: installCommand(repo.HTMLURL, dir),
		GithubURL:      repo.HTMLURL,
		GithubStars:    formatStars(repo.Stars),
		License:        license,
		IsOfficial:     c.IsOfficial(owner),
		SkillPath:      dir,
		Content:        ParseContent(string(body)),
	}, nil
}

// defaultOfficialOrgs 内置官方组织（owner → 展示信息）。
// 爬虫启动时作为默认；后台可通过 official_orgs 表动态覆盖（Client.SetOfficialOrgs）。
var defaultOfficialOrgs = buildDefaultOfficialOrgs()

func buildDefaultOfficialOrgs() map[string]OrgInfo {
	m := make(map[string]OrgInfo, len(officialOrgAvatars))
	for owner, avatar := range officialOrgAvatars {
		m[owner] = OrgInfo{Avatar: avatar, DisplayName: officialDisplayNames[owner]}
	}
	return m
}

// officialOrgAvatars 官方组织默认头像 emoji（GitHub owner → emoji）。
// 爬取到的仓库 owner 在此列表内 → 官方技能；否则 → 个人/社区技能。
var officialOrgAvatars = map[string]string{
	"anthropics":          "🅰️",
	"openai":              "🤖",
	"microsoft":           "🪟",
	"vercel":              "▲",
	"google":              "🇬",
	"googlecloudplatform": "🇬",
	"deepmind":            "🇬",
	"github":              "🐙",
	"cloudflare":          "☁️",
	"figma":               "🎨",
	"notion":              "📝",
	"stripe":              "💳",
	"aws":                 "☁️",
	"aws-samples":         "☁️",
	"sst":                 "▲",
	"meta":                "🔵",
	"facebook":            "🔵",
	"huggingface":         "🤗",
	"ibm":                 "🔷",
	"oracle":              "🟠",
	"apple":               "🍎",
	"netflix":             "🎬",
	"linkedin":            "💼",
	"amazon":              "🛒",
	"alibaba":             "🅰️",
	"tencent":             "🐧",
	"baidu":               "🐻",
	"xai":                 "🐦",
	"mistralai":           "🌀",
	"cohere":              "🌊",
	"databricks":          "🔥",
	"snowflake":           "❄️",
	"anthropic":           "🅰️",
}

// officialDisplayNames 官方组织默认展示名（GitHub owner → 显示名）。
// 与 mcpservers.org 等站点的作者展示一致（如 anthropics → Anthropic）。
var officialDisplayNames = map[string]string{
	"anthropics":          "Anthropic",
	"openai":              "OpenAI",
	"microsoft":           "Microsoft",
	"vercel":              "Vercel",
	"google":              "Google",
	"googlecloudplatform": "Google",
	"deepmind":            "DeepMind",
	"github":              "GitHub",
	"cloudflare":          "Cloudflare",
	"figma":               "Figma",
	"notion":              "Notion",
	"stripe":              "Stripe",
	"aws":                 "AWS",
	"aws-samples":         "AWS",
	"sst":                 "SST",
	"meta":                "Meta",
	"facebook":            "Meta",
	"huggingface":         "Hugging Face",
	"ibm":                 "IBM",
	"oracle":              "Oracle",
	"apple":               "Apple",
	"netflix":             "Netflix",
	"linkedin":            "LinkedIn",
	"amazon":              "Amazon",
	"alibaba":             "Alibaba",
	"tencent":             "Tencent",
	"baidu":               "Baidu",
	"xai":                 "xAI",
	"mistralai":           "Mistral AI",
	"cohere":              "Cohere",
	"databricks":          "Databricks",
	"snowflake":           "Snowflake",
}

// IsOfficialOrg 判断仓库 owner 是否为内置默认官方组织。
// 仅供 import 等外部工具使用；运行时爬虫使用 Client.IsOfficial（数据来自数据库动态配置）。
func IsOfficialOrg(owner string) bool {
	_, ok := defaultOfficialOrgs[strings.ToLower(owner)]
	return ok
}

// OfficialAvatar 返回内置官方组织的头像 emoji；非官方组织返回空字符串。
// 兼容 GitHub owner 与展示名（如 "Hugging Face"、"Anthropic"）。
func OfficialAvatar(owner string) string {
	lower := strings.ToLower(owner)
	if info, ok := defaultOfficialOrgs[lower]; ok {
		return info.Avatar
	}
	for _, info := range defaultOfficialOrgs {
		if info.DisplayName != "" && strings.EqualFold(info.DisplayName, owner) {
			return info.Avatar
		}
	}
	return ""
}

// OfficialDisplayName 返回内置官方组织的展示名；非官方组织返回 GitHub 用户名本身。
func OfficialDisplayName(owner string) string {
	if info, ok := defaultOfficialOrgs[strings.ToLower(owner)]; ok && info.DisplayName != "" {
		return info.DisplayName
	}
	return owner
}

// installCommand 生成官方 skills CLI 的安装命令。
// 仓库根技能：npx skills add <repo-url>；目录技能：npx skills add <repo-url> --skill <目录名>。
func installCommand(repoURL, skillPath string) string {
	cmd := "npx skills add " + repoURL
	if skillPath != "" {
		cmd += " --skill " + path.Base(skillPath)
	}
	return cmd
}

// skillDownloadURL 生成技能下载地址：技能所在目录的 GitHub tree 链接。
// 与 mcpservers.org 等站点的「下载 ZIP」指向一致（配合 DownGit 可下载单个技能目录）；
// 仓库根技能返回仓库主页。
func skillDownloadURL(repo Repo, skillPath string) string {
	if skillPath == "" {
		return repo.HTMLURL
	}
	return fmt.Sprintf("%s/tree/%s/%s", repo.HTMLURL, repo.DefaultBranch, skillPath)
}

// categoryRules 分类推断规则（按顺序匹配，命中即返回）。
var categoryRules = []struct {
	category string
	keywords []string
}{
	{"browser-automation", []string{"playwright", "selenium", "puppeteer", "web automation", "browser automation", "web scraping", "scraping", "scrape", "headless browser"}},
	{"database", []string{"database", "sql", "postgres", "mysql", "redis", "mongo", "sqlite", "storage", "data lake", "cassandra"}},
	{"document", []string{"document", "pdf", "word", "excel", "markdown", "office", "spreadsheet", "presentation", "docx", "slide", "google docs"}},
	{"media", []string{"image", "video", "audio", "media", "photo", "3d", "animation", "music", "voice"}},
	{"creative", []string{"design", "ui", "ux", "art", "creative", "brand", "logo", "figma", "illustration"}},
	{"productivity", []string{"productivity", "task", "meeting", "notes", "calendar", "email", "gmail", "notion", "workflow", "schedule", "todo", "reminder", "inbox"}},
	{"testing", []string{"testing", "test automation", "unit test", "e2e", "end-to-end", "qa", "test coverage", "test suite", "regression"}},
	{"security", []string{"security", "auth", "encryption", "penetration", "vulnerability", "cyber"}},
	{"development", []string{"code", "develop", "programming", "git", "github", "api", "backend", "frontend", "engineering", "debug", "refactor", "sdk"}},
}

// inferCategory 从名称与描述推断技能分类；无法匹配时默认 development。
func inferCategory(name, desc string) string {
	text := strings.ToLower(name + " " + desc)
	for _, rule := range categoryRules {
		for _, kw := range rule.keywords {
			if strings.Contains(text, kw) {
				return rule.category
			}
		}
	}
	return "development"
}

// formatStars 将 star 数格式化为可读字符串（如 239、1.2k、3.4k）。
func formatStars(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

// inferTags 从名称与描述推断标签（简单关键词匹配）。
func inferTags(name, desc string) []string {
	text := strings.ToLower(name + " " + desc)
	tags := make([]string, 0)
	for _, kw := range []string{
		"development", "testing", "document", "browser-automation",
		"database", "creative", "media", "productivity", "security",
		"web-scraping", "code-review", "research", "accessibility",
		"payment", "audio", "api",
	} {
		if strings.Contains(text, kw) {
			tags = append(tags, kw)
		}
	}
	return tags
}

// Slugify 转小写并将非字母数字字符替换为 -。
func Slugify(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// slugify 保留旧名以兼容。
func slugify(s string) string { return Slugify(s) }
