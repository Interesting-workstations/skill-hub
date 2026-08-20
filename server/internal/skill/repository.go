// Package skill 是技能资源库核心业务模块。
// 采用分层：Handler → Service → Repository（按后端架构规范）。
package skill

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// Repository 定义技能数据的访问接口。
// 业务层只依赖此接口，不关心底层存储实现（内存/数据库均可替换）。
type Repository interface {
	// AllSkills 返回全部技能（精选 + 各分类）。
	AllSkills() []domain.Skill
	// SearchSkills 按关键词搜索已发布技能（名称/作者/描述/标签/分类，LIKE 匹配），
	// 官方优先 + 高星排序，返回 limit 条（支持 offset 分页，避免搜索全量加载）。
	SearchSkills(query, category, author string, official bool, limit, offset int) []domain.Skill
	// SkillByID 按 ID 查询已发布技能（官网可见）。
	SkillByID(id string) (domain.Skill, bool)
	// SkillByIDAny 按 ID 查询技能（不限状态，供内部查重/管理使用）。
	SkillByIDAny(id string) (domain.Skill, bool)
	// AllAuthors 返回全部作者。
	AllAuthors() []domain.Author
	// AllCategories 返回全部分类（含分类下技能）。
	AllCategories() []domain.Category
	// OfficialOrgSummaries 返回官方组织概览（含各组织官方技能数，启用且按排序）。
	OfficialOrgSummaries() []domain.OfficialOrgSummary
	// OfficialOrgLogoURL 返回官方组织显式锁定的 logo 来源 URL（官网等），
	// 未设置（走 GitHub 头像）时返回空串。用于本地图片缓存下载。
	OfficialOrgLogoURL(owner string) (string, bool)

	// ListArticles 返回全部已发布文章。
	ListArticles() []domain.Article
	// ArticleByID 按 ID 查询文章（ip+countView 控制浏览量去重累加）。
	ArticleByID(id, ip string, countView bool) (domain.Article, bool)
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

// SearchSkills 内存实现：仅匹配技能标题（中英文）/作者/标签，支持 offset 分页。
func (r *memoryRepo) SearchSkills(query, category, author string, official bool, limit, offset int) []domain.Skill {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]domain.Skill, 0, limit)
	for _, sk := range r.AllSkills() {
		if category != "" && sk.Category != category {
			continue
		}
		if author != "" && !strings.EqualFold(sk.Author, author) {
			continue
		}
		if official && !sk.IsOfficial {
			continue
		}
		if q != "" && !searchMatches(sk, q) {
			continue
		}
		out = append(out, sk)
		if len(out) >= offset+limit {
			break
		}
	}
	if offset > 0 {
		if offset < len(out) {
			out = out[offset:]
		} else {
			out = nil
		}
	}
	return out
}

// searchMatches 判断技能是否匹配搜索词：仅标题（中英文）/作者/标签。
func searchMatches(sk domain.Skill, q string) bool {
	if strings.Contains(strings.ToLower(sk.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sk.NameZh), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sk.Author), q) {
		return true
	}
	for _, tag := range sk.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func (r *memoryRepo) SkillByID(id string) (domain.Skill, bool) {
	for _, s := range r.AllSkills() {
		if s.ID == id {
			return s, true
		}
	}
	return domain.Skill{}, false
}

// SkillByIDAny 内存实现与 SkillByID 相同（无状态概念）。
func (r *memoryRepo) SkillByIDAny(id string) (domain.Skill, bool) {
	return r.SkillByID(id)
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

// OfficialOrgSummaries 内存实现：以 store.Authors 作为官方组织（与统计语义一致）。
func (r *memoryRepo) OfficialOrgSummaries() []domain.OfficialOrgSummary {
	out := make([]domain.OfficialOrgSummary, 0, len(r.store.Authors))
	for _, a := range r.store.Authors {
		out = append(out, domain.OfficialOrgSummary{
			Owner:         a.Slug,
			DisplayName:   a.Name,
			Avatar:        a.Avatar,
			LogoURL:       "/org-logo/" + a.Slug, // 本地图片缓存路径
			OfficialCount: a.OfficialSkills,
		})
	}
	return out
}

// OfficialOrgLogoURL 内存实现：无独立官方组织表，返回空（走 GitHub 头像）。
func (r *memoryRepo) OfficialOrgLogoURL(owner string) (string, bool) {
	return "", false
}

// ---------- 公开内容（内存实现返回空/默认值） ----------

func (r *memoryRepo) ListArticles() []domain.Article { return nil }
func (r *memoryRepo) ArticleByID(id, ip string, countView bool) (domain.Article, bool) {
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
