// Package skill 是技能资源库核心业务模块。
// 采用分层：Handler → Service → Repository（按后端架构规范）。
package skill

import (
	"encoding/json"
	"os"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// Repository 定义技能数据的访问接口。
// 业务层只依赖此接口，不关心底层存储实现（内存/数据库均可替换）。
type Repository interface {
	// AllSkills 返回全部技能（精选 + 各分类）。
	AllSkills() []domain.Skill
	// SkillByID 按 ID 查询技能。
	SkillByID(id string) (domain.Skill, bool)
	// AllAuthors 返回全部作者。
	AllAuthors() []domain.Author
	// AllCategories 返回全部分类（含分类下技能）。
	AllCategories() []domain.Category

	// ListArticles 返回全部已发布文章。
	ListArticles() []domain.Article
	// ArticleByID 按 ID 查询文章。
	ArticleByID(id string) (domain.Article, bool)
	// GetSiteConfig 返回站点配置。
	GetSiteConfig() (domain.SiteConfig, bool)
	// GetSeo 返回 SEO 配置。
	GetSeo() (domain.SeoConfig, bool)
	// SubmitSkill 保存用户提交的技能（进入待审核状态）。
	SubmitSkill(s *domain.Skill) error
}

// memoryRepo 基于内存 + JSON 种子文件的实现。
type memoryRepo struct {
	store domain.Store
}

// NewMemoryRepository 从 JSON 种子文件加载数据，返回内存 Repository。
func NewMemoryRepository(dataPath string) (Repository, error) {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}
	var store domain.Store
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &memoryRepo{store: store}, nil
}

func (r *memoryRepo) AllSkills() []domain.Skill {
	skills := make([]domain.Skill, 0, len(r.store.FeaturedSkills))
	skills = append(skills, r.store.FeaturedSkills...)
	for _, c := range r.store.SkillCategories {
		skills = append(skills, c.Skills...)
	}
	return skills
}

func (r *memoryRepo) SkillByID(id string) (domain.Skill, bool) {
	for _, s := range r.AllSkills() {
		if s.ID == id {
			return s, true
		}
	}
	return domain.Skill{}, false
}

func (r *memoryRepo) AllAuthors() []domain.Author {
	authors := make([]domain.Author, len(r.store.Authors))
	copy(authors, r.store.Authors)
	// 实时统计官方技能数
	official := make(map[string]int)
	for _, s := range r.AllSkills() {
		if s.IsOfficial {
			official[s.Author]++
		}
	}
	for i := range authors {
		authors[i].OfficialSkills = official[authors[i].Name]
	}
	return authors
}

func (r *memoryRepo) AllCategories() []domain.Category {
	return r.store.SkillCategories
}

// ---------- 公开内容（内存实现返回空/默认值） ----------

func (r *memoryRepo) ListArticles() []domain.Article { return nil }
func (r *memoryRepo) ArticleByID(id string) (domain.Article, bool) {
	return domain.Article{}, false
}
func (r *memoryRepo) GetSiteConfig() (domain.SiteConfig, bool) {
	return domain.SiteConfig{}, false
}
func (r *memoryRepo) GetSeo() (domain.SeoConfig, bool) {
	return domain.SeoConfig{}, false
}
func (r *memoryRepo) SubmitSkill(s *domain.Skill) error {
	return nil
}
