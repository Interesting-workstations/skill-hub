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
