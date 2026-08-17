// Agent Skills API 服务启动入口。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/admin"
	"github.com/Interesting-workstations/skill-hub/server/internal/config"
	"github.com/Interesting-workstations/skill-hub/server/internal/router"
	"github.com/Interesting-workstations/skill-hub/server/internal/skill"
)

func main() {
	// 加载环境配置（本地 .env / 线上 .env.prod，由 SKILLHUB_ENV 区分），不覆盖已存在的环境变量
	if err := config.LoadEnv(config.EnvFile()); err != nil {
		log.Printf("⚠️ 加载 %s 失败: %v", config.EnvFile(), err)
	}

	addr := getenv("SERVER_ADDR", ":8080")
	dsn := getenv("MYSQL_DSN",
		"root:root@tcp(127.0.0.1:3306)/skillhub?charset=utf8mb4&parseTime=true")
	seedPath := getenv("SEED_PATH", "data/skills.json")

	// 数据层（MySQL Repository，首次启动自动建库并从种子 JSON 初始化）
	repo, err := skill.NewMySQLRepository(dsn, seedPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	svc := skill.NewService(repo)

	// 管理后台数据层（MySQL，独立建表 + 种子）
	adminRepo, err := admin.NewMySQLRepository(dsn)
	if err != nil {
		log.Fatalf("初始化后台数据库失败: %v", err)
	}
	adminSvc := admin.NewService(adminRepo)

	// 清理上次进程退出/崩溃残留的 running 执行记录与任务，避免永久卡在“运行中”
	if err := adminSvc.RecoverStaleExecutions(); err != nil {
		log.Printf("⚠️ 清理残留执行记录失败: %v", err)
	} else {
		log.Println("✅ 已清理残留的 running 执行记录/任务")
	}

	// 启动定时调度器：按任务 schedule 字段自动触发（每天 HH:MM / 每 N 小时 / 每小时）。
	// 本地连线上库调试时设置 SKILLHUB_DISABLE_SCHEDULER=1 可禁用，避免与线上 server 双跑爬虫。
	if os.Getenv("SKILLHUB_DISABLE_SCHEDULER") == "1" {
		log.Println("⏸️ 定时调度器已禁用（SKILLHUB_DISABLE_SCHEDULER=1）")
	} else {
		adminSvc.StartScheduler(context.Background())
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(svc, adminSvc),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Agent Skills API 已启动，监听 %s（数据库 MySQL）", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务…")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务关闭异常: %v", err)
	}
	log.Println("服务已退出")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
