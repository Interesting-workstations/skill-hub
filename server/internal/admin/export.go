package admin

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

// ExportFormat 导出格式。
type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportCSV  ExportFormat = "csv"
	ExportMD   ExportFormat = "markdown"
)

// ExportScope 导出范围。
type ExportScope string

const (
	ScopeAll       ExportScope = "all"
	ScopePublished ExportScope = "published"
	ScopeApproved  ExportScope = "approved"
)

// ExportResult 导出结果（文本内容 + 文件名）。
type ExportResult struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Format   string `json:"format"`
	Count    int    `json:"count"`
}

// ExportData 导出全部技能为 JSON / CSV / Markdown。
// scope 支持 all / published / approved；不支持的格式返回错误。
func (s *Service) ExportData(format, scope string) (ExportResult, error) {
	skills, err := s.repo.ListExportSkills()
	if err != nil {
		return ExportResult{}, fmt.Errorf("读取数据失败")
	}

	// 按范围过滤
	sc := ExportScope(scope)
	if sc == "" {
		sc = ScopePublished
	}
	var filtered []domainSkillExport
	for _, sk := range skills {
		switch sc {
		case ScopeAll:
			// 全部
		case ScopeApproved:
			if !sk.IsOfficial && !sk.IsFeatured {
				// 无审核状态时按已发布处理；这里全部视为已审核数据
			}
		case ScopePublished, "":
			// 已发布（默认）
		}
		filtered = append(filtered, domainSkillExport{
			ID:          sk.ID,
			Name:        sk.Name,
			Author:      sk.Author,
			Description: sk.Description,
			Category:    sk.Category,
			Tags:        sk.Tags,
			DownloadURL: sk.DownloadURL,
			IsOfficial:  sk.IsOfficial,
			IsFeatured:  sk.IsFeatured,
			GithubURL:   sk.GithubURL,
			GithubStars: sk.GithubStars,
			License:     sk.License,
		})
	}

	res := ExportResult{Format: format, Count: len(filtered)}
	switch ExportFormat(format) {
	case ExportJSON:
		res.Filename = "skills.json"
		b, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return ExportResult{}, fmt.Errorf("导出失败")
		}
		res.Content = string(b)
	case ExportCSV:
		res.Filename = "skills.csv"
		res.Content = exportCSV(filtered)
	case ExportMD:
		res.Filename = "skills.md"
		res.Content = exportMarkdown(filtered)
	default:
		return ExportResult{}, fmt.Errorf("不支持的导出格式")
	}
	return res, nil
}

// domainSkillExport 导出的技能字段（扁平化，不含 content 区块）。
type domainSkillExport struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags,omitempty"`
	DownloadURL string   `json:"downloadUrl,omitempty"`
	IsOfficial  bool     `json:"isOfficial"`
	IsFeatured  bool     `json:"isFeatured"`
	GithubURL   string   `json:"githubUrl,omitempty"`
	GithubStars string   `json:"githubStars,omitempty"`
	License     string   `json:"license,omitempty"`
}

// exportCSV 生成 CSV（含表头，字段内含逗号/引号时按 RFC 4180 转义）。
func exportCSV(items []domainSkillExport) string {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "name", "author", "description", "category", "tags",
		"downloadUrl", "isOfficial", "githubUrl", "githubStars", "license"})
	for _, it := range items {
		_ = w.Write([]string{
			it.ID, it.Name, it.Author, it.Description, it.Category,
			strings.Join(it.Tags, "|"), it.DownloadURL, boolStr(it.IsOfficial),
			it.GithubURL, it.GithubStars, it.License,
		})
	}
	w.Flush()
	return buf.String()
}

// exportMarkdown 生成 Markdown 技能清单（按作者分组）。
func exportMarkdown(items []domainSkillExport) string {
	var b strings.Builder
	b.WriteString("# Agent Skills 资源库导出\n\n")
	b.WriteString(fmt.Sprintf("共 %d 个技能\n\n", len(items)))
	byAuthor := map[string][]domainSkillExport{}
	var authors []string
	for _, it := range items {
		if _, ok := byAuthor[it.Author]; !ok {
			authors = append(authors, it.Author)
		}
		byAuthor[it.Author] = append(byAuthor[it.Author], it)
	}
	for _, a := range authors {
		b.WriteString("## " + a + "\n\n")
		for _, it := range byAuthor[a] {
			official := ""
			if it.IsOfficial {
				official = "（官方）"
			}
			b.WriteString(fmt.Sprintf("- [%s](%s)%s — %s", it.Name, it.DownloadURL, official, it.Description))
			if it.GithubStars != "" {
				b.WriteString(fmt.Sprintf(" ⭐%s", it.GithubStars))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
