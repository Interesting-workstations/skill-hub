// Package admin HTTP 层：参数解析、认证校验、调用 Service、统一响应。
package admin

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/Interesting-workstations/skill-hub/server/internal/response"
	"github.com/gorilla/websocket"
)

// Handler 后台管理 HTTP 处理器。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册后台管理路由。登录/刷新/退出公开，其余需要 Bearer Token。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/admin/login", h.login)
	mux.HandleFunc("POST /api/v1/admin/refresh", h.refresh)
	mux.HandleFunc("POST /api/v1/admin/logout", h.logout)
	mux.HandleFunc("GET /api/v1/admin/health", h.health)

	h.protected(mux, "GET /api/v1/admin/session", h.session)
	h.protected(mux, "GET /api/v1/admin/login-logs", h.loginLogs)
	h.protected(mux, "GET /api/v1/admin/stats", h.stats)

	// 执行实时推送：先换一次性 WS 票据，再建立 WebSocket 长连接
	h.protected(mux, "POST /api/v1/admin/ws-ticket", h.wsTicket)
	mux.HandleFunc("GET /api/v1/admin/executions/{id}/ws", h.execWS)

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
	h.protected(mux, "POST /api/v1/admin/executions/{id}/stop", h.stopExecution)
	h.protected(mux, "DELETE /api/v1/admin/executions/{id}", h.deleteExecution)

	// 失败任务
	h.protected(mux, "GET /api/v1/admin/failures", h.listFailures)
	h.protected(mux, "DELETE /api/v1/admin/failures/{id}", h.ignoreFailure)
	h.protected(mux, "POST /api/v1/admin/failures/{id}/retry", h.retryFailure)

	// 爬虫配置
	h.protected(mux, "GET /api/v1/admin/config", h.getConfig)
	h.protected(mux, "PUT /api/v1/admin/config", h.saveConfig)
	// GitHub Token 池（独立页：表格管理 + 一键检测）
	h.protected(mux, "GET /api/v1/admin/tokens", h.listTokens)
	h.protected(mux, "POST /api/v1/admin/tokens", h.createToken)
	h.protected(mux, "PUT /api/v1/admin/tokens/{id}", h.updateToken)
	h.protected(mux, "DELETE /api/v1/admin/tokens/{id}", h.deleteToken)
	h.protected(mux, "POST /api/v1/admin/tokens/check", h.checkTokens)

	// 官方组织（动态管理）
	h.protected(mux, "GET /api/v1/admin/official-orgs", h.listOfficialOrgs)
	h.protected(mux, "POST /api/v1/admin/official-orgs", h.createOfficialOrg)
	h.protected(mux, "PUT /api/v1/admin/official-orgs/{owner}", h.updateOfficialOrg)
	h.protected(mux, "DELETE /api/v1/admin/official-orgs/{owner}", h.deleteOfficialOrg)
	// 一键校验：owner 是否为真正的 GitHub 组织（排查个人账号 / 不存在）
	h.protected(mux, "GET /api/v1/admin/official-orgs/verify", h.verifyOfficialOrgs)
	h.protected(mux, "GET /api/v1/admin/export", h.exportData)

	// 抓取数据（数据审核）
	h.protected(mux, "GET /api/v1/admin/data", h.listData)
	h.protected(mux, "PUT /api/v1/admin/data/{id}/status", h.updateDataStatus)
	h.protected(mux, "POST /api/v1/admin/data/batch-status", h.batchUpdateDataStatus)
	h.protected(mux, "POST /api/v1/admin/data/auto-audit", h.autoAuditData)
	h.protected(mux, "POST /api/v1/admin/data/publish-all-approved", h.publishAllApproved)
	h.protected(mux, "DELETE /api/v1/admin/data/{id}", h.deleteData)

	// 文章
	h.protected(mux, "GET /api/v1/admin/articles", h.listArticles)
	h.protected(mux, "POST /api/v1/admin/articles", h.createArticle)
	h.protected(mux, "PUT /api/v1/admin/articles/{id}", h.updateArticle)
	h.protected(mux, "DELETE /api/v1/admin/articles/{id}", h.deleteArticle)

	// 赞助商：公开接口（官网渲染 / 点击上报）+ 后台管理
	mux.HandleFunc("GET /api/v1/sponsors", h.listPublicSponsors)
	mux.HandleFunc("POST /api/v1/sponsors/{id}/click", h.incrSponsorClick)
	h.protected(mux, "GET /api/v1/admin/sponsors", h.listSponsors)
	h.protected(mux, "POST /api/v1/admin/sponsors", h.createSponsor)
	h.protected(mux, "PUT /api/v1/admin/sponsors/{id}", h.updateSponsor)
	h.protected(mux, "DELETE /api/v1/admin/sponsors/{id}", h.deleteSponsor)

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
	// 参数合法性校验（bcrypt 密码上限 72 字节）
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) > 64 || req.Password == "" || len(req.Password) > 72 {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	result, err := h.svc.Login(req.Username, req.Password, clientIP(r), r.UserAgent())
	if err != nil {
		switch {
		case errors.Is(err, ErrAccountLocked):
			response.Fail(w, http.StatusUnauthorized, 40103, ErrAccountLocked.Error())
		case errors.Is(err, ErrTooManyAttempts):
			response.Fail(w, http.StatusTooManyRequests, 42901, ErrTooManyAttempts.Error())
		default:
			// 统一提示，不暴露用户名是否存在或密码错误的具体原因
			response.Fail(w, http.StatusUnauthorized, 40102, ErrBadCredentials.Error())
		}
		return
	}
	response.OK(w, result)
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

// refresh 用 Refresh Token 换取新凭证（一次性旋转）。
func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	result, err := h.svc.RefreshToken(strings.TrimSpace(req.RefreshToken))
	if err != nil {
		response.Fail(w, http.StatusUnauthorized, 40102, "登录已过期，请重新登录")
		return
	}
	response.OK(w, result)
}

