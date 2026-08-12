package crawler

import (
	"regexp"
	"strings"
)

// 匹配 SKILL.md 顶部的 YAML frontmatter：--- ... ---
var frontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)

// ParseSkillMD 从 SKILL.md 内容中提取技能 name 与 description。
// 优先读取 frontmatter 中的 name/description 字段；
// 缺失时从首个一级标题或首行文本推断。
func ParseSkillMD(content []byte) (name, description string) {
	text := string(content)

	if m := frontmatterRe.FindStringSubmatch(text); len(m) > 1 {
		for _, line := range strings.Split(m[1], "\n") {
			kv := strings.SplitN(line, ":", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			val := strings.Trim(strings.TrimSpace(kv[1]), `"'`)
			switch strings.ToLower(key) {
			case "name":
				name = val
			case "description":
				description = val
			}
		}
	}

	if name == "" {
		name = inferName(text)
	}
	if description == "" {
		description = inferDescription(text)
	}
	return name, description
}

// inferName 从第一个 # 标题或首个非空文本行推断名称。
func inferName(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "---") {
			return line
		}
	}
	return ""
}

// inferDescription 取正文中第一个非标题段落作为描述。
func inferDescription(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "---") {
			continue
		}
		return truncate(line, 200)
	}
	return ""
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// ParseContent 从 SKILL.md 正文解析为内容区块。
// 规则：`# ` 主标题跳过（作为 name）；`## / ### ` 开始新内容区块；
// 其余非空行追加到当前区块正文；主标题后的简介段落归入「概述」。
func ParseContent(text string) []ContentSection {
	text = stripFrontmatter(text)
	var sections []ContentSection
	current := -1

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "##") || strings.HasPrefix(trimmed, "###") {
			heading := strings.TrimSpace(strings.TrimLeft(trimmed, "# "))
			sections = append(sections, ContentSection{Heading: heading})
			current = len(sections) - 1
			continue
		}
		if strings.HasPrefix(trimmed, "# ") {
			continue // 主标题 = name
		}
		if current < 0 {
			sections = append(sections, ContentSection{Heading: "概述"})
			current = len(sections) - 1
		}
		sections[current].Body = append(sections[current].Body, trimmed)
	}

	// 丢弃只有标题没有正文的区块
	var out []ContentSection
	for _, s := range sections {
		if len(s.Body) == 0 {
			continue
		}
		out = append(out, s)
	}
	return out
}

// stripFrontmatter 移除文档开头的 YAML frontmatter（--- ... ---）。
func stripFrontmatter(text string) string {
	return frontmatterRe.ReplaceAllString(text, "")
}
