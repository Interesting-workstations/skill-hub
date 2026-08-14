package crawler

// ContentSection SKILL.md 正文的一个内容区块（## 标题 + 段落）。
type ContentSection struct {
	Heading string   `json:"heading"`
	Body    []string `json:"body"`
}

// Skill 爬取到的技能数据，与 server 数据模型（domain.Skill）字段对齐。
type Skill struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Author         string   `json:"author"`
	Description    string   `json:"description"`
	Tags           []string `json:"tags"`
	Category       string   `json:"category"`
	DownloadURL    string   `json:"downloadUrl"`
	IsOfficial     bool     `json:"isOfficial"`
	IsFeatured     bool     `json:"isFeatured"`
	InstallCommand string   `json:"installCommand,omitempty"`
	GithubURL      string   `json:"githubUrl,omitempty"`
	GithubStars    string   `json:"githubStars,omitempty"`
	License        string   `json:"license,omitempty"`
	// SkillPath 技能在仓库中的目录路径（如 skills/frontend-design；仓库根技能为空字符串）。
	// 用于生成安装命令与按目录下载技能 ZIP。
	SkillPath string           `json:"skillPath,omitempty"`
	Content   []ContentSection `json:"content,omitempty"`
}

// Repo GitHub 仓库信息。
type Repo struct {
	FullName      string       `json:"full_name"`
	Description   string       `json:"description"`
	Stars         int          `json:"stargazers_count"`
	License       *LicenseInfo `json:"license"`
	HTMLURL       string       `json:"html_url"`
	DefaultBranch string       `json:"default_branch"`
}

// LicenseInfo 仓库许可信息（GitHub API 中 license 为对象）。
type LicenseInfo struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdx_id"`
}

// ContentEntry GitHub contents API 返回的目录项。
type ContentEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}