// logout 主动退出：Access Token 进黑名单、Refresh Token 作废。
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	access := ""
	if header := r.Header.Get("Authorization"); len(header) > 7 {
		access = header[7:]
	}
	refresh := ""
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
		refresh = req.RefreshToken
	}
	h.svc.Logout(access, refresh, clientIP(r), r.UserAgent())
	response.OK(w, map[string]string{"message": "已退出登录"})
}

// session 会话状态检查（受保护，前端启动时用于校验 Token 有效性）。
func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	username := authUser(h.svc, r)
	u, err := h.svc.Session(username)
	if err != nil {
		response.Fail(w, http.StatusUnauthorized, 40101, "未登录或登录已过期")
		return
	}
	response.OK(w, u)
}

// loginLogs 登录审计日志（受保护）。
func (h *Handler) loginLogs(w http.ResponseWriter, _ *http.Request) {
	logs, err := h.svc.ListLoginLogs(50)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, logs)
}

type pwdReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	username := authUser(h.svc, r)
	var req pwdReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.OldPassword == "" || req.NewPassword == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.ChangePassword(username, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, ErrWeakPassword):
			response.Fail(w, http.StatusBadRequest, 40002, ErrWeakPassword.Error())
		default:
			response.Fail(w, http.StatusBadRequest, 40002, ErrBadOldPassword.Error())
		}
		return
	}
	response.OK(w, map[string]string{"message": "密码已更新，请重新登录"})
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

// clientIP 提取客户端 IP（优先 X-Forwarded-For，回退 RemoteAddr）。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
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

// GET /api/v1/admin/official-orgs —— 官方组织列表。
func (h *Handler) listOfficialOrgs(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListOfficialOrgs()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, list)
}

// GET /api/v1/admin/official-orgs/verify —— 一键校验所有官方组织的 GitHub 类型。
func (h *Handler) verifyOfficialOrgs(w http.ResponseWriter, _ *http.Request) {
	response.OK(w, h.svc.VerifyOfficialOrgs())
}

// POST /api/v1/admin/official-orgs —— 新增官方组织。
func (h *Handler) createOfficialOrg(w http.ResponseWriter, r *http.Request) {
	var o domain.OfficialOrg
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	created, err := h.svc.CreateOfficialOrg(o)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, err.Error())
		return
	}
	response.OK(w, created)
}

