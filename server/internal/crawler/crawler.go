package crawler

import (
	"fmt"
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
func (c *Client) CrawlDetailed(opts CrawlOptions) ([]Skill, []RepoFailure, error) {
	seen := make(map[string]bool)
	var skills []Skill
	var failures []RepoFailure

	// 1. 指定仓库（通常是官方仓库）
	for _, fullName := range opts.Repos {
		repo, err := c.GetRepo(fullName)
		if err != nil {
			failures = append(failures, RepoFailure{Repo: fullName, Reason: "获取仓库失败", Error: err.Error()})
			continue
		}
		got, err := c.crawlRepo(repo)
		if err != nil {
			failures = append(failures, RepoFailure{Repo: fullName, Reason: "解析仓库失败", Error: err.Error()})
			continue
		}
		for _, s := range got {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			skills = append(skills, s)
		}
	}

	// 2. 搜索结果补充
	repos, err := c.SearchRepos(opts.Query, opts.PerPage, 1)
	if err != nil {
		return nil, failures, fmt.Errorf("搜索仓库失败: %w", err)
	}
	if len(repos) > opts.Limit {
		repos = repos[:opts.Limit]
	}
	for _, repo := range repos {
		got, err := c.crawlRepo(repo)
		if err != nil {
			failures = append(failures, RepoFailure{Repo: repo.FullName, Reason: "解析仓库失败", Error: err.Error()})
			continue
		}
		for _, s := range got {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			skills = append(skills, s)
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

// buildSkill 读取 SKILL.md 并组装 Skill 记录。
// 官方组织仓库的技能会标记为官方（IsOfficial），并根据内容推断分类。
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
	return Skill{
		ID:          id,
		Name:        name,
		Author:      owner,
		Description: desc,
		Tags:        inferTags(name, desc),
		Category:    inferCategory(name, desc),
		DownloadURL: repo.HTMLURL + "/archive/refs/heads/" + repo.DefaultBranch + ".zip",
		GithubURL:   repo.HTMLURL,
		GithubStars: formatStars(repo.Stars),
		License:     license,
		IsOfficial:  IsOfficialOrg(owner),
	}, nil
}

// officialOrgs 官方组织及其在站点上展示的头像 emoji。
var officialOrgs = map[string]string{
	"anthropics":          "🅰️",
	"openai":              "🤖",
	"microsoft":           "🪟",
	"vercel":              "▲",
	"google":              "🇬",
	"googlecloudplatform": "🇬",
	"github":              "🐙",
	"cloudflare":          "☁️",
	"figma":               "🎨",
	"notion":              "📝",
	"stripe":              "💳",
	"aws":                 "☁️",
	"aws-samples":         "☁️",
	"sst":                 "▲",
}

// IsOfficialOrg 判断仓库 owner 是否为官方组织。
func IsOfficialOrg(owner string) bool {
	_, ok := officialOrgs[strings.ToLower(owner)]
	return ok
}

// OfficialAvatar 返回官方组织的头像 emoji；非官方组织返回空字符串。
func OfficialAvatar(owner string) string {
	return officialOrgs[strings.ToLower(owner)]
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
