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
	TotalSkills     int `json:"totalSkills"`
	TotalAuthors    int `json:"totalAuthors"`
	TotalCategories int `json:"totalCategories"`
	OfficialSkills  int `json:"officialSkills"`
	FeaturedSkills  int `json:"featuredSkills"`
}
