// 存量数据翻译命令：把 skills 表里未翻译（name_zh 为空）的记录
// 的标题（name）与描述（description）翻译成中文写入 name_zh / description_zh。
//
// 用法：go run ./cmd/translate
// 可选参数：-limit N（最多处理 N 条，默认全部）；-concurrency N（并发数，默认 4）
package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/config"
	"github.com/Interesting-workstations/skill-hub/server/internal/translate"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	limit := flag.Int("limit", 0, "最多处理 N 条（0=全部）")
	conc := flag.Int("concurrency", 4, "并发翻译数")
	flag.Parse()

	_ = config.LoadEnv(".env")
	dsn := getenv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库不可用: %v", err)
	}

	tr := translate.New()
	if !tr.Enabled() {
		log.Fatalf("翻译未启用：TRANSLATE_PROVIDER=%s（google 无需 Key，baidu/deepl 需配置 Key）", os.Getenv("TRANSLATE_PROVIDER"))
	}

	// 查询待翻译记录
	rows, err := db.Query(`SELECT id, name, description FROM skills
		WHERE name_zh = '' OR description_zh IS NULL OR description_zh = ''`)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}
	type item struct {
		id          string
		name        string
		description string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.name, &it.description); err != nil {
			continue
		}
		items = append(items, it)
		if *limit > 0 && len(items) >= *limit {
			break
		}
	}
	rows.Close()
	log.Printf("待翻译记录: %d 条", len(items))

	if len(items) == 0 {
		log.Println("没有需要翻译的记录，退出")
		return
	}

	// 并发翻译 + 落库
	start := time.Now()
	jobs := make(chan item)
	var wg sync.WaitGroup
	var mu sync.Mutex
	done, failed := 0, 0

	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range jobs {
				nameZh := tr.Translate(it.name, "zh-CN")
				descZh := tr.Translate(it.description, "zh-CN")
				if _, err := db.Exec(`UPDATE skills SET name_zh = ?, description_zh = ? WHERE id = ?`,
					nameZh, descZh, it.id); err != nil {
					mu.Lock()
					failed++
					mu.Unlock()
					continue
				}
				mu.Lock()
				done++
				if done%50 == 0 {
					log.Printf("已翻译 %d/%d（耗时 %s）", done, len(items), time.Since(start).Round(time.Second))
				}
				mu.Unlock()
			}
		}()
	}
	for _, it := range items {
		jobs <- it
	}
	close(jobs)
	wg.Wait()

	log.Printf("完成：成功 %d 条，失败 %d 条，总耗时 %s", done, failed, time.Since(start).Round(time.Second))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
