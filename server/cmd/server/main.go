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
	"github.com/Interesting-workstations/skill-hub/server/internal/router"
	"github.com/Interesting-workstations/skill-hub/server/internal/skill"
)

func main() {
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
