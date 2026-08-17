package crawler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// TestPickBody 验证正文选择：SKILL.md 优先，无正文才回退 README。
func TestPickBody(t *testing.T) {
	skillMD := []byte("---\nname: demo\n---\n# Demo\n## 用法\n执行命令")
	readme := []byte("# Repo README\n项目介绍")

	if got := pickBody(skillMD, readme); string(got) != string(skillMD) {
		t.Fatalf("SKILL.md 有正文时应优先使用 SKILL.md，实际用了 README")
	}

	emptySkillMD := []byte("---\nname: demo\n---\n")
	if got := pickBody(emptySkillMD, readme); string(got) != string(readme) {
		t.Fatalf("SKILL.md 无正文时应回退到 README")
	}

	if got := pickBody(emptySkillMD, nil); string(got) != string(emptySkillMD) {
		t.Fatalf("无 README 时应保持 SKILL.md 自身")
	}
}

// TestParseQuerySort 验证 sort:xxx 标记的解析（用于“每日最新”等定时任务）。
func TestParseQuerySort(t *testing.T) {
	cases := []struct {
		in       string
		wantQ    string
		wantSort string
	}{
		{"agent skills sort:updated", "agent skills", "updated"},
		{"claude skills sort:created", "claude skills", "created"},
		{"mcp server sort:stars", "mcp server", "stars"},
		{"agent skills", "agent skills", ""},
		{"skills in:name sort:updated extra", "skills in:name extra", "updated"},
	}
	for _, c := range cases {
		q, s := parseQuerySort(c.in)
		if q != c.wantQ || s != c.wantSort {
			t.Fatalf("parseQuerySort(%q) = (%q, %q)，期望 (%q, %q)", c.in, q, s, c.wantQ, c.wantSort)
		}
	}
}

// TestSearchReposPaginated 验证分页翻取直到凑够 limit，且不请求多余页。
func TestSearchReposPaginated(t *testing.T) {
	var pages atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/search/repositories") {
			http.NotFound(w, r)
			return
		}
		page := 1
		if v := r.URL.Query().Get("page"); v != "" {
			_ = json.Unmarshal([]byte(v), &page)
		}
		pages.Store(int32(page))

		items := make([]Repo, 0, 10)
		for i := 0; i < 10; i++ {
			items = append(items, Repo{
				FullName:      "org/repo-" + strconv.Itoa(int(page)) + "-" + strconv.Itoa(i),
				DefaultBranch: "main",
				HTMLURL:       "https://github.com/org/repo",
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	}))
	defer srv.Close()

	old := apiBase
	apiBase = srv.URL
	defer func() { apiBase = old }()

	c := NewClient("")
	got, err := c.searchReposPaginated("skills", 10, 25)
	if err != nil {
		t.Fatalf("searchReposPaginated 出错: %v", err)
	}
	if len(got) != 25 {
		t.Fatalf("limit=25 应返回 25 个仓库，实际 %d", len(got))
	}
	if pages.Load() != 3 {
		t.Fatalf("应翻到第 3 页（25/10 向上取整）即停止，实际请求到第 %d 页", pages.Load())
	}

	// limit 小于一页时，只请求第 1 页并截断
	pages.Store(0)
	got, err = c.searchReposPaginated("skills", 10, 5)
	if err != nil {
		t.Fatalf("searchReposPaginated 出错: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("limit=5 应返回 5 个仓库，实际 %d", len(got))
	}
	if pages.Load() != 1 {
		t.Fatalf("limit=5 只应请求第 1 页，实际请求到第 %d 页", pages.Load())
	}
}

