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

func TestParseContent_Sections(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
---
# My Skill

这是简介段落。

## 使用方法

先安装，再使用。

- 步骤一
- 步骤二

## 注意事项

请小心。
`
	sections := ParseContent(content)
	if len(sections) != 3 {
		t.Fatalf("期望 3 个内容区块，实际 %d", len(sections))
	}
	if sections[0].Heading != "概述" || len(sections[0].Body) != 1 || sections[0].Body[0] != "这是简介段落。" {
		t.Fatalf("概述区块解析错误: %+v", sections[0])
	}
	if sections[1].Heading != "使用方法" || len(sections[1].Body) != 3 {
		t.Fatalf("使用方法区块解析错误: %+v", sections[1])
	}
	if sections[2].Heading != "注意事项" {
		t.Fatalf("注意事项区块解析错误: %+v", sections[2])
	}
}

func TestParseContent_NoHeading(t *testing.T) {
	content := `# Tool
Some intro line.
Another line.`
	sections := ParseContent(content)
	if len(sections) != 1 {
		t.Fatalf("期望 1 个区块，实际 %d", len(sections))
	}
	if sections[0].Heading != "概述" || len(sections[0].Body) != 2 {
		t.Fatalf("无标题时应归入概述区块: %+v", sections[0])
	}
}
