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

	"github.com/Interesting-workstations/skill-hub/server/internal/router"
	"github.com/Interesting-workstations/skill-hub/server/internal/skill"
)

func main() {
	addr := getenv("SERVER_ADDR", ":8080")
	dataPath := getenv("DATA_PATH", "data/skills.json")

	// 数据层（内存 Repository，从 JSON 种子加载）
	repo, err := skill.NewMemoryRepository(dataPath)
	if err != nil {
		log.Fatalf("加载种子数据失败: %v", err)
	}
	svc := skill.NewService(repo)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router.New(svc),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Agent Skills API 已启动，监听 %s（数据源 %s）", addr, dataPath)
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