// TestBuildSkill_ContentFromSkillMD 集成验证：SKILL.md 有正文时，Content 来自 SKILL.md。
func TestBuildSkill_ContentFromSkillMD(t *testing.T) {
	skillMD := "---\nname: demo-skill\ndescription: 演示技能\n---\n# Demo\n## 用法\n执行命令"
	readme := "# Repo README\n仓库级介绍"

	c, srv := newMockClient(t, skillMD, readme)
	defer srv.Close()

	repo := Repo{FullName: "testorg/demo", DefaultBranch: "main", HTMLURL: "https://github.com/testorg/demo", Stars: 100}
	s, err := c.buildSkill(repo, "", "SKILL.md")
	if err != nil {
		t.Fatalf("buildSkill 出错: %v", err)
	}
	if len(s.Content) == 0 {
		t.Fatalf("SKILL.md 有正文时 Content 不应为空")
	}
	if s.Content[0].Heading != "用法" {
		t.Fatalf("正文应来自 SKILL.md（首个区块 %q），而不是被 README 覆盖", s.Content[0].Heading)
	}
}

// TestBuildSkill_ContentFallbackToReadme 集成验证：SKILL.md 无正文时回退 README。
func TestBuildSkill_ContentFallbackToReadme(t *testing.T) {
	skillMD := "---\nname: demo-skill\n---\n" // 只有 frontmatter，无正文
	readme := "# Repo README\n## 快速开始\n参考文档"

	c, srv := newMockClient(t, skillMD, readme)
	defer srv.Close()

	repo := Repo{FullName: "testorg/demo", DefaultBranch: "main", HTMLURL: "https://github.com/testorg/demo", Stars: 100}
	s, err := c.buildSkill(repo, "", "SKILL.md")
	if err != nil {
		t.Fatalf("buildSkill 出错: %v", err)
	}
	if len(s.Content) == 0 {
		t.Fatalf("SKILL.md 无正文、有 README 时 Content 应回退到 README")
	}
	if s.Content[0].Heading != "快速开始" {
		t.Fatalf("应回退使用 README 正文（首个区块 %q）", s.Content[0].Heading)
	}
}

// TestBuildSkill_ContentNoReadme 集成验证：无 README 时保持 SKILL.md 自身。
func TestBuildSkill_ContentNoReadme(t *testing.T) {
	skillMD := "---\nname: demo-skill\n---\n# Demo\n## 用法\n执行命令"
	c, srv := newMockClient(t, skillMD, "") // readme 返回 404
	defer srv.Close()

	repo := Repo{FullName: "testorg/demo", DefaultBranch: "main", HTMLURL: "https://github.com/testorg/demo", Stars: 100}
	s, err := c.buildSkill(repo, "", "SKILL.md")
	if err != nil {
		t.Fatalf("buildSkill 出错: %v", err)
	}
	if len(s.Content) == 0 || s.Content[0].Heading != "用法" {
		t.Fatalf("无 README 时应保持 SKILL.md 正文，实际 %+v", s.Content)
	}
}

// newMockClient 构造指向 mock GitHub API 的 Client。
// readme 为空串时 readme 端点返回 404。
func newMockClient(t *testing.T, skillMD, readme string) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/testorg/demo/contents/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		writeFileJSON(w, skillMD)
	})
	mux.HandleFunc("/repos/testorg/demo/readme", func(w http.ResponseWriter, _ *http.Request) {
		if readme == "" {
			http.NotFound(w, nil)
			return
		}
		writeFileJSON(w, readme)
	})
	srv := httptest.NewServer(mux)

	old := apiBase
	apiBase = srv.URL
	t.Cleanup(func() { apiBase = old })

	return NewClient(""), srv
}

