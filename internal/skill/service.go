package skill

import (
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// Service 是技能业务逻辑层，不感知 HTTP 细节。
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListSkills 按筛选条件返回技能列表。
// filter 为空字段表示不筛选。
func (s *Service) ListSkills(filter SkillFilter) []domain.Skill {
	all := s.repo.AllSkills()
	if filter.Category == "" && filter.Author == "" && !filter.Official && !filter.Featured {
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
		if filter.Featured && !sk.IsFeatured {
			continue
		}
		out = append(out, sk)
	}
	return out
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

// Stats 返回站点聚合统计。
func (s *Service) Stats() Stats {
	all := s.repo.AllSkills()
	official, featured := 0, 0
	for _, sk := range all {
		if sk.IsOfficial {
			official++
		}
		if sk.IsFeatured {
			featured++
		}
	}
	return Stats{
		TotalSkills:     len(all),
		TotalAuthors:    len(s.repo.AllAuthors()),
		TotalCategories: len(s.repo.AllCategories()),
		OfficialSkills:  official,
		FeaturedSkills:  featured,
	}
}
