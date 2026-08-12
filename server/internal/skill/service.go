package skill

import (
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
// filter 为空字段表示不筛选；featured 筛选由精选规则生成。
func (s *Service) ListSkills(filter SkillFilter) []domain.Skill {
	if filter.Featured {
		return s.filteredFeatured(filter)
	}
	all := s.repo.AllSkills()
	if filter.Category == "" && filter.Author == "" && !filter.Official {
		return all
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
	return out
}

// filteredFeatured 在精选规则结果之上应用其余筛选。
func (s *Service) filteredFeatured(filter SkillFilter) []domain.Skill {
	featured := s.FeaturedSkills(DefaultFeaturedLimit)
	if filter.Category == "" && filter.Author == "" && !filter.Official {
		return featured
	}
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

// GetCategory 按 slug 返回分类及其技能。
func (s *Service) GetCategory(slug string) (domain.Category, bool) {
	for _, c := range s.repo.AllCategories() {
		if c.Slug == slug {
			return c, true
		}
	}
	return domain.Category{}, false
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
		TotalAuthors:    len(s.repo.AllAuthors()),
		TotalCategories: len(s.repo.AllCategories()),
		OfficialSkills:  official,
		FeaturedSkills:  len(s.FeaturedSkills(DefaultFeaturedLimit)),
	}
}
