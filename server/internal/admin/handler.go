// Package admin HTTP 层：参数解析、认证校验、调用 Service、统一响应。
package admin

import (
	"encoding/json"
	"net/http"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/Interesting-workstations/skill-hub/server/internal/response"
)

// Handler 后台管理 HTTP 处理器。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册后台管理路由。登录与健康检查公开，其余需要 Bearer Token。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/login", h.login)
	mux.HandleFunc("GET /api/v1/admin/health", h.health)

	h.protected(mux, "GET /api/v1/admin/stats", h.stats)

	// 爬虫任务
	h.protected(mux, "GET /api/v1/admin/tasks", h.listTasks)
	h.protected(mux, "POST /api/v1/admin/tasks", h.createTask)
	h.protected(mux, "PUT /api/v1/admin/tasks/{id}", h.updateTask)
	h.protected(mux, "DELETE /api/v1/admin/tasks/{id}", h.deleteTask)
	h.protected(mux, "POST /api/v1/admin/tasks/{id}/run", h.runTask)
	h.protected(mux, "POST /api/v1/admin/tasks/{id}/stop", h.stopTask)

	// 执行记录
	h.protected(mux, "GET /api/v1/admin/executions", h.listExecutions)
	h.protected(mux, "GET /api/v1/admin/executions/{id}", h.getExecution)

	// 失败任务
	h.protected(mux, "GET /api/v1/admin/failures", h.listFailures)
	h.protected(mux, "DELETE /api/v1/admin/failures/{id}", h.deleteFailure)

	// 爬虫配置
	h.protected(mux, "GET /api/v1/admin/config", h.getConfig)
	h.protected(mux, "PUT /api/v1/admin/config", h.saveConfig)

	// 抓取数据
	h.protected(mux, "GET /api/v1/admin/data", h.listData)
	h.protected(mux, "PUT /api/v1/admin/data/{id}/status", h.updateDataStatus)
	h.protected(mux, "DELETE /api/v1/admin/data/{id}", h.deleteData)

	// 文章
	h.protected(mux, "GET /api/v1/admin/articles", h.listArticles)
	h.protected(mux, "POST /api/v1/admin/articles", h.createArticle)
	h.protected(mux, "DELETE /api/v1/admin/articles/{id}", h.deleteArticle)

	// SEO / 站点
	h.protected(mux, "GET /api/v1/admin/seo", h.getSeo)
	h.protected(mux, "PUT /api/v1/admin/seo", h.saveSeo)
	h.protected(mux, "GET /api/v1/admin/site-config", h.getSiteConfig)
	h.protected(mux, "PUT /api/v1/admin/site-config", h.saveSiteConfig)

	// 管理员
	h.protected(mux, "PUT /api/v1/admin/password", h.changePassword)
}

// protected 为路由套上管理员认证。
func (h *Handler) protected(mux *http.ServeMux, pattern string, hfn http.HandlerFunc) {
	mux.HandleFunc(pattern, h.auth(hfn))
}

// auth 校验 Authorization: Bearer <token>。
func (h *Handler) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(header) < len(prefix) || header[:len(prefix)] != prefix {
			response.Fail(w, http.StatusUnauthorized, 40101, "未登录或登录已过期")
			return
		}
		if _, ok := h.svc.VerifyToken(header[len(prefix):]); !ok {
			response.Fail(w, http.StatusUnauthorized, 40101, "未登录或登录已过期")
			return
		}
		next(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, map[string]string{"status": "admin ok"})
}

// ---------- 认证 ----------

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	user, token, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(w, http.StatusUnauthorized, 40102, "用户名或密码错误")
		return
	}
	response.OK(w, map[string]any{"token": token, "user": user})
}

type pwdReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	username := authUser(h.svc, r)
	var req pwdReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewPassword == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.ChangePassword(username, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, map[string]string{"message": "密码已更新"})
}

// authUser 从 token 中取出用户名（认证中间件已保证通过）。
func authUser(svc *Service, r *http.Request) string {
	header := r.Header.Get("Authorization")
	if len(header) > 7 {
		u, _ := svc.VerifyToken(header[7:])
		return u
	}
	return ""
}

// ---------- 工作台 ----------

func (h *Handler) stats(w http.ResponseWriter, _ *http.Request) {
	st, err := h.svc.Stats()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, st)
}

// ---------- 爬虫任务 ----------

func (h *Handler) listTasks(w http.ResponseWriter, _ *http.Request) {
	tasks, err := h.svc.ListTasks()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, tasks)
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var t domain.CrawlTask
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil || t.Name == "" || t.Query == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "任务名称与关键词必填")
		return
	}
	created, err := h.svc.CreateTask(t)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, created)
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	var t domain.CrawlTask
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.UpdateTask(r.PathValue("id"), t); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已更新"})
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteTask(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

func (h *Handler) runTask(w http.ResponseWriter, r *http.Request) {
	record, err := h.svc.RunTask(r.PathValue("id"))
	if err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, record)
}

func (h *Handler) stopTask(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.StopTask(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已停止"})
}

// ---------- 执行记录 ----------

func (h *Handler) listExecutions(w http.ResponseWriter, _ *http.Request) {
	records, err := h.svc.ListExecutions()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, records)
}

func (h *Handler) getExecution(w http.ResponseWriter, r *http.Request) {
	record, err := h.svc.GetExecution(r.PathValue("id"))
	if err != nil {
		response.Fail(w, http.StatusNotFound, 40401, "执行记录不存在")
		return
	}
	response.OK(w, record)
}

// ---------- 失败任务 ----------

func (h *Handler) listFailures(w http.ResponseWriter, _ *http.Request) {
	failures, err := h.svc.ListFailures()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, failures)
}

func (h *Handler) deleteFailure(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RetryFailure(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已处理"})
}

// ---------- 爬虫配置 ----------

func (h *Handler) getConfig(w http.ResponseWriter, _ *http.Request) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, cfg)
}

func (h *Handler) saveConfig(w http.ResponseWriter, r *http.Request) {
	var cfg domain.CrawlerConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.SaveConfig(cfg); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, cfg)
}

// ---------- 抓取数据 ----------

func (h *Handler) listData(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListData(r.URL.Query().Get("status"))
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, items)
}

func (h *Handler) updateDataStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.UpdateDataStatus(r.PathValue("id"), body.Status); err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, map[string]string{"message": "状态已更新"})
}

func (h *Handler) deleteData(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteData(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// ---------- 文章 ----------

func (h *Handler) listArticles(w http.ResponseWriter, _ *http.Request) {
	arts, err := h.svc.ListArticles()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, arts)
}

func (h *Handler) createArticle(w http.ResponseWriter, r *http.Request) {
	var a domain.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil || a.Title == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "标题必填")
		return
	}
	created, err := h.svc.CreateArticle(a)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, created)
}

func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteArticle(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// ---------- SEO / 站点 ----------

func (h *Handler) getSeo(w http.ResponseWriter, _ *http.Request) {
	s, err := h.svc.GetSeo()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, s)
}

func (h *Handler) saveSeo(w http.ResponseWriter, r *http.Request) {
	var s domain.SeoConfig
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.SaveSeo(s); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, s)
}

func (h *Handler) getSiteConfig(w http.ResponseWriter, _ *http.Request) {
	s, err := h.svc.GetSiteConfig()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, s)
}

func (h *Handler) saveSiteConfig(w http.ResponseWriter, r *http.Request) {
	var s domain.SiteConfig
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.SaveSiteConfig(s); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, s)
}
