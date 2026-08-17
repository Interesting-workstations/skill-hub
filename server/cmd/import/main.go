// Command import 将爬虫输出的技能 JSON 全量导入 MySQL。
// 官方组织的技能会被标记为官方，并按分类组织。
//
// 用法：
//
//	# 先运行爬虫生成 JSON
//	go run ./cmd/crawler -query "claude skills" -limit 30 -output data/crawled-skills.json
//	# 再导入 MySQL（全量替换）
//	go run ./cmd/import -input data/crawled-skills.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/config"
	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/Interesting-workstations/skill-hub/server/internal/skill"
)

func main() {
	// 加载 .env（MYSQL_DSN 等）
	if err := config.LoadEnv(config.EnvFile()); err != nil {
		log.Printf("⚠️ 加载 .env 失败: %v", err)
	}

	input := flag.String("input", "", "爬虫输出 JSON 文件路径（必填）")
	dsn := flag.String("dsn",
		"root:root@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true",
		"MySQL 连接串")
	flag.Parse()

	if *input == "" {
		log.Fatal("请指定 -input 爬虫输出 JSON 路径")
	}
	data, err := os.ReadFile(*input)
	if err != nil {
		log.Fatalf("读取爬取数据失败: %v", err)
	}
	var skills []crawler.Skill
	if err := json.Unmarshal(data, &skills); err != nil {
		log.Fatalf("解析爬取数据失败: %v", err)
	}
	if len(skills) == 0 {
		log.Fatal("爬取数据为空，请先运行爬虫")
	}

	store := toStore(skills)
	official := 0
	for _, s := range skills {
		if s.IsOfficial {
			official++
		}
	}

	fmt.Printf("📦 开始导入：%d 个技能（官方 %d 个）、%d 个作者、%d 个分类\n",
		len(skills), official, len(store.Authors), len(store.SkillCategories))
	if err := skill.ReplaceAll(*dsn, store); err != nil {
		log.Fatalf("导入 MySQL 失败: %v", err)
	}
	fmt.Printf("✅ 已导入 MySQL（%s），首页与官方技能区块将展示爬取的真实数据\n", *dsn)
}

// toStore 将爬取到的扁平技能列表转换为站点数据模型（作者/分类聚合）。
func toStore(skills []crawler.Skill) domain.Store {
	authorMap := make(map[string]*domain.Author)
	for _, s := range skills {
		key := strings.ToLower(s.Author)
		if a, ok := authorMap[key]; ok {
			a.SkillCount++
			continue
		}
		avatar := crawler.OfficialAvatar(s.Author)
		if avatar == "" {
			avatar = "📦"
		}
		authorMap[key] = &domain.Author{
			Name:       s.Author,
			Slug:       crawler.Slugify(s.Author),
			Avatar:     avatar,
			SkillCount: 1,
		}
	}

	catMap := make(map[string]*domain.Category)
	for _, s := range skills {
		cat, ok := catMap[s.Category]
		if !ok {
			cat = &domain.Category{Slug: s.Category, Name: categoryName(s.Category)}
			catMap[s.Category] = cat
		}
		cat.Skills = append(cat.Skills, toDomainSkill(s))
	}

	authors := make([]domain.Author, 0, len(authorMap))
	for _, a := range authorMap {
		authors = append(authors, *a)
	}
	sort.Slice(authors, func(i, j int) bool { return authors[i].Slug < authors[j].Slug })

	categories := make([]domain.Category, 0, len(catMap))
	for _, c := range catMap {
		categories = append(categories, *c)
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i].Slug < categories[j].Slug })

	return domain.Store{
		Authors:         authors,
		SkillCategories: categories,
	}
}

// toDomainSkill 将爬虫技能转换为领域技能。
func toDomainSkill(s crawler.Skill) domain.Skill {
	content := make([]domain.ContentSection, 0, len(s.Content))
	for _, sec := range s.Content {
		content = append(content, domain.ContentSection{Heading: sec.Heading, Body: sec.Body})
	}
	return domain.Skill{
		ID:             s.ID,
		Name:           s.Name,
		Author:         s.Author,
		Description:    s.Description,
		Tags:           s.Tags,
		Category:       s.Category,
		DownloadURL:    s.DownloadURL,
		IsOfficial:     s.IsOfficial,
		IsFeatured:     s.IsFeatured,
		InstallCommand: s.InstallCommand,
		GithubURL:      s.GithubURL,
		GithubStars:    s.GithubStars,
		License:        s.License,
		SkillPath:      s.SkillPath,
		Content:        content,
	}
}

// categoryNames 分类 slug 的可读名称。
var categoryNames = map[string]string{
	"development":        "开发技能",
	"testing":            "测试技能",
	"document":           "文档处理",
	"browser-automation": "浏览器自动化",
	"database":           "数据库",
	"creative":           "创意设计",
	"media":              "媒体处理",
	"productivity":       "效率工具",
	"security":           "安全",
}

func categoryName(slug string) string {
	if name, ok := categoryNames[slug]; ok {
		return name
	}
	return slug
}
