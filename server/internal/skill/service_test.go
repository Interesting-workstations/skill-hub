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
	// 精选由规则生成：官方优先 + 分类多样性
	got := svc.ListSkills(SkillFilter{Featured: true})
	if len(got) != 3 {
		t.Fatalf("期望 3 个精选技能，实际 %d", len(got))
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
	// 精选数量按规则计算（官方 2 个 + 补充非官方 1 个）
	if st.FeaturedSkills != 3 {
		t.Fatalf("期望精选技能 3，实际 %d", st.FeaturedSkills)
	}
}

// featuredRuleService 构造用于测试精选规则的数据：
// 多个官方/非官方、不同星标、跨分类。
func featuredRuleService() *Service {
	repo := &memoryRepo{store: domain.Store{
		FeaturedSkills: []domain.Skill{
			{ID: "f1", Name: "F1", Author: "A", Category: "featured", IsOfficial: true, GithubStars: "1k"},
			{ID: "f2", Name: "F2", Author: "A", Category: "featured", IsOfficial: true, GithubStars: "500"},
		},
		SkillCategories: []domain.Category{
			{Slug: "dev", Name: "Dev", Skills: []domain.Skill{
				{ID: "d1", Name: "D1", Author: "A", Category: "dev", IsOfficial: true, GithubStars: "2k"},
				{ID: "d2", Name: "D2", Author: "A", Category: "dev", GithubStars: "10k"},
				{ID: "d3", Name: "D3", Author: "A", Category: "dev", GithubStars: "3k"},
			}},
			{Slug: "media", Name: "Media", Skills: []domain.Skill{
				{ID: "m1", Name: "M1", Author: "A", Category: "media", GithubStars: "50k"},
				{ID: "m2", Name: "M2", Author: "A", Category: "media", GithubStars: "20"},
			}},
		},
	}}
	return NewService(repo)
}

// 验证规则：官方优先 → 星标降序 → 分类多样性 → 非官方补齐。
func TestFeaturedSkills_Rule(t *testing.T) {
	svc := featuredRuleService()
	got := svc.FeaturedSkills(6)
	want := []string{"d1", "f1", "f2", "m1", "d2", "m2"}
	if len(got) != len(want) {
		t.Fatalf("期望 %d 个精选，实际 %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].ID != w {
			t.Fatalf("精选顺序 %v，期望 %v", skillIDs(got), want)
		}
	}
}

func TestFeaturedSkills_Shortfall(t *testing.T) {
	svc := featuredRuleService()
	if got := len(svc.FeaturedSkills(2)); got != 2 {
		t.Fatalf("limit=2 期望 2 个，实际 %d", got)
	}
	if got := len(svc.FeaturedSkills(0)); got != DefaultFeaturedLimit {
		t.Fatalf("limit=0 期望默认 %d 个，实际 %d", DefaultFeaturedLimit, got)
	}
}

func TestParseStars(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"861", 861},
		{"168.1k", 168100},
		{"2.3k", 2300},
		{"1.2m", 1200000},
		{"", 0},
		{"abc", 0},
	}
	for _, c := range cases {
		if got := parseStars(c.in); got != c.want {
			t.Fatalf("parseStars(%q) = %v，期望 %v", c.in, got, c.want)
		}
	}
}

func skillIDs(skills []domain.Skill) []string {
	ids := make([]string, 0, len(skills))
	for _, s := range skills {
		ids = append(ids, s.ID)
	}
	return ids
}
