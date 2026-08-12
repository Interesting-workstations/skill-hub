package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
)

func main() {
	query := flag.String("query", "agent skills", "GitHub 搜索关键词（如 \"claude skills\"、\"codex skills\"）")
	limit := flag.Int("limit", 20, "最多处理的仓库数量")
	perPage := flag.Int("per-page", 10, "每次搜索 API 请求返回的仓库数（最大 100）")
	repos := flag.String("repos", "", "额外指定要爬取的仓库，逗号分隔（如 \"anthropics/skills,openai/codex\"）")
	output := flag.String("output", "", "输出 JSON 文件路径（默认输出到 stdout）")
	flag.Parse()

	// 认证：优先 GITHUB_TOKEN 环境变量；未提供则以匿名方式请求（速率受限）。
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("⚠️ 未设置 GITHUB_TOKEN，使用匿名请求（GitHub API 速率限制 60 次/小时）")
	}

	client := crawler.NewClient(token)
	opts := crawler.CrawlOptions{
		Query:   *query,
		Repos:   splitRepos(*repos),
		Limit:   *limit,
		PerPage: *perPage,
	}

	fmt.Printf("🔍 开始爬取：关键词=%q，仓库上限=%d，指定仓库=%v\n", *query, *limit, opts.Repos)
	skills, err := client.Crawl(opts)
	if err != nil {
		log.Fatalf("爬取失败: %v", err)
	}

	data, err := json.MarshalIndent(skills, "", "  ")
	if err != nil {
		log.Fatalf("序列化失败: %v", err)
	}

	if *output != "" {
		if err := os.WriteFile(*output, data, 0o644); err != nil {
			log.Fatalf("写入文件失败: %v", err)
		}
		fmt.Printf("✅ 共爬取 %d 个技能，已写入 %s\n", len(skills), *output)
		return
	}
	fmt.Printf("✅ 共爬取 %d 个技能：\n", len(skills))
	_, _ = os.Stdout.Write(data)
}

// splitRepos 将逗号分隔的仓库列表解析为切片。
func splitRepos(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
