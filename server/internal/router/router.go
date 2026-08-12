// Package router 组装路由与中间件。
package router

import (
	"net/http"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/middleware"
	"github.com/Interesting-workstations/skill-hub/server/internal/skill"
)

// New 构建应用路由：RequestID → Recovery → Logger → CORS → 业务路由。
func New(svc *skill.Service) http.Handler {
	mux := http.NewServeMux()
	skill.NewHandler(svc).Register(mux)

	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)
	h = middleware.Logger(h)
	h = middleware.CORS(h)
	return h
}

// Server 持有 HTTP 服务配置。
type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultServer 返回默认服务配置。
func DefaultServer() Server {
	return Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}
