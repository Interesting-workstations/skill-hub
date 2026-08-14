// Package middleware 提供 HTTP 中间件：请求 ID、异常恢复、访问日志、CORS。
package middleware

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

type ctxKey string

const reqIDKey ctxKey = "request_id"

func contextWithReqID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, reqIDKey, id)
}

func reqIDFromCtx(ctx context.Context) string {
	if id, ok := ctx.Value(reqIDKey).(string); ok {
		return id
	}
	return "-"
}

// RequestID 为每个请求生成唯一 ID，注入 Context 并回写响应头。
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := contextWithReqID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Recovery 捕获 handler panic，避免进程崩溃并返回 500。
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] request_id=%s err=%v", reqIDFromCtx(r.Context()), rec)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"code":50001,"message":"系统繁忙，请稍后重试","data":null}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Logger 记录请求日志：时间 / 请求 ID / 方法 / 路径 / 状态码 / 耗时。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Hijack 实现 http.Hijacker：WebSocket 升级需要底层连接可被劫持。
// 若不转发，Logger 包装后的 ResponseWriter 将无法支持 WS 握手。
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter 不支持 Hijack")
	}
	return hj.Hijack()
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("[access] request_id=%s method=%s path=%s status=%d duration=%s",
			reqIDFromCtx(r.Context()), r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Microsecond))
	})
}

// CORS 允许前端跨域访问（开发环境 5173）。
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
