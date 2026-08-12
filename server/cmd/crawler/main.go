package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
)

func main() {
	query := flag.String("query", "agent skills", "GitHub 搜索关键词（如 \"claude skills\"、\"codex skills\"）")
	limit := flag.Int("limit", 20, "最多处理的仓库数量")
	perPage := flag.Int("per-page", 10, "每次搜索 API 请求返回的仓库数（最大 100）")
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
		Limit:   *limit,
		PerPage: *perPage,
	}

	fmt.Printf("🔍 开始爬取：关键词=%q，仓库上限=%d\n", *query, *limit)
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
