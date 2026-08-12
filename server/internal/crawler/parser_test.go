package crawler

import (
	"strings"
	"testing"
)

func TestParseSkillMD_Frontmatter(t *testing.T) {
	content := []byte(`---
name: my-skill
description: A test skill for agent
---
# My Skill
Body here`)
	name, desc := ParseSkillMD(content)
	if name != "my-skill" {
		t.Fatalf("期望 name=my-skill，实际 %q", name)
	}
	if desc != "A test skill for agent" {
		t.Fatalf("期望 desc 含 frontmatter 描述，实际 %q", desc)
	}
}

func TestParseSkillMD_NoFrontmatter(t *testing.T) {
	content := []byte("# My Skill\n\nSome description text here.")
	name, desc := ParseSkillMD(content)
	if name != "My Skill" {
		t.Fatalf("期望从标题推断 name=My Skill，实际 %q", name)
	}
	if !strings.Contains(desc, "Some description") {
		t.Fatalf("期望从正文推断描述，实际 %q", desc)
	}
}

func TestParseSkillMD_Empty(t *testing.T) {
	name, desc := ParseSkillMD([]byte(""))
	if name != "" || desc != "" {
		t.Fatalf("空内容应返回空，实际 name=%q desc=%q", name, desc)
	}
}

func TestFormatStars(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{239, "239"},
		{1200, "1.2k"},
		{239600, "239.6k"},
	}
	for _, c := range cases {
		if got := formatStars(c.in); got != c.want {
			t.Fatalf("formatStars(%d) = %q，期望 %q", c.in, got, c.want)
		}
	}
}
