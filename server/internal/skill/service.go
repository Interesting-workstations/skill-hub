package skill

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// 精选规则常量。
const (
	// DefaultFeaturedLimit 精选技能总数上限。
	DefaultFeaturedLimit = 6
	// maxFeaturedPerCategory 每个分类最多入选的精选技能数（保证分类多样性）。
	maxFeaturedPerCategory = 2
)

// Service 是技能业务逻辑层，不感知 HTTP 细节。
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListSkills 按筛选条件返回技能列表。
// filter 为空字段表示不筛选；featured 筛选由精选规则生成；
// Query 为关键词，匹配名称/作者/描述/标签/分类（不区分大小写）；
// Limit/Offset 支持分页（Limit<=0 返回全部）。
func (s *Service) ListSkills(filter SkillFilter) []domain.Skill {
	if filter.Featured {
		return paginate(s.filteredFeatured(filter), filter.Limit, filter.Offset)
	}
	// 关键词搜索：走 SQL LIKE（相关性排序 + LIMIT/OFFSET 分页），避免全量加载 2000+ 条含 content 的记录
	if strings.TrimSpace(filter.Query) != "" {
		limit := filter.Limit
		if limit <= 0 {
			limit = 20
		}
		return s.repo.SearchSkills(filter.Query, filter.Category, filter.Author, filter.Official, limit, filter.Offset)
	}
	all := s.repo.AllSkills()
	if filter.Category == "" && filter.Author == "" && !filter.Official {
		return paginate(all, filter.Limit, filter.Offset)
	}
	out := make([]domain.Skill, 0, len(all))
	for _, sk := range all {
		if filter.Category != "" && sk.Category != filter.Category {
			continue
		}
		if filter.Author != "" && !strings.EqualFold(sk.Author, filter.Author) {
			continue
		}
		if filter.Official && !sk.IsOfficial {
			continue
		}
		out = append(out, sk)
	}
	return paginate(out, filter.Limit, filter.Offset)
}

// paginate 对技能列表应用 limit/offset 分页（limit<=0 原样返回）。
func paginate(list []domain.Skill, limit, offset int) []domain.Skill {
	if limit <= 0 {
		return list
	}
	start := offset
	if start < 0 {
		start = 0
	}
	if start >= len(list) {
		return nil
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	return list[start:end]
}

// skillMatches 判断技能是否匹配搜索关键词（名称/作者/描述/标签/分类，不区分大小写）。
func skillMatches(sk domain.Skill, q string) bool {
	if strings.Contains(strings.ToLower(sk.Name), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sk.Author), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sk.Description), q) {
		return true
	}
	if strings.Contains(strings.ToLower(sk.Category), q) {
		return true
	}
	for _, tag := range sk.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

// filteredFeatured 在精选规则结果之上应用其余筛选。
func (s *Service) filteredFeatured(filter SkillFilter) []domain.Skill {
	featured := s.FeaturedSkills(DefaultFeaturedLimit)
	if filter.Category == "" && filter.Author == "" && !filter.Official && filter.Query == "" {
		return featured
	}
	q := strings.ToLower(strings.TrimSpace(filter.Query))
	out := make([]domain.Skill, 0, len(featured))
	for _, sk := range featured {
		if filter.Category != "" && sk.Category != filter.Category {
			continue
		}
		if filter.Author != "" && !strings.EqualFold(sk.Author, filter.Author) {
			continue
		}
		if filter.Official && !sk.IsOfficial {
			continue
		}
		if q != "" && !skillMatches(sk, q) {
			continue
		}
		out = append(out, sk)
	}
	return out
}

// FeaturedSkills 按精选规则返回技能：
//  1. 官方技能优先，同一优先级下按 GitHub 星标数降序；
//  2. 每个分类最多 maxFeaturedPerCategory 个，保证分类多样性。
func (s *Service) FeaturedSkills(limit int) []domain.Skill {
	if limit <= 0 {
		limit = DefaultFeaturedLimit
	}
	pool := s.repo.AllSkills()
	sort.SliceStable(pool, func(i, j int) bool {
		if pool[i].IsOfficial != pool[j].IsOfficial {
			return pool[i].IsOfficial
		}
		return parseStars(pool[i].GithubStars) > parseStars(pool[j].GithubStars)
	})

	picked := make([]domain.Skill, 0, limit)
	used := make(map[string]int)
	for _, sk := range pool {
		if used[sk.Category] >= maxFeaturedPerCategory {
			continue
		}
		picked = append(picked, sk)
		used[sk.Category]++
		if len(picked) >= limit {
			break
		}
	}
	return picked
}

// parseStars 将 "168.1k"、"2.3m"、"861" 等星标格式解析为数值。
func parseStars(s string) float64 {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0
	}
	mult := 1.0
	switch {
	case strings.HasSuffix(s, "k"):
		mult = 1000
		s = strings.TrimSuffix(s, "k")
	case strings.HasSuffix(s, "m"):
		mult = 1_000_000
		s = strings.TrimSuffix(s, "m")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v * mult
}

// GetSkill 按 ID 获取技能详情。
func (s *Service) GetSkill(id string) (domain.Skill, bool) {
	return s.repo.SkillByID(id)
}

