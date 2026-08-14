package admin

import "testing"

func TestFilterDuplicates(t *testing.T) {
	mk := func(id, name, author, url string) InsertSkill {
		return InsertSkill{ID: id, Name: name, Author: author, GithubURL: url}
	}

	t.Run("无重复全部保留", func(t *testing.T) {
		in := []InsertSkill{
			mk("a", "A", "org1", "https://github.com/org1/repo"),
			mk("b", "B", "org2", "https://github.com/org2/repo"),
		}
		out, dup := filterDuplicates(in, nil, nil)
		if len(out) != 2 || dup != 0 {
			t.Fatalf("期望保留 2 条、重复 0，实际 out=%d dup=%d", len(out), dup)
		}
	})

	t.Run("批内同源同名只保留第一条", func(t *testing.T) {
		in := []InsertSkill{
			mk("a1", "Skill", "org", "https://github.com/org/repo"),
			mk("a2", "Skill", "org", "https://github.com/org/repo"), // ID 不同但同源同名
		}
		out, dup := filterDuplicates(in, nil, nil)
		if len(out) != 1 || dup != 1 {
			t.Fatalf("期望保留 1 条、重复 1，实际 out=%d dup=%d", len(out), dup)
		}
		if out[0].ID != "a1" {
			t.Fatalf("应保留第一条 a1，实际 %s", out[0].ID)
		}
	})

	t.Run("不同仓库同名技能不算重复", func(t *testing.T) {
		in := []InsertSkill{
			mk("a", "Skill", "org1", "https://github.com/org1/repo"),
			mk("b", "Skill", "org2", "https://github.com/org2/repo"),
		}
		out, dup := filterDuplicates(in, nil, nil)
		if len(out) != 2 || dup != 0 {
			t.Fatalf("不同来源同名应全部保留，实际 out=%d dup=%d", len(out), dup)
		}
	})

	t.Run("数据库 ID 已存在则跳过", func(t *testing.T) {
		in := []InsertSkill{mk("a", "A", "org", "https://github.com/org/repo")}
		out, dup := filterDuplicates(in, map[string]bool{"a": true}, nil)
		if len(out) != 0 || dup != 1 {
			t.Fatalf("期望全部跳过，实际 out=%d dup=%d", len(out), dup)
		}
	})

	t.Run("数据库同源同名已存在则跳过", func(t *testing.T) {
		in := []InsertSkill{mk("new", "Skill", "org", "https://github.com/org/repo")}
		out, dup := filterDuplicates(in, nil, map[string]bool{"Skill|org|https://github.com/org/repo": true})
		if len(out) != 0 || dup != 1 {
			t.Fatalf("期望全部跳过，实际 out=%d dup=%d", len(out), dup)
		}
	})

	t.Run("组合场景", func(t *testing.T) {
		in := []InsertSkill{
			mk("a", "A", "org1", "https://github.com/org1/repo"),
			mk("a2", "A", "org1", "https://github.com/org1/repo"), // 批内重复
			mk("b", "B", "org2", "https://github.com/org2/repo"),  // ID 已存在
			mk("c", "C", "org3", "https://github.com/org3/repo"),  // 正常新增
			mk("d", "D", "org4", "https://github.com/org4/repo"),  // 同源已存在
		}
		existIDs := map[string]bool{"b": true}
		existSources := map[string]bool{"D|org4|https://github.com/org4/repo": true}
		out, dup := filterDuplicates(in, existIDs, existSources)
		if len(out) != 2 || dup != 3 {
			t.Fatalf("期望保留 2 条（a、c）、重复 3，实际 out=%d dup=%d", len(out), dup)
		}
		for _, s := range out {
			if s.ID != "a" && s.ID != "c" {
				t.Fatalf("保留列表不应包含 %s", s.ID)
			}
		}
	})
}
