// Package domain 定义技能资源库的核心业务模型。
// Domain 层不依赖任何 HTTP / 框架 / 数据库实现。
package domain

// ContentSection 技能详情中的一个内容区块。
type ContentSection struct {
	Heading string   `json:"heading"`
	Body    []string `json:"body"`
}

// Skill 表示一个可复用的 Agent 技能。
type Skill struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	NameZh         string   `json:"nameZh,omitempty"` // 中文标题（切换语言用，无翻译为空）
	Author         string   `json:"author"`
	Description    string   `json:"description"`
	DescriptionZh  string   `json:"descriptionZh,omitempty"` // 中文描述
	Tags           []string `json:"tags"`
	Category       string   `json:"category"`
	DownloadURL    string   `json:"downloadUrl"`
	IsOfficial     bool     `json:"isOfficial"`
	IsFeatured     bool     `json:"isFeatured"`
	InstallCommand string   `json:"installCommand,omitempty"`
	GithubURL      string   `json:"githubUrl,omitempty"`
	GithubStars    string   `json:"githubStars,omitempty"`
	License        string   `json:"license,omitempty"`
	// SkillPath 技能在仓库中的目录路径（如 skills/frontend-design；仓库根技能为空）。
	// 用于下载接口按目录抓取技能文件生成 ZIP。
	SkillPath string           `json:"skillPath,omitempty"`
	Content   []ContentSection `json:"content,omitempty"`
}

// Author 表示技能作者/官方维护方。
type Author struct {
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	SkillCount int    `json:"skillCount"`
	Slug       string `json:"slug"`
	// OfficialSkills 该作者的官方技能数（实时统计）。
	OfficialSkills int `json:"officialSkills"`
}

// Category 表示技能分类（含该分类下的技能列表）。
type Category struct {
	Name   string  `json:"name"`
	Slug   string  `json:"slug"`
	Count  int     `json:"count"`
	Skills []Skill `json:"skills"`
}

// Store 是种子数据的整体结构。
type Store struct {
	Authors         []Author   `json:"authors"`
	FeaturedSkills  []Skill    `json:"featuredSkills"`
	SkillCategories []Category `json:"skillCategories"`
}