// ListAuthors 返回全部作者。
func (s *Service) ListAuthors() []domain.Author {
	return s.repo.AllAuthors()
}

// GetAuthor 返回作者及其发布的技能。
func (s *Service) GetAuthor(slug string) (AuthorDetail, bool) {
	for _, a := range s.repo.AllAuthors() {
		if a.Slug == slug {
			skills := s.ListSkills(SkillFilter{Author: a.Name})
			return AuthorDetail{Author: a, Skills: skills}, true
		}
	}
	return AuthorDetail{}, false
}

// ListCategories 返回全部分类。
func (s *Service) ListCategories() []domain.Category {
	return s.repo.AllCategories()
}

// ListOfficialOrgs 返回官方组织概览（官网「官方技能/官方组织」区块统一数据源）。
func (s *Service) ListOfficialOrgs() []domain.OfficialOrgSummary {
	return s.repo.OfficialOrgSummaries()
}

// OrgLogoURL 返回官方组织显式锁定的 logo 来源（官网等），未设置时返回空（走 GitHub 头像）。
// 供 orgLogo 图片接口确定下载来源：优先读本地缓存，无缓存时按该来源下载。
func (s *Service) OrgLogoURL(owner string) (string, bool) {
	return s.repo.OfficialOrgLogoURL(owner)
}

// GetCategory 按 slug 返回分类及其技能。
func (s *Service) GetCategory(slug string) (domain.Category, bool) {
	for _, c := range s.repo.AllCategories() {
		if c.Slug == slug {
			return c, true
		}
	}
	return domain.Category{}, false
}

// ---------- 公开内容（文章 / 站点配置 / SEO / 提交技能） ----------

// ListArticles 返回全部已发布文章。
func (s *Service) ListArticles() []domain.Article {
	return s.repo.ListArticles()
}

// GetArticle 按 ID 返回已发布文章（countView=true 时按 IP 当天去重累加浏览量）。
func (s *Service) GetArticle(id, ip string, countView bool) (domain.Article, bool) {
	return s.repo.ArticleByID(id, ip, countView)
}

// GetSiteConfig 返回站点配置。
func (s *Service) GetSiteConfig() (domain.SiteConfig, bool) {
	return s.repo.GetSiteConfig()
}

// GetSeo 返回 SEO 配置。
func (s *Service) GetSeo() (domain.SeoConfig, bool) {
	return s.repo.GetSeo()
}

// SubmitSkillInput 用户提交技能的表单数据。
type SubmitSkillInput struct {
	Name        string   `json:"name"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	DownloadURL string   `json:"downloadUrl"`
	GithubURL   string   `json:"githubUrl"`
}

// SubmitSkill 校验并保存用户提交的技能（进入待审核状态，官方标记恒为 false）。
func (s *Service) SubmitSkill(in SubmitSkillInput) (domain.Skill, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Author = strings.TrimSpace(in.Author)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" || in.Author == "" || in.Description == "" {
		return domain.Skill{}, fmt.Errorf("技能名称、作者与描述必填")
	}
	if len(in.Name) > 120 || len(in.Author) > 120 {
		return domain.Skill{}, fmt.Errorf("名称或作者过长")
	}
	if in.Category == "" {
		in.Category = "other"
	}
	if in.DownloadURL == "" {
		in.DownloadURL = "#"
	}
	// 分类 slug 化；作者允许中文，仅生成 ID 用
	cat := Slugify(in.Category)
	if cat == "" {
		cat = "other"
	}
	// ID：author-name 形式，与爬虫规则一致
	base := Slugify(in.Author) + "-" + Slugify(in.Name)
	id := base
	for i := 2; ; i++ {
		if _, ok := s.repo.SkillByIDAny(id); !ok {
			break
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
	skill := domain.Skill{
		ID:          id,
		Name:        in.Name,
		Author:      in.Author,
		Description: in.Description,
		Category:    cat,
		Tags:        in.Tags,
		DownloadURL: in.DownloadURL,
		GithubURL:   in.GithubURL,
		IsOfficial:  false,
		IsFeatured:  false,
	}
	if err := s.repo.SubmitSkill(&skill); err != nil {
		return domain.Skill{}, fmt.Errorf("提交失败，请稍后重试")
	}
	return skill, nil
}

// Stats 返回站点聚合统计（全部基于数据库实时计算）。
func (s *Service) Stats() Stats {
	all := s.repo.AllSkills()
	official := 0
	for _, sk := range all {
		if sk.IsOfficial {
			official++
		}
	}
	return Stats{
		TotalSkills:     len(all),
		TotalAuthors:    len(s.repo.OfficialOrgSummaries()), // 官方作者 = 启用中的官方组织数
		TotalCategories: len(s.repo.AllCategories()),
		OfficialSkills:  official,
		FeaturedSkills:  len(s.FeaturedSkills(DefaultFeaturedLimit)),
	}
}

// Slugify 转小写并将非字母数字字符替换为 -（与爬虫 ID 规则一致）。
// 支持中文保留（用于 ID 中保留中文作者/技能名）时请改用 slugifyASCII。
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r >= 0x4e00 && r <= 0x9fff: // CJK 统一汉字
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
