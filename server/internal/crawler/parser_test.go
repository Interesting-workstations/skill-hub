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

func TestParseSkillMD_BlockScalar(t *testing.T) {
	// folded 块标量 >-
	content := []byte(`---
name: claude-api
description: >-
  This skill helps you build LLM-powered
  applications with Claude. Choose the right
  surface based on your needs.
---
# Claude API`)
	name, desc := ParseSkillMD(content)
	if name != "claude-api" {
		t.Fatalf("期望 name=claude-api，实际 %q", name)
	}
	want := "This skill helps you build LLM-powered applications with Claude. Choose the right surface based on your needs."
	if desc != want {
		t.Fatalf("期望 desc=%q，实际 %q", want, desc)
	}
}

func TestParseSkillMD_LiteralBlockScalar(t *testing.T) {
	// literal 块标量 |-
	content := []byte(`---
name: pptx
description: |-
  Use this skill any time a .pptx or .potx
  file is involved in any way.
---
# Pptx`)
	_, desc := ParseSkillMD(content)
	want := "Use this skill any time a .pptx or .potx file is involved in any way."
	if desc != want {
		t.Fatalf("期望 desc=%q，实际 %q", want, desc)
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
	// 概述：主标题后的简介段落（前后空行保留）
	if sections[0].Heading != "概述" {
		t.Fatalf("概述区块解析错误: %+v", sections[0])
	}
	if got := strings.Join(sections[0].Body, "\n"); !strings.Contains(got, "这是简介段落。") {
		t.Fatalf("概述区块应含简介段落: %q", got)
	}
	// 使用方法：含列表与保留的空行
	if sections[1].Heading != "使用方法" {
		t.Fatalf("使用方法区块解析错误: %+v", sections[1])
	}
	if got := strings.Join(sections[1].Body, "\n"); !strings.Contains(got, "- 步骤一") || !strings.Contains(got, "- 步骤二") {
		t.Fatalf("使用方法区块应含列表项: %q", got)
	}
	if sections[2].Heading != "注意事项" {
		t.Fatalf("注意事项区块解析错误: %+v", sections[2])
	}
}

// TestParseContent_PreservesIndentAndBlankLines 验证代码块缩进与空行在解析后不被丢弃。
func TestParseContent_PreservesIndentAndBlankLines(t *testing.T) {
	content := "# Demo\n\n## 示例\n\n```json\n{\n  \"name\": \"demo\",\n  \"items\": [\n    \"a\",\n    \"b\"\n  ]\n}\n```\n\n## 下一节\n\n正文。\n"
	sections := ParseContent(content)
	var joined string
	for _, s := range sections {
		if s.Heading == "示例" {
			joined = strings.Join(s.Body, "\n")
		}
	}
	if joined == "" {
		t.Fatalf("未找到「示例」区块")
	}
	// 代码块顶层缩进必须保留
	if !strings.Contains(joined, "  \"name\": \"demo\"") {
		t.Fatalf("代码块前导空格（缩进）丢失:\n%q", joined)
	}
	// 嵌套缩进必须保留
	if !strings.Contains(joined, "    \"a\",") {
		t.Fatalf("代码块嵌套缩进丢失:\n%q", joined)
	}
	// 代码块内的空行必须保留
	if !strings.Contains(joined, "  \"items\": [\n    \"a\",") {
		t.Fatalf("代码块内空行/换行异常:\n%q", joined)
	}
	// 围栏行本身也应保留
	if !strings.Contains(joined, "```json") {
		t.Fatalf("代码块围栏行丢失:\n%q", joined)
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

func TestTitleCase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"frontend-design", "Frontend Design"},
		{"webapp-testing", "Webapp Testing"},
		{"meeting-notes", "Meeting Notes"},
		{"algorithmic-art", "Algorithmic Art"},
		{"brand-guidelines", "Brand Guidelines"},
		{"canvas-design", "Canvas Design"},
		{"internal-comms", "Internal Comms"},
		{"artifacts-builder", "Artifacts Builder"},
		{"docx", "DOCX"},
		{"xlsx", "XLSX"},
		{"pdf-chat", "PDF Chat"},
		{"task-automation", "Task Automation"},
		{"accessibility", "Accessibility"},
		{"e2e-testing", "E2E Testing"},
		{"NotebookLM Skill", "NotebookLM Skill"}, // 混合大小写保持不变
		{"Frontend Design", "Frontend Design"},   // 已格式化幂等
		{"3d-modeling", "3D Modeling"},
		{"hello_world", "Hello World"},
		{"a.b.c", "A B C"},
	}
	for _, c := range cases {
		if got := TitleCase(c.in); got != c.want {
			t.Fatalf("TitleCase(%q) = %q，期望 %q", c.in, got, c.want)
		}
	}
}
