package skill

import (
	"testing"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// newTestService 用内存数据构造 Service（不依赖 JSON 文件）。
func newTestService() *Service {
	repo := &memoryRepo{store: domain.Store{
		Authors: []domain.Author{
			{Name: "Anthropic", Avatar: "🅰️", Slug: "anthropic"},
		},
		FeaturedSkills: []domain.Skill{
			{ID: "featured-1", Name: "Featured One", Author: "Anthropic", Category: "featured", IsOfficial: true, IsFeatured: true},
		},
		SkillCategories: []domain.Category{
			{
				Name:  "开发技能",
				Slug:  "development",
				Count: 2,
				Skills: []domain.Skill{
					{ID: "dev-1", Name: "Dev One", Author: "Alice", Category: "development"},
					{ID: "dev-2", Name: "Dev Two", Author: "Anthropic", Category: "development", IsOfficial: true},
				},
			},
		},
	}}
	return NewService(repo)
}

func TestListSkills_All(t *testing.T) {
	svc := newTestService()
	got := svc.ListSkills(SkillFilter{})
	if len(got) != 3 {
		t.Fatalf("期望 3 个技能，实际 %d", len(got))
	}
}

func TestListSkills_FilterOfficial(t *testing.T) {
	svc := newTestService()
	got := svc.ListSkills(SkillFilter{Official: true})
	if len(got) != 2 {
		t.Fatalf("期望 2 个官方技能，实际 %d", len(got))
	}
}

func TestListSkills_FilterFeatured(t *testing.T) {
	svc := newTestService()
	got := svc.ListSkills(SkillFilter{Featured: true})
	if len(got) != 1 || got[0].ID != "featured-1" {
		t.Fatalf("期望 1 个精选技能，实际 %d", len(got))
	}
}

func TestListSkills_FilterCategory(t *testing.T) {
	svc := newTestService()
	got := svc.ListSkills(SkillFilter{Category: "development"})
	if len(got) != 2 {
		t.Fatalf("期望 development 分类 2 个技能，实际 %d", len(got))
	}
}

func TestListSkills_FilterAuthor(t *testing.T) {
	svc := newTestService()
	got := svc.ListSkills(SkillFilter{Author: "anthropic"})
	// 作者匹配不区分大小写，应匹配 Anthropic 的 2 个技能
	if len(got) != 2 {
		t.Fatalf("期望 Anthropic 的 2 个技能，实际 %d", len(got))
	}
}

func TestGetSkill(t *testing.T) {
	svc := newTestService()
	s, ok := svc.GetSkill("dev-1")
	if !ok || s.Name != "Dev One" {
		t.Fatalf("按 ID 查询失败: ok=%v", ok)
	}
	if _, ok := svc.GetSkill("not-exist"); ok {
		t.Fatal("不存在的 ID 应返回 ok=false")
	}
}

func TestGetAuthor(t *testing.T) {
	svc := newTestService()
	detail, ok := svc.GetAuthor("anthropic")
	if !ok {
		t.Fatal("作者应存在")
	}
	if len(detail.Skills) != 2 {
		t.Fatalf("期望 Anthropic 2 个技能，实际 %d", len(detail.Skills))
	}
	if _, ok := svc.GetAuthor("nobody"); ok {
		t.Fatal("不存在的作者应返回 ok=false")
	}
}

func TestStats(t *testing.T) {
	svc := newTestService()
	st := svc.Stats()
	if st.TotalSkills != 3 {
		t.Fatalf("期望技能总数 3，实际 %d", st.TotalSkills)
	}
	if st.TotalAuthors != 1 {
		t.Fatalf("期望作者数 1，实际 %d", st.TotalAuthors)
	}
	if st.TotalCategories != 1 {
		t.Fatalf("期望分类数 1，实际 %d", st.TotalCategories)
	}
	if st.OfficialSkills != 2 {
		t.Fatalf("期望官方技能 2，实际 %d", st.OfficialSkills)
	}
	if st.FeaturedSkills != 1 {
		t.Fatalf("期望精选技能 1，实际 %d", st.FeaturedSkills)
	}
}
