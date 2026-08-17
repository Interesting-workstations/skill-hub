package crawler

import "testing"

func TestIsOfficialOrg(t *testing.T) {
	cases := []struct {
		owner string
		want  bool
	}{
		{"anthropics", true},
		{"openai", true},
		{"microsoft", true},
		{"vercel", true},
		{"google", true},
		{"Anthropics", true}, // 大小写不敏感
		{"github", true},
		{"addyosmani", false},
		{"pleaseprompto", false},
		{"", false},
	}
	for _, c := range cases {
		if got := IsOfficialOrg(c.owner); got != c.want {
			t.Fatalf("IsOfficialOrg(%q) = %v，期望 %v", c.owner, got, c.want)
		}
	}
}

func TestOfficialAvatar(t *testing.T) {
	if got := OfficialAvatar("anthropics"); got != "🅰️" {
		t.Fatalf("OfficialAvatar(anthropics) = %q", got)
	}
	if got := OfficialAvatar("random-org"); got != "" {
		t.Fatalf("非官方组织应返回空串，实际 %q", got)
	}
	// 展示名也应能查到头像（与 GitHub owner 一致）
	if got := OfficialAvatar("Anthropic"); got != "🅰️" {
		t.Fatalf("OfficialAvatar(Anthropic) = %q", got)
	}
	if got := OfficialAvatar("Hugging Face"); got != "🤗" {
		t.Fatalf("OfficialAvatar(Hugging Face) = %q", got)
	}
}

func TestOfficialDisplayName(t *testing.T) {
	cases := []struct {
		owner, want string
	}{
		{"anthropics", "Anthropic"},
		{"openai", "OpenAI"},
		{"microsoft", "Microsoft"},
		{"github", "GitHub"},
		{"huggingface", "Hugging Face"},
		{"Anthropics", "Anthropic"},  // 大小写不敏感
		{"addyosmani", "addyosmani"}, // 非官方保留用户名
		{"", ""},
	}
	for _, c := range cases {
		if got := OfficialDisplayName(c.owner); got != c.want {
			t.Fatalf("OfficialDisplayName(%q) = %q，期望 %q", c.owner, got, c.want)
		}
	}
}

func TestInstallCommand(t *testing.T) {
	repoURL := "https://github.com/anthropics/skills"
	if got := installCommand(repoURL, "skills/frontend-design"); got != "npx skills add https://github.com/anthropics/skills --skill frontend-design" {
		t.Fatalf("目录技能安装命令错误: %q", got)
	}
	if got := installCommand(repoURL, ""); got != "npx skills add https://github.com/anthropics/skills" {
		t.Fatalf("仓库根技能安装命令错误: %q", got)
	}
}

func TestSkillDownloadURL(t *testing.T) {
	repo := Repo{
		FullName:      "anthropics/skills",
		HTMLURL:       "https://github.com/anthropics/skills",
		DefaultBranch: "main",
	}
	if got := skillDownloadURL(repo, "skills/frontend-design"); got != "https://github.com/anthropics/skills/tree/main/skills/frontend-design" {
		t.Fatalf("目录技能下载地址错误: %q", got)
	}
	if got := skillDownloadURL(repo, ""); got != "https://github.com/anthropics/skills" {
		t.Fatalf("仓库根技能下载地址错误: %q", got)
	}
}

func TestInferCategory(t *testing.T) {
	cases := []struct {
		name, desc, want string
	}{
		{"Playwright Automation", "Automate browser testing with Playwright", "browser-automation"},
		{"SQL Query Helper", "Write and optimize SQL queries for Postgres", "database"},
		{"PDF Summarizer", "Extract and summarize PDF documents", "document"},
		{"Video Transcriber", "Transcribe video and audio content", "media"},
		{"UI Designer", "Generate Figma design mockups", "creative"},
		{"Meeting Notes", "Auto-generate meeting notes and action items", "productivity"},
		{"E2E Test Runner", "Run end-to-end tests across browsers", "testing"},
		{"Security Audit", "Audit authentication and encryption", "security"},
		{"Code Refactor", "Refactor Go code with best practices", "development"},
		{"Generic Tool", "A random utility with no hints", "development"},
	}
	for _, c := range cases {
		if got := InferCategory(c.name, c.desc); got != c.want {
			t.Fatalf("InferCategory(%q, %q) = %q，期望 %q", c.name, c.desc, got, c.want)
		}
	}
}
