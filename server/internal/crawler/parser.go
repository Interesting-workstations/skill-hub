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
// 其余行追加到当前区块正文（保留原始内容：前导空格缩进与空行，避免代码块、
// ASCII 图在渲染时丢失结构）；主标题后的简介段落归入「概述」。
func ParseContent(text string) []ContentSection {
	text = stripFrontmatter(text)
	var sections []ContentSection
	current := -1

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		trimmed := strings.TrimSpace(line)

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
		// 正文行保留原始内容（含前导/行内空格与空行）
		sections[current].Body = append(sections[current].Body, line)
	}

	// 丢弃只有标题、正文为空（仅空行）的区块
	var out []ContentSection
	for _, s := range sections {
		hasText := false
		for _, l := range s.Body {
			if strings.TrimSpace(l) != "" {
				hasText = true
				break
			}
		}
		if !hasText {
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

// commonAcronyms 展示时保持全大写的常见缩写（文件格式 / 技术术语）。
// 用于把 skill slug（如 xlsx、pdf-chat）转成站点展示标题（XLSX、PDF Chat）。
var commonAcronyms = map[string]bool{
	"pdf": true, "xlsx": true, "docx": true, "csv": true, "tsv": true,
	"html": true, "css": true, "xml": true, "json": true,
	"api": true, "ai": true, "ui": true, "ux": true, "svg": true,
	"png": true, "jpeg": true, "jpg": true, "gif": true, "sql": true,
	"mcp": true, "cli": true, "sdk": true, "aws": true, "gcp": true,
	"http": true, "https": true, "url": true, "uri": true,
	"js": true, "ts": true, "md": true, "seo": true,
	"gpt": true, "llm": true, "rag": true, "cot": true,
	"ppt": true, "doc": true, "zip": true, "rest": true,
	"ide": true, "qa": true, "e2e": true, "git": true,
	"cpu": true, "gpu": true, "db": true, "3d": true, "2d": true,
}

// TitleCase 将技能 slug（如 frontend-design）转为展示标题（Frontend Design）。
// 按 -、_、空格、点分词，每个全小写词首字母大写；常见缩写（commonAcronyms）保持全大写；
// 已含大写字母的词（如 NotebookLM、PDF）保持不变。对已格式化名称幂等。
func TitleCase(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' ' || r == '.'
	})
	for i, w := range fields {
		if w == "" {
			continue
		}
		lw := strings.ToLower(w)
		if commonAcronyms[lw] {
			fields[i] = strings.ToUpper(lw)
			continue
		}
		if isAllUpper(w) && len([]rune(w)) > 1 {
			continue // 已全大写（如传入 PDF）
		}
		if strings.ToLower(w) != w {
			continue // 已含大写（如 NotebookLM），保持原样
		}
		// 全小写词：仅首字母大写（如 frontend → Frontend）
		fields[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(fields, " ")
}

// isAllUpper 判断字符串是否全部为大写字母/数字（忽略符号）。
func isAllUpper(s string) bool {
	hasLetter := false
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			hasLetter = true
		}
	}
	return hasLetter
}
