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
		if got := inferCategory(c.name, c.desc); got != c.want {
			t.Fatalf("inferCategory(%q, %q) = %q，期望 %q", c.name, c.desc, got, c.want)
		}
	}
}
