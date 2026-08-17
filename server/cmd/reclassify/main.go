package main

// 技能分类重算命令：用最新分类规则（crawler.InferCategory）重新推断
// skills 表中所有存量技能的 category，并更新变化的记录。
// 用法：go -C server run ./cmd/reclassify

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Interesting-workstations/skill-hub/server/internal/config"
	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	_ = config.LoadEnv(config.EnvFile())
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN 未设置（检查 .env）")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, name, description, category FROM skills`)
	if err != nil {
		log.Fatalf("查询技能失败: %v", err)
	}
	defer rows.Close()

	var total, changed int
	changes := make(map[string]int) // 旧分类 -> 新分类 计数
	for rows.Next() {
		var id, name, desc, oldCat string
		if err := rows.Scan(&id, &name, &desc, &oldCat); err != nil {
			continue
		}
		total++
		newCat := crawler.InferCategory(name, desc)
		if newCat == oldCat {
			continue
		}
		if _, err := db.Exec(`UPDATE skills SET category = ? WHERE id = ?`, newCat, id); err != nil {
			log.Printf("更新 %s 失败: %v", id, err)
			continue
		}
		changed++
		changes[oldCat+" → "+newCat]++
	}

	fmt.Printf("共 %d 条技能，重新分类 %d 条\n", total, changed)
	if changed > 0 {
		fmt.Println("变更明细（旧 → 新）:")
		for k, v := range changes {
			fmt.Printf("  %-30s %d\n", k, v)
		}
	}
}