// TestBuildSkill_ReferenceStyleFields 验证 buildSkill 生成与 mcpservers.org 一致的
// 展示字段：标题化名称、官方展示作者、安装命令、技能目录下载地址与 SkillPath。
func TestBuildSkill_ReferenceStyleFields(t *testing.T) {
	skillMD := "---\nname: frontend-design\ndescription: 独特的前端界面设计指导\n---\n# Frontend Design\n## 设计原则\n正文"
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/anthropics/skills/contents/skills/frontend-design/SKILL.md", func(w http.ResponseWriter, _ *http.Request) {
		writeFileJSON(w, skillMD)
	})
	mux.HandleFunc("/repos/anthropics/skills/readme/skills/frontend-design", func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	})
	srv := httptest.NewServer(mux)
	old := apiBase
	apiBase = srv.URL
	t.Cleanup(func() { apiBase = old })

	c := NewClient("")
	repo := Repo{
		FullName:      "anthropics/skills",
		DefaultBranch: "main",
		HTMLURL:       "https://github.com/anthropics/skills",
		Stars:         168700,
	}
	s, err := c.buildSkill(repo, "frontend-design", "skills/frontend-design/SKILL.md")
	if err != nil {
		t.Fatalf("buildSkill 出错: %v", err)
	}

	if s.Name != "Frontend Design" {
		t.Fatalf("展示名应为标题化 Frontend Design，实际 %q", s.Name)
	}
	if s.Author != "Anthropic" {
		t.Fatalf("作者应为官方展示名 Anthropic，实际 %q", s.Author)
	}
	if want := "npx skills add https://github.com/anthropics/skills --skill frontend-design"; s.InstallCommand != want {
		t.Fatalf("安装命令应为 %q，实际 %q", want, s.InstallCommand)
	}
	if want := "https://github.com/anthropics/skills/tree/main/skills/frontend-design"; s.DownloadURL != want {
		t.Fatalf("下载地址应为 %q，实际 %q", want, s.DownloadURL)
	}
	if s.SkillPath != "skills/frontend-design" {
		t.Fatalf("SkillPath 应为 skills/frontend-design，实际 %q", s.SkillPath)
	}
	if !s.IsOfficial {
		t.Fatalf("anthropics 组织应标记为官方")
	}
}

// TestDownloadSkillFolder 验证按技能目录拉取文件（git trees + blobs）。
func TestDownloadSkillFolder(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/anthropics/skills/git/trees/main", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"truncated": false,
			"tree": []map[string]any{
				{"path": "skills/frontend-design/SKILL.md", "type": "blob", "sha": "sha1", "size": 10},
				{"path": "skills/frontend-design/references/style.md", "type": "blob", "sha": "sha2", "size": 5},
				{"path": "skills/other/SKILL.md", "type": "blob", "sha": "sha3", "size": 10},
				{"path": "README.md", "type": "blob", "sha": "sha4", "size": 10},
			},
		})
	})
	mux.HandleFunc("/repos/anthropics/skills/git/blobs/sha1", func(w http.ResponseWriter, _ *http.Request) {
		writeFileJSON(w, "# 技能内容")
	})
	mux.HandleFunc("/repos/anthropics/skills/git/blobs/sha2", func(w http.ResponseWriter, _ *http.Request) {
		writeFileJSON(w, "参考")
	})
	srv := httptest.NewServer(mux)
	old := apiBase
	apiBase = srv.URL
	t.Cleanup(func() { apiBase = old })

	c := NewClient("")
	files, err := c.DownloadSkillFolder("anthropics/skills", "main", "skills/frontend-design")
	if err != nil {
		t.Fatalf("DownloadSkillFolder 出错: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("应只包含技能目录下 2 个文件，实际 %d（%v）", len(files), files)
	}
	if string(files["SKILL.md"]) != "# 技能内容" {
		t.Fatalf("SKILL.md 内容错误: %q", files["SKILL.md"])
	}
	if string(files["references/style.md"]) != "参考" {
		t.Fatalf("子目录文件内容错误: %q", files["references/style.md"])
	}
	if _, ok := files["skills/other/SKILL.md"]; ok {
		t.Fatalf("不应包含技能目录外的文件")
	}
	if _, ok := files["README.md"]; ok {
		t.Fatalf("不应包含技能目录外的文件")
	}
}

func writeFileJSON(w http.ResponseWriter, content string) {
	_ = json.NewEncoder(w).Encode(map[string]string{
		"content":  base64.StdEncoding.EncodeToString([]byte(content)),
		"encoding": "base64",
	})
}