// PUT /api/v1/admin/official-orgs/{owner} —— 更新官方组织。
func (h *Handler) updateOfficialOrg(w http.ResponseWriter, r *http.Request) {
	var o domain.OfficialOrg
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.UpdateOfficialOrg(r.PathValue("owner"), o); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已更新"})
}

// DELETE /api/v1/admin/official-orgs/{owner} —— 删除官方组织。
func (h *Handler) deleteOfficialOrg(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteOfficialOrg(r.PathValue("owner")); err != nil {
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

// stopExecution 停止一次执行（取消运行中的爬虫，标记执行记录为已停止）。
func (h *Handler) stopExecution(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.StopExecution(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已停止"})
}

// deleteExecution 删除一条执行记录（仍在运行会先取消）。
func (h *Handler) deleteExecution(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteExecution(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// ---------- 执行实时推送（WebSocket）----------

// wsTicket 为 WebSocket 连接签发一次性票据（避免把 Access Token 放进 URL）。
func (h *Handler) wsTicket(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExecID string `json:"execId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ExecID) == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	ticket, err := h.svc.CreateWsTicket(strings.TrimSpace(req.ExecID))
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"ticket": ticket})
}

// execWS 建立 WebSocket 长连接，实时推送执行进度 / 日志 / 状态事件。
// 鉴权：连接前先经 ws-ticket 换取一次性票据（?ticket=xxx），避免 token 进 URL。
func (h *Handler) execWS(w http.ResponseWriter, r *http.Request) {
	execID := r.PathValue("id")
	if got, ok := h.svc.ConsumeWsTicket(r.URL.Query().Get("ticket")); !ok || got != execID {
		http.Error(w, "无效或过期的连接票据", http.StatusUnauthorized)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 4096,
		// 单管理员后台 + 一次性票据鉴权；生产多域部署时可收紧为白名单域名
		CheckOrigin: func(*http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade exec=%s failed: %v", execID, err)
		return
	}
	defer conn.Close()

	ch, cancel := h.svc.SubscribeExec(execID)
	defer cancel()

	// 推送初始快照：历史日志 + 当前进度 / 状态
	if record, err := h.svc.GetExecution(execID); err == nil {
		_ = conn.WriteJSON(domain.ExecEvent{
			Type:     "snapshot",
			ExecID:   execID,
			Progress: record.Progress,
			Status:   record.Status,
			Logs:     record.Logs,
			Duration: record.Duration,
		})
	}

	// 读循环：消费客户端消息（含 Ping）以检测连接存活
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.SetReadLimit(1024)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteJSON(ev); err != nil {
				return
			}
		case <-ping.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				return
			}
		case <-done:
			return
		}
	}
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
	if err := h.svc.IgnoreFailure(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已处理"})
}

// retryFailure 重新执行失败任务（POST /api/v1/admin/failures/{id}/retry）。
func (h *Handler) retryFailure(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.RetryFailure(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已重新执行"})
}

// ignoreFailure 忽略失败任务（DELETE /api/v1/admin/failures/{id}）。
func (h *Handler) ignoreFailure(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.IgnoreFailure(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已忽略"})
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

// checkTokens 一键检测 GitHub Token 池中每个 token 是否可用（POST /api/v1/admin/tokens/check）。
// 请求体可选 tokens 数组；为空时检测当前配置中的 token。
func (h *Handler) checkTokens(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Tokens []string `json:"tokens"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	results := h.svc.CheckTokens(body.Tokens)
	response.OK(w, results)
}

// ---------- GitHub Token 池 ----------

// listTokens 返回 token 池全部条目（token 脱敏展示）。
func (h *Handler) listTokens(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListTokenPool()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, list)
}

// createToken 新增一个 token。
func (h *Handler) createToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token  string `json:"token"`
		Remark string `json:"remark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Token) == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "token 不能为空")
		return
	}
	created, err := h.svc.CreateToken(body.Token, body.Remark)
	if err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, created)
}

// updateToken 更新 token 条目（token/remark/enabled）。
func (h *Handler) updateToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token   string `json:"token"`
		Remark  string `json:"remark"`
		Enabled *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.UpdateToken(r.PathValue("id"), body.Token, body.Remark, body.Enabled); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已更新"})
}

// deleteToken 删除一个 token。
func (h *Handler) deleteToken(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteToken(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// ---------- 抓取数据 ----------

func (h *Handler) listData(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := DataFilter{
		Status:   q.Get("status"),
		Source:   q.Get("source"),
		Category: q.Get("category"),
		Author:   q.Get("author"),
		Query:    q.Get("q"),
		Sort:     q.Get("sort"),
	}
	items, err := h.svc.ListData(f)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, items)
}

// POST /api/v1/admin/data/batch-status —— 批量更新数据状态（审核页全选操作）。
func (h *Handler) batchUpdateDataStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Fail(w, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if err := h.svc.UpdateDataStatusBatch(body.IDs, body.Status); err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, map[string]int{"updated": len(body.IDs)})
}

// POST /api/v1/admin/data/auto-audit —— 机器人自动审核（内容完整规范的直接通过）。
func (h *Handler) autoAuditData(w http.ResponseWriter, _ *http.Request) {
	res, err := h.svc.AutoAuditPending()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, res)
}

// POST /api/v1/admin/data/publish-all-approved —— 一键发布全部已审核数据。
func (h *Handler) publishAllApproved(w http.ResponseWriter, _ *http.Request) {
	n, err := h.svc.PublishAllApproved()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]int64{"published": n})
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

func (h *Handler) updateArticle(w http.ResponseWriter, r *http.Request) {
	var a domain.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil || a.Title == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "标题必填")
		return
	}
	updated, err := h.svc.UpdateArticle(r.PathValue("id"), a)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, updated)
}

func (h *Handler) deleteArticle(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteArticle(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// ---------- 赞助商 ----------

// listPublicSponsors 官网公开接口：返回启用中的赞助商。
// GET /api/v1/sponsors?position=home|sidebar（缺省返回全部位置的启用赞助商）
func (h *Handler) listPublicSponsors(w http.ResponseWriter, r *http.Request) {
	position := r.URL.Query().Get("position")
	all, err := h.svc.ListSponsors()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	out := make([]domain.Sponsor, 0, len(all))
	for _, s := range all {
		if !s.Enabled {
			continue
		}
		if position != "" && s.Position != position && s.Position != "both" {
			continue
		}
		out = append(out, s)
	}
	response.OK(w, out)
}

// incrSponsorClick 官网点击上报（公开）。
func (h *Handler) incrSponsorClick(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.IncrSponsorClicks(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]bool{"ok": true})
}

func (h *Handler) listSponsors(w http.ResponseWriter, _ *http.Request) {
	list, err := h.svc.ListSponsors()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, list)
}

func (h *Handler) createSponsor(w http.ResponseWriter, r *http.Request) {
	var s domain.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil || s.Name == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "名称必填")
		return
	}
	created, err := h.svc.CreateSponsor(s)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, created)
}

func (h *Handler) updateSponsor(w http.ResponseWriter, r *http.Request) {
	var s domain.Sponsor
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil || s.Name == "" {
		response.Fail(w, http.StatusBadRequest, 40001, "名称必填")
		return
	}
	updated, err := h.svc.UpdateSponsor(r.PathValue("id"), s)
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, updated)
}

func (h *Handler) deleteSponsor(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteSponsor(r.PathValue("id")); err != nil {
		response.Fail(w, http.StatusInternalServerError, 50001, "系统错误")
		return
	}
	response.OK(w, map[string]string{"message": "已删除"})
}

// exportData 导出技能数据（GET /api/v1/admin/export?format=json|csv|markdown&scope=all|published|approved）。
func (h *Handler) exportData(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	result, err := h.svc.ExportData(format, r.URL.Query().Get("scope"))
	if err != nil {
		response.Fail(w, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.OK(w, result)
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
