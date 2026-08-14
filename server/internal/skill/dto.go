package skill

import "github.com/Interesting-workstations/skill-hub/server/internal/domain"

// SkillFilter 技能列表筛选条件。
type SkillFilter struct {
	Category string
	Author   string
	Official bool
	Featured bool
	Query    string
}

// AuthorDetail 作者及其技能（作者详情响应）。
type AuthorDetail struct {
	Author domain.Author  `json:"author"`
	Skills []domain.Skill `json:"skills"`
}

// Stats 站点统计数据（Hero 等聚合展示）。
type Stats struct {
	TotalSkills     int `json:"totalSkills"`     // 已发布技能总数
	TotalAuthors    int `json:"totalAuthors"`    // 官方作者数 = 启用中的官方组织数
	TotalCategories int `json:"totalCategories"` // 技能分类数
	OfficialSkills  int `json:"officialSkills"`  // 官方技能数（is_official=1 且已发布）
	FeaturedSkills  int `json:"featuredSkills"`  // 精选技能数
}
