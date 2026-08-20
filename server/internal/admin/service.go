// Package admin 管理后台业务层：任务/执行/失败/配置/内容/认证。
package admin

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/Interesting-workstations/skill-hub/server/internal/translate"
)

// ---------- 认证安全常量 ----------
const (
	accessTokenTTL  = 2 * time.Hour      // Access Token 有效期
	refreshTokenTTL = 7 * 24 * time.Hour // Refresh Token 有效期
	maxLoginFails   = 5                  // 连续失败锁定阈值
	lockDuration    = 15 * time.Minute   // 临时锁定时长
	loginRateLimit  = 10                 // 每分钟每 IP 最大登录尝试次数
	loginRateWindow = time.Minute
)

// 认证错误（对外统一提示，不暴露具体失败原因）
var (
	ErrBadCredentials  = errors.New("用户名或密码错误")
	ErrAccountLocked   = errors.New("尝试次数过多，账号已临时锁定，请稍后再试")
	ErrTooManyAttempts = errors.New("尝试过于频繁，请稍后再试")
	ErrWeakPassword    = errors.New("新密码长度至少 8 位")
	ErrBadOldPassword  = errors.New("当前密码错误")
)

// refreshEntry Refresh Token 内存条目（一次性，使用后删除）。
type refreshEntry struct {
	username string
	version  int
	expiry   int64
}

// wsTicket WebSocket 一次性连接票据。
type wsTicket struct {
	execID string
	expiry time.Time
}

// Service 后台管理业务逻辑。
// 认证状态（黑名单 / Refresh Token / Token 版本 / 登录限流）为进程内内存态：
// 与进程级签名密钥生命周期一致（重启后旧 Token 全部失效，状态自洽）。
type Service struct {
	repo   Repository
	client *crawler.Client
	token  []byte // 进程内签名密钥（重启后旧 token 失效）
	mu     sync.Mutex

	// tokenPool GitHub Token 池（多 token 故障切换）；后台配置变更后刷新。
	tokenPool *crawler.TokenPool

	// crawlSem 全局爬虫并发信号量：限制同时运行的爬虫任务数，
	// 避免多个定时任务同时触发打满 GitHub API 配额。
	crawlSem chan struct{}

	// translator 技能标题/描述的中文翻译器（新爬取数据入库时生成 name_zh / description_zh）
	translator *translate.Translator

	blacklist     map[string]int64        // access token hash → 过期时间（主动登出）
	refreshTokens map[string]refreshEntry // refresh token hash → 条目
	tokenVersions map[string]int          // username → token version（改密后 +1 全失效）
	loginAttempts map[string][]int64      // ip → 登录尝试时间戳（滑动窗口限流）

	execSubs    map[string]map[chan domain.ExecEvent]struct{} // execID → 订阅者（执行进度/日志推送）
	wsTickets   map[string]wsTicket                           // ticket → 一次性 WS 票据
	execCancels map[string]context.CancelFunc                 // execID → 取消函数（手动停止任务）
}

// NewService 创建后台管理服务。
func NewService(repo Repository) *Service {
	s := &Service{
		repo:          repo,
		token:         adminTokenSecret(),
		translator:    translate.New(),
		blacklist:     make(map[string]int64),
		refreshTokens: make(map[string]refreshEntry),
		tokenVersions: make(map[string]int),
		loginAttempts: make(map[string][]int64),
		execSubs:      make(map[string]map[chan domain.ExecEvent]struct{}),
		wsTickets:     make(map[string]wsTicket),
		execCancels:   make(map[string]context.CancelFunc),
		crawlSem:      make(chan struct{}, 1), // 同一时间最多 1 个爬虫任务在跑，保护 API 配额
	}
	// 从环境变量初始化 token 池（后台配置加载后由 RefreshTokenPool 覆盖）
	s.client = crawler.NewClientFromEnv()
	// github_tokens 表为空时，把环境变量中的 token 同步入库，便于后台配置页直接展示与管理
	s.seedTokenPoolFromEnv()
	s.RefreshTokenPool()
	// 应用后台保存的翻译主通道配置（覆盖环境变量默认）
	s.applyTranslateProvider()
	return s
}

// seedTokenPoolFromEnv github_tokens 表为空时，把环境变量（GITHUB_TOKENS / GITHUB_TOKEN）
// 中的 token 逐条写入独立表（首次启动或从旧版升级时兜底），之后以表为准。
func (s *Service) seedTokenPoolFromEnv() {
	rows, err := s.repo.ListGitHubTokens()
	if err != nil || len(rows) > 0 {
		return // 表已有数据或读取失败，跳过
	}
	for _, t := range crawler.TokensFromEnv() {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		_ = s.repo.CreateGitHubToken(&domain.GitHubToken{
			ID:        newID("tok"),
			Token:     t,
			Remark:    "环境变量初始化",
			Enabled:   true,
			CreatedAt: today(),
		})
	}
}

// RefreshTokenPool 从数据库 github_tokens 表刷新 token 池（仅启用项），
// 并同步到共享 client。表为空时回退环境变量（GITHUB_TOKENS / GITHUB_TOKEN）。
// 后台增删改 token 后调用，立即对后续爬虫生效。
func (s *Service) RefreshTokenPool() {
	tokens := s.repoGitHubTokens()
	if len(tokens) == 0 {
		tokens = crawler.TokensFromEnv()
	}
	s.tokenPool = crawler.NewTokenPool(tokens)
	if s.client == nil {
		s.client = crawler.NewClient("")
	}
	s.client.SetTokenPool(s.tokenPool)
	if len(tokens) > 0 {
		log.Printf("🔑 GitHub Token 池已刷新：%d 个 token", len(tokens))
	}
}

// repoGitHubTokens 从数据库 github_tokens 表读取启用的 token 列表。
func (s *Service) repoGitHubTokens() []string {
	rows, err := s.repo.ListGitHubTokens()
	if err != nil {
		return nil
	}
	var out []string
	for _, t := range rows {
		if t.Enabled && strings.TrimSpace(t.Token) != "" {
			out = append(out, strings.TrimSpace(t.Token))
		}
	}
	return out
}

// currentTokens 返回当前生效的 token 列表：优先 token 池（数据库启用项），回退环境变量。
func (s *Service) currentTokens() []string {
	if s.tokenPool != nil {
		if ts := s.tokenPool.Tokens(); len(ts) > 0 {
			return ts
		}
	}
	if ts := s.repoGitHubTokens(); len(ts) > 0 {
		return ts
	}
	return crawler.TokensFromEnv()
}

// CheckTokens 健康检查给定 token 列表（为空则检查当前池），返回脱敏结果供后台展示。
func (s *Service) CheckTokens(tokens []string) []crawler.TokenHealth {
	if len(tokens) == 0 {
		tokens = s.repoGitHubTokens()
	}
	if len(tokens) == 0 {
		tokens = crawler.TokensFromEnv()
	}
	hc := &http.Client{Timeout: 12 * time.Second}
	out := make([]crawler.TokenHealth, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, crawler.CheckTokenHealth(hc, t))
	}
	return out
}

// ---------- GitHub Token 池（后台管理） ----------

// ListTokenPool 返回 token 池全部条目（token 脱敏展示，避免后台页面回显明文），
// 并附带运行时状态：当前是否被熔断（限流/失效冷却中）。
func (s *Service) ListTokenPool() ([]domain.GitHubToken, error) {
	rows, err := s.repo.ListGitHubTokens()
	if err != nil {
		return nil, err
	}
	// 当前 token 池的运行时熔断状态（原始 token → 状态）
	states := map[string]crawler.TokenState{}
	if s.tokenPool != nil {
		for _, st := range s.tokenPool.States() {
			states[st.Token] = st
		}
	}
	for i := range rows {
		// 先用原始 token 匹配熔断状态，再脱敏展示
		if st, ok := states[rows[i].Token]; ok {
			rows[i].Broken = st.Broken
			rows[i].CooldownAt = st.CooldownAt
		}
		rows[i].Token = maskSecret(rows[i].Token)
	}
	return rows, nil
}

// CreateToken 新增一个 token 到池中，并刷新运行中的 token 池。
func (s *Service) CreateToken(token, remark string) (domain.GitHubToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return domain.GitHubToken{}, fmt.Errorf("token 不能为空")
	}
	// 去重：同一 token 不重复入库
	if rows, err := s.repo.ListGitHubTokens(); err == nil {
		for _, t := range rows {
			if t.Token == token {
				return domain.GitHubToken{}, fmt.Errorf("该 token 已存在")
			}
		}
	}
	t := domain.GitHubToken{
		ID:        newID("tok"),
		Token:     token,
		Remark:    strings.TrimSpace(remark),
		Enabled:   true,
		CreatedAt: today(),
	}
	if err := s.repo.CreateGitHubToken(&t); err != nil {
		return domain.GitHubToken{}, err
	}
	s.RefreshTokenPool()
	t.Token = maskSecret(t.Token)
	return t, nil
}

// UpdateToken 更新 token 条目（token/remark/enabled；token 为空表示不修改），并刷新池。
func (s *Service) UpdateToken(id, token, remark string, enabled *bool) error {
	fields := map[string]any{}
	if token = strings.TrimSpace(token); token != "" {
		fields["token"] = token
	}
	fields["remark"] = strings.TrimSpace(remark)
	if enabled != nil {
		v := 0
		if *enabled {
			v = 1
		}
		fields["enabled"] = v
	}
	if err := s.repo.UpdateGitHubToken(id, fields); err != nil {
		return err
	}
	s.RefreshTokenPool()
	return nil
}

// DeleteToken 删除 token 条目，并刷新池。
func (s *Service) DeleteToken(id string) error {
	if err := s.repo.DeleteGitHubToken(id); err != nil {
		return err
	}
	s.RefreshTokenPool()
	return nil
}

// maskSecret 脱敏 token/密钥：保留前 8 位与后 4 位，中间省略。
func maskSecret(s string) string {
	if len(s) <= 12 {
		return "****"
	}
	return s[:8] + "****" + s[len(s)-4:]
}

// adminTokenSecret 获取持久化的 Access Token 签名密钥。
// 优先环境变量 ADMIN_TOKEN_SECRET（推荐：重启容器后已签发的 Token 依然有效）；
// 否则尝试读取 /app/data/.token_secret 文件（进程可写时持久化，避免每次重启随机导致旧 Token 全部失效）；
// 最后随机生成（仅兜底）。
func adminTokenSecret() []byte {
	if env := os.Getenv("ADMIN_TOKEN_SECRET"); len(env) >= 32 {
		return []byte(env)
	}
	const path = "/app/data/.token_secret"
	if b, err := os.ReadFile(path); err == nil && len(b) >= 32 {
		return b
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	_ = os.WriteFile(path, b, 0o600)
	return b
}

func now() string   { return time.Now().Format("2006-01-02 15:04:05") }
func today() string { return time.Now().Format("2006-01-02") }

func newID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}

// ---------- 认证 ----------

// Login 完整登录校验流程，返回 Access + Refresh 凭证。
// 流程：限流 → 查用户 → 锁定检查 → 失败次数检查 → 密码校验 → 失败计数/锁定 →
// 成功则重置失败计数、更新登录信息、记录审计日志并签发 Token。
func (s *Service) Login(username, password, ip, ua string) (domain.LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return domain.LoginResult{}, ErrBadCredentials
	}

	// 1. 登录限流（IP 维度，防暴力破解）
	if !s.allowLoginAttempt(ip) {
		s.logLogin(username, "fail", ip, ua)
		return domain.LoginResult{}, ErrTooManyAttempts
	}

	// 2. 查询账号（不存在与密码错误返回同一提示，避免暴露账号是否存在）
	rec, err := s.repo.GetAdminAuth(username)
	if err != nil {
		s.recordLoginFail(username, ip, ua)
		return domain.LoginResult{}, ErrBadCredentials
	}

	// 3. 账号锁定检查（临时锁定）
	if rec.LockedUntil != "" && rec.LockedUntil > now() {
		return domain.LoginResult{}, ErrAccountLocked
	}

	// 4. 连续失败已达阈值 → 自动锁定
	if rec.FailCount >= maxLoginFails {
		s.lockAccount(username)
		return domain.LoginResult{}, ErrAccountLocked
	}

	// 5. 密码校验（bcrypt 安全比较；兼容旧 SHA-256 格式）
	ok, legacy := VerifyPassword(rec.PasswordHash, password)
	if !ok {
		s.recordLoginFail(username, ip, ua)
		return domain.LoginResult{}, ErrBadCredentials
	}

	// 6. 旧格式密码自动迁移为 bcrypt（登录时一次性的）
	if legacy {
		if h, err := HashPassword(password); err == nil {
			_ = s.repo.UpdateAdminPassword(username, h)
		}
	}

	// 7. 登录成功：重置失败计数、更新最后登录信息、记录日志、签发凭证
	_ = s.repo.ResetAdminFail(username)
	_ = s.repo.UpdateAdminLogin(username, truncate(ip, 64), truncate(ua, 255), now())
	s.logLogin(username, "success", ip, ua)

	access, refresh, err := s.issueTokens(username)
	if err != nil {
		return domain.LoginResult{}, err
	}
	user := domain.AdminUser{Username: rec.Username, DisplayName: rec.DisplayName, Role: "admin"}
	return domain.LoginResult{Token: access, RefreshToken: refresh, User: user}, nil
}

// RefreshToken 用 Refresh Token 换取新的 Access + Refresh（一次性，旋转）。
// 使用后立即作废旧 Refresh，防止重放攻击。
func (s *Service) RefreshToken(refresh string) (domain.LoginResult, error) {
	rh := hashToken(refresh)
	s.mu.Lock()
	entry, ok := s.refreshTokens[rh]
	if ok {
		delete(s.refreshTokens, rh) // 一次性：用后即焚
	}
	s.mu.Unlock()
	if !ok {
		return domain.LoginResult{}, ErrBadCredentials
	}
	if time.Now().Unix() > entry.expiry {
		return domain.LoginResult{}, ErrBadCredentials
	}

	// Token Version 校验：改密后旧 Refresh 一律失效
	s.mu.Lock()
	v := s.tokenVersions[entry.username]
	s.mu.Unlock()
	if v != entry.version {
		return domain.LoginResult{}, ErrBadCredentials
	}

	access, newRefresh, err := s.issueTokens(entry.username)
	if err != nil {
		return domain.LoginResult{}, err
	}
	user, err := s.sessionUser(entry.username)
	if err != nil {
		return domain.LoginResult{}, ErrBadCredentials
	}
	return domain.LoginResult{Token: access, RefreshToken: newRefresh, User: user}, nil
}

// Logout 主动退出：Access Token 加入黑名单、Refresh Token 作废，并记录审计日志。
func (s *Service) Logout(accessToken, refreshToken, ip, ua string) {
	username := ""
	s.mu.Lock()
	if accessToken != "" {
		parts := strings.Split(accessToken, ".")
		if len(parts) == 4 {
			username = parts[0]
			if expiry, err := strconv.ParseInt(parts[2], 10, 64); err == nil && expiry > time.Now().Unix() {
				s.blacklist[hashToken(accessToken)] = expiry
			}
		}
	}
	if refreshToken != "" {
		delete(s.refreshTokens, hashToken(refreshToken))
	}
	s.mu.Unlock()
	if username != "" {
		s.logLogin(username, "logout", ip, ua)
	}
}

// VerifyToken 校验 Access Token：格式、签名、过期、Token 版本、黑名单。
func (s *Service) VerifyToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 4 {
		return "", false
	}
	username, versionStr, expiryStr, sig := parts[0], parts[1], parts[2], parts[3]
	var version, expiry int64
	if _, err := fmt.Sscanf(versionStr, "%d", &version); err != nil {
		return "", false
	}
	if _, err := fmt.Sscanf(expiryStr, "%d", &expiry); err != nil {
		return "", false
	}
	if time.Now().Unix() > expiry {
		return "", false
	}

	// 签名校验（恒定时间比较）
	mac := hmac.New(sha256.New, s.token)
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d:%d", username, version, expiry)))
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return "", false
	}

	// Token Version 校验（改密后旧 Access 失效）
	s.mu.Lock()
	v, hasVer := s.tokenVersions[username]
	_, banned := s.blacklist[hashToken(token)]
	s.mu.Unlock()
	if hasVer && int(version) != v {
		return "", false
	}
	if banned {
		return "", false
	}
	return username, true
}

// Session 返回当前登录用户信息（供会话状态检查）。
func (s *Service) Session(username string) (domain.AdminUser, error) {
	return s.sessionUser(username)
}

// ListLoginLogs 查询最近登录审计日志（供后台审计查看）。
func (s *Service) ListLoginLogs(limit int) ([]domain.AdminLoginLog, error) {
	return s.repo.GetLoginLogs(limit)
}

// ---------- 执行实时推送（WebSocket）----------

// SubscribeExec 订阅某执行记录的实时事件，返回事件通道与取消函数。
func (s *Service) SubscribeExec(execID string) (<-chan domain.ExecEvent, func()) {
	ch := make(chan domain.ExecEvent, 128)
	s.mu.Lock()
	if s.execSubs[execID] == nil {
		s.execSubs[execID] = make(map[chan domain.ExecEvent]struct{})
	}
	s.execSubs[execID][ch] = struct{}{}
	s.mu.Unlock()

	cancel := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if m := s.execSubs[execID]; m != nil {
			delete(m, ch)
			if len(m) == 0 {
				delete(s.execSubs, execID)
			}
		}
	}
	return ch, cancel
}

// publishExec 向某执行记录的所有订阅者广播事件（非阻塞，慢消费者丢弃）。
func (s *Service) publishExec(execID string, ev domain.ExecEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for ch := range s.execSubs[execID] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// CreateWsTicket 为指定执行记录生成一次性 WebSocket 票据（5 分钟有效）。
func (s *Service) CreateWsTicket(execID string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(b)
	s.mu.Lock()
	s.wsTickets[ticket] = wsTicket{execID: execID, expiry: time.Now().Add(5 * time.Minute)}
	s.mu.Unlock()
	return ticket, nil
}

// ConsumeWsTicket 校验并消费一次性票据（用后即焚，防重放），返回执行 ID。
func (s *Service) ConsumeWsTicket(ticket string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.wsTickets[ticket]
	if !ok {
		return "", false
	}
	delete(s.wsTickets, ticket)
	if time.Now().After(t.expiry) {
		return "", false
	}
	return t.execID, true
}

// ChangePassword 修改密码：校验旧密码、强度校验、bcrypt 重哈希，并使全部旧 Token 失效。
func (s *Service) ChangePassword(username, oldPwd, newPwd string) error {
	if len(newPwd) < 8 {
		return ErrWeakPassword
	}
	rec, err := s.repo.GetAdminAuth(username)
	if err != nil {
		return ErrBadOldPassword
	}
	ok, _ := VerifyPassword(rec.PasswordHash, oldPwd)
	if !ok {
		return ErrBadOldPassword
	}
	hash, err := HashPassword(newPwd)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateAdminPassword(username, hash); err != nil {
		return err
	}
	// 使全部旧 Token 失效：版本号 +1，清空该用户所有 Refresh Token
	s.mu.Lock()
	s.tokenVersions[username]++
	for rh, e := range s.refreshTokens {
		if e.username == username {
			delete(s.refreshTokens, rh)
		}
	}
	s.mu.Unlock()
	return nil
}

// ---------- 认证内部辅助 ----------

// issueTokens 为指定用户签发 Access + Refresh（互斥保护内存态）。
func (s *Service) issueTokens(username string) (access, refresh string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	version := s.tokenVersions[username]
	expiry := time.Now().Add(accessTokenTTL).Unix()
	mac := hmac.New(sha256.New, s.token)
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d:%d", username, version, expiry)))
	access = fmt.Sprintf("%s.%d.%d.%s", username, version, expiry, hex.EncodeToString(mac.Sum(nil)))

	rb := make([]byte, 32)
	if _, err = rand.Read(rb); err != nil {
		return "", "", err
	}
	refresh = hex.EncodeToString(rb)
	s.refreshTokens[hashToken(refresh)] = refreshEntry{
		username: username,
		version:  version,
		expiry:   time.Now().Add(refreshTokenTTL).Unix(),
	}
	return access, refresh, nil
}

// sessionUser 从仓库读取用户展示信息。
func (s *Service) sessionUser(username string) (domain.AdminUser, error) {
	rec, err := s.repo.GetAdmin(username)
	if err != nil {
		return domain.AdminUser{}, err
	}
	return domain.AdminUser{Username: rec.Username, DisplayName: rec.DisplayName, Role: "admin"}, nil
}

// recordLoginFail 记录一次登录失败：失败计数 +1、审计日志、达阈值则锁定。
func (s *Service) recordLoginFail(username, ip, ua string) {
	_ = s.repo.IncAdminFail(username)
	s.logLogin(username, "fail", ip, ua)
	if rec, err := s.repo.GetAdminAuth(username); err == nil && rec.FailCount >= maxLoginFails {
		s.lockAccount(username)
	}
}

// lockAccount 锁定账号直到 lockDuration 之后。
func (s *Service) lockAccount(username string) {
	until := time.Now().Add(lockDuration).Format("2006-01-02 15:04:05")
	_ = s.repo.SetAdminLocked(username, until)
}

// allowLoginAttempt 基于 IP 的滑动窗口登录限流。
func (s *Service) allowLoginAttempt(ip string) bool {
	nowTS := time.Now().Unix()
	cutoff := nowTS - int64(loginRateWindow/time.Second)
	s.mu.Lock()
	defer s.mu.Unlock()
	arr := s.loginAttempts[ip]
	kept := arr[:0]
	for _, t := range arr {
		if t > cutoff {
			kept = append(kept, t)
		}
	}
	if len(kept) >= loginRateLimit {
		s.loginAttempts[ip] = kept
		return false
	}
	s.loginAttempts[ip] = append(kept, nowTS)
	return true
}

// logLogin 写一条登录审计日志（绝不记录密码 / Token）。
func (s *Service) logLogin(username, action, ip, ua string) {
	_ = s.repo.InsertLoginLog(domain.AdminLoginLog{
		ID:        newID("login"),
		Username:  truncate(username, 64),
		Action:    action,
		IP:        truncate(ip, 64),
		UserAgent: truncate(ua, 255),
		CreatedAt: now(),
	})
}

// hashToken 对 Token 做 SHA-256 摘要（内存态 key，避免明文留存）。
func hashToken(t string) string {
	sum := sha256.Sum256([]byte(t))
	return hex.EncodeToString(sum[:])
}

// truncate 截断字符串到 n 个字符（用于长度受限的日志字段）。
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// ---------- 爬虫任务 ----------

func (s *Service) ListTasks() ([]domain.CrawlTask, error) {
	return s.repo.ListTasks()
}

func (s *Service) CreateTask(input domain.CrawlTask) (domain.CrawlTask, error) {
	t := domain.CrawlTask{
		ID:        newID("task"),
		Name:      input.Name,
		Type:      input.Type,
		Query:     input.Query,
		Status:    domain.TaskWaiting,
		Schedule:  input.Schedule,
		CreatedAt: today(),
	}
	if t.Schedule == "" {
		t.Schedule = "手动"
	}
	if t.Type == "" {
		t.Type = "skill"
	}
	return t, s.repo.CreateTask(&t)
}

func (s *Service) UpdateTask(id string, patch domain.CrawlTask) error {
	fields := map[string]any{}
	if patch.Name != "" {
		fields["name"] = patch.Name
	}
	if patch.Type != "" {
		fields["type"] = patch.Type
	}
	if patch.Query != "" {
		fields["query"] = patch.Query
	}
	if patch.Schedule != "" {
		fields["schedule"] = patch.Schedule
	}
	return s.repo.UpdateTask(id, fields)
}

func (s *Service) DeleteTask(id string) error {
	return s.repo.DeleteTask(id)
}

// ---------- 官方组织（动态管理） ----------

func (s *Service) ListOfficialOrgs() ([]domain.OfficialOrg, error) {
	return s.repo.ListOfficialOrgs()
}

func (s *Service) CreateOfficialOrg(input domain.OfficialOrg) (domain.OfficialOrg, error) {
	input.Owner = strings.TrimSpace(input.Owner)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.Owner == "" || input.DisplayName == "" {
		return domain.OfficialOrg{}, fmt.Errorf("GitHub 组织名与展示名必填")
	}
	input.CreatedAt = today()
	err := s.repo.CreateOfficialOrg(&input)
	return input, err
}

func (s *Service) UpdateOfficialOrg(owner string, patch domain.OfficialOrg) error {
	fields := map[string]any{}
	if patch.DisplayName != "" {
		fields["display_name"] = patch.DisplayName
	}
	if patch.Avatar != "" {
		fields["avatar"] = patch.Avatar
	}
	fields["logo_url"] = patch.LogoURL
	fields["sort_order"] = patch.SortOrder
	fields["enabled"] = 0
	if patch.Enabled {
		fields["enabled"] = 1
	}
	return s.repo.UpdateOfficialOrg(owner, fields)
}

func (s *Service) DeleteOfficialOrg(owner string) error {
	return s.repo.DeleteOfficialOrg(owner)
}

// VerifyOfficialOrgs 用 GitHub API 校验所有官方组织的 owner 类型与头像有效性，
// 找出个人账号（type=User）、不存在（NotFound）或头像无效（默认 identicon）的条目。
// 并发校验（限流）以缩短耗时：70+ 组织串行需 30s+，并发 8 约 4s。
func (s *Service) VerifyOfficialOrgs() []domain.OrgVerifyResult {
	orgs, err := s.repo.ListOfficialOrgs()
	if err != nil {
		return nil
	}
	out := make([]domain.OrgVerifyResult, len(orgs))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, o := range orgs {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, o domain.OfficialOrg) {
			defer wg.Done()
			defer func() { <-sem }()
			gt := "Error"
			avatarOK := false
			if chk, err := s.client.CheckOrg(o.Owner); err == nil {
				gt = chk.Type
				avatarOK = chk.AvatarOK
			}
			// 已显式使用官网 logo（非 GitHub 头像）的组织，GitHub 头像无效不再视为问题
			if usesOfficialLogo(o.LogoURL) {
				avatarOK = true
			}
			out[idx] = domain.OrgVerifyResult{
				Owner:       o.Owner,
				DisplayName: o.DisplayName,
				GitHubType:  gt,
				AvatarOK:    avatarOK,
				LogoURL:     o.LogoURL,
			}
		}(i, o)
	}
	wg.Wait()
	return out
}

// usesOfficialLogo 判断组织是否已显式使用官网 logo（非 GitHub 头像）。
// 这类组织的 GitHub 头像即使无效（默认 identicon）也不再需要关注。
func usesOfficialLogo(logoURL string) bool {
	return logoURL != "" && !strings.HasPrefix(logoURL, "https://github.com/")
}

// RunTask 启动一次真实爬虫执行（后台 goroutine，前端轮询执行记录）。
func (s *Service) RunTask(id string) (domain.ExecutionRecord, error) {
	task, err := s.repo.GetTask(id)
	if err != nil {
		return domain.ExecutionRecord{}, fmt.Errorf("任务不存在")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	record := domain.ExecutionRecord{
		ID:        newID("exec"),
		TaskID:    task.ID,
		TaskName:  task.Name,
		Status:    domain.TaskRunning,
		StartTime: now(),
		Logs: []domain.LogLine{
			{Time: time.Now().Format("15:04:05"), Level: "info", Text: fmt.Sprintf("开始执行任务「%s」", task.Name)},
			{Time: time.Now().Format("15:04:05"), Level: "info", Text: "初始化爬虫客户端"},
		},
	}
	if err := s.repo.CreateExecution(&record); err != nil {
		return domain.ExecutionRecord{}, err
	}
	if err := s.repo.UpdateTask(task.ID, map[string]any{"status": domain.TaskRunning, "last_run_at": now()[:16]}); err != nil {
		return domain.ExecutionRecord{}, err
	}

	// 每次执行使用独立爬虫客户端与可取消上下文（互不影响，支持后台手动停止）
	// token 池与共享 client 一致：后台配置的 GITHUB_TOKENS（失败自动切换）
	client := crawler.NewClientWithTokens(s.currentTokens())
	ctx, cancel := context.WithCancel(context.Background())
	s.execCancels[record.ID] = cancel

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.execCancels, record.ID)
			s.mu.Unlock()
		}()
		// 全局并发信号量：同一时间只跑 1 个爬虫任务，其余排队等待。
		// 避免多个定时任务同时触发，瞬间打满 GitHub API 配额导致大面积限流/失败。
		select {
		case s.crawlSem <- struct{}{}:
		case <-ctx.Done():
			// 等待期间被手动停止
			_ = s.repo.UpdateExecution(record.ID, map[string]any{
				"status": domain.TaskStopped, "progress": 100, "end_time": now(),
				"duration": "已手动停止（排队中）",
			})
			_ = s.repo.UpdateTask(task.ID, map[string]any{"status": domain.TaskWaiting})
			return
		}
		defer func() { <-s.crawlSem }()
		s.executeTask(ctx, client, task.ID, task.Name, task.Query, record.ID)
	}()
	return record, nil
}

// StopTask 停止任务（对等待中的任务生效）。
func (s *Service) StopTask(id string) error {
	return s.repo.UpdateTask(id, map[string]any{"status": domain.TaskStopped})
}

// StopExecution 停止一次执行：取消运行中的爬虫 goroutine，并把执行记录标记为已停止。
// 若该执行已无运行中的 goroutine（如服务重启后的残留 running），直接标记停止。
func (s *Service) StopExecution(id string) error {
	s.mu.Lock()
	if cancel, ok := s.execCancels[id]; ok {
		cancel()
	}
	s.mu.Unlock()
	return s.repo.UpdateExecution(id, map[string]any{
		"status": domain.TaskStopped, "progress": 100, "end_time": now(),
		"duration": "已手动停止",
	})
}

// DeleteExecution 删除执行记录；若仍在运行先取消其 goroutine。
func (s *Service) DeleteExecution(id string) error {
	s.mu.Lock()
	if cancel, ok := s.execCancels[id]; ok {
		cancel()
		delete(s.execCancels, id)
	}
	s.mu.Unlock()
	return s.repo.DeleteExecution(id)
}

// RecoverStaleExecutions 服务启动时清理残留状态：
// 上次进程退出/崩溃留下的 running 执行记录与任务全部标记为中断，避免永久卡在“运行中”。
func (s *Service) RecoverStaleExecutions() error {
	return s.repo.RecoverStale()
}

// ---------- 定时调度器 ----------
// 根据任务 schedule 字段自动触发执行。支持三种格式：
//   每天 HH:MM   （如 “每天 03:00”）—— 每天该时刻执行一次
//   每 N 小时     （如 “每 6 小时”）—— 距上次自动执行满 N 小时再执行
//   每小时         —— 每小时整点执行一次
// “手动”或不匹配的 schedule 不会被自动触发（后台仍可手动执行）。

var (
	dailyRe      = regexp.MustCompile(`每天\s*(\d{1,2}:\d{2})`)
	everyHoursRe = regexp.MustCompile(`每\s*(\d+)\s*小时`)
)

// StartScheduler 启动定时调度 goroutine，直到 ctx 取消。
// lastAuto 记录每个任务上次自动触发时间（进程内存态，重启后清空）。
func (s *Service) StartScheduler(ctx context.Context) {
	lastAuto := make(map[string]time.Time)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		log.Println("✅ 定时调度器已启动（每 30 秒检查一次任务计划）")
		for {
			select {
			case <-ctx.Done():
				log.Println("定时调度器已退出")
				return
			case <-ticker.C:
				s.runDueTasks(lastAuto, time.Now())
			}
		}
	}()
}

// runDueTasks 检查全部任务，对到期且未在运行的任务发起执行（异步）。
func (s *Service) runDueTasks(lastAuto map[string]time.Time, now time.Time) {
	tasks, err := s.repo.ListTasks()
	if err != nil {
		return
	}
	for _, t := range tasks {
		if t.Schedule == "" || strings.TrimSpace(t.Schedule) == "手动" {
			continue
		}
		// 正在运行的任务不重复触发
		if t.Status == domain.TaskRunning {
			continue
		}
		if !taskDue(t.Schedule, lastAuto[t.ID], now) {
			continue
		}
		lastAuto[t.ID] = now
		go func(id, name string) {
			if _, err := s.RunTask(id); err != nil {
				log.Printf("[scheduler] 触发任务 %s(%s) 失败: %v", id, name, err)
			} else {
				log.Printf("[scheduler] ✅ 已自动触发定时任务 %s(%s)", id, name)
			}
		}(t.ID, t.Name)
	}
}

// taskDue 判断任务是否到执行时间。
// lastAuto 为 0 值表示从未自动执行过（直接到期）；否则按间隔/当日判断。
func taskDue(schedule string, lastAuto time.Time, now time.Time) bool {
	sched := strings.TrimSpace(schedule)

	// 每天 HH:MM：当前时间匹配且当天未执行过
	if m := dailyRe.FindStringSubmatch(sched); len(m) == 2 {
		hhmm := strings.Split(m[1], ":")
		if len(hhmm) != 2 {
			return false
		}
		h, _ := strconv.Atoi(hhmm[0])
		min, _ := strconv.Atoi(hhmm[1])
		if now.Hour() == h && now.Minute() == min {
			return lastAuto.IsZero() || !sameDay(lastAuto, now)
		}
		return false
	}

	// 每 N 小时：距上次自动执行满 N 小时
	if m := everyHoursRe.FindStringSubmatch(sched); len(m) == 2 {
		n, _ := strconv.Atoi(m[1])
		if n <= 0 {
			return false
		}
		return lastAuto.IsZero() || now.Sub(lastAuto) >= time.Duration(n)*time.Hour
	}

	// 每小时：整点且该小时未执行过
	if sched == "每小时" {
		if now.Minute() != 0 {
			return false
		}
		return lastAuto.IsZero() || lastAuto.Hour() != now.Hour() || !sameDay(lastAuto, now)
	}
	return false
}

// sameDay 判断两个时间是否同一天。
func sameDay(a, b time.Time) bool {
	y1, m1, d1 := a.Date()
	y2, m2, d2 := b.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// executeTask 后台执行爬虫并更新记录与统计，同时通过 WebSocket 实时推送进度与日志。
func (s *Service) executeTask(ctx context.Context, client *crawler.Client, taskID, taskName, query, execID string) {
	start := time.Now()

	// 兜底：任何 panic 都不得让执行记录永久卡在 running
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[exec %s] panic: %v", execID, r)
			_ = s.repo.UpdateExecution(execID, map[string]any{
				"status": domain.TaskFailed, "progress": 100, "end_time": now(),
				"duration": time.Since(start).Round(time.Second).String(),
			})
			_ = s.repo.UpdateTask(taskID, map[string]any{"status": domain.TaskFailed, "fail_count": increment(s, taskID, "fail_count")})
		}
	}()

	// 进度：写库 + 广播
	progress := func(p int, step string) {
		_ = s.repo.UpdateExecution(execID, map[string]any{"progress": p})
		s.publishExec(execID, domain.ExecEvent{Type: "progress", ExecID: execID, Progress: p, Step: step})
	}
	// 追加日志：写库 + 广播
	appendLog := func(level, text string) {
		line := domain.LogLine{Time: time.Now().Format("15:04:05"), Level: level, Text: text}
		_ = s.repo.UpdateExecution(execID, map[string]any{"logs": append(getLogs(s, execID), line)})
		s.publishExec(execID, domain.ExecEvent{Type: "log", ExecID: execID, Log: &line})
	}
	// 收尾：写库最终状态 + 广播日志与状态；extra 可附带统计字段
	finish := func(status domain.TaskStatus, level, text string, extra ...map[string]any) {
		line := domain.LogLine{Time: time.Now().Format("15:04:05"), Level: level, Text: text}
		fields := map[string]any{
			"status": status, "end_time": now(), "progress": 100,
			"duration": time.Since(start).Round(time.Second).String(),
			"logs":     append(getLogs(s, execID), line),
		}
		if len(extra) > 0 {
			for k, v := range extra[0] {
				fields[k] = v
			}
		}
		_ = s.repo.UpdateExecution(execID, fields)
		s.publishExec(execID, domain.ExecEvent{Type: "log", ExecID: execID, Log: &line})
		s.publishExec(execID, domain.ExecEvent{Type: "status", ExecID: execID, Status: status, Progress: 100, Step: text})
	}

	appendLog("info", "开始抓取目标数据")
	// token 池诊断：确认任务运行时 client 是否有可用 token
	{
		toks := s.currentTokens()
		appendLog("info", fmt.Sprintf("GitHub Token 池：%d 个 token 可用，HasToken=%v", len(toks), client.HasToken()))
	}
	// 动态加载官方组织（official_orgs 表）并注入爬虫客户端，识别官方来源无需改代码
	var orgOwners []string
	if orgs, err := s.repo.ListOfficialOrgs(); err == nil && len(orgs) > 0 {
		clientOrgs := make([]crawler.OfficialOrg, 0, len(orgs))
		for _, o := range orgs {
			if !o.Enabled {
				continue
			}
			orgOwners = append(orgOwners, o.Owner)
			clientOrgs = append(clientOrgs, crawler.OfficialOrg{
				Owner: o.Owner, DisplayName: o.DisplayName, Avatar: o.Avatar,
			})
		}
		client.SetOfficialOrgs(clientOrgs)
		appendLog("info", fmt.Sprintf("已加载官方组织 %d 个（动态配置）", len(clientOrgs)))
	}
	progress(5, "读取爬虫配置")
	cfg, err := s.repo.GetConfig()
	if err != nil {
		cfg = domain.CrawlerConfig{}
	}
	// 爬虫总开关：配置中关闭时直接结束，不发起真实抓取
	if !cfg.Enabled {
		appendLog("warn", "爬虫总开关已关闭（爬虫配置 → 启用爬虫），本次任务已跳过")
		finish(domain.TaskFailed, "err", "任务未执行：爬虫总开关已关闭")
		return
	}
	// 单次后台执行限制搜索仓库数量，避免耗时过长
	limit := 15
	if cfg.MaxPagesPerRun > 0 && cfg.MaxPagesPerRun < limit {
		limit = cfg.MaxPagesPerRun
	}
	// 匿名模式（无 GITHUB_TOKEN）GitHub 限流仅 60 次/小时：把单次抓取量压到 6 个仓库以内，避免耗尽配额
	if !client.HasToken() && limit > 6 {
		limit = 6
	}

	repos := splitRepos(cfg.OfficialRepos)
	// 自动发现官方组织的技能仓库（名称含 skill/agent/mcp 优先，否则取高星仓库）。
	// 匿名模式配额不足（75 组织 × 列表 API 会瞬间打满 60 次/小时），跳过自动发现。
	if len(orgOwners) > 0 && client.HasToken() {
		discovered := discoverOfficialRepos(client, orgOwners)
		if len(discovered) > 0 {
			repos = append(repos, discovered...)
			appendLog("info", fmt.Sprintf("自动发现官方组织技能仓库 %d 个（与官方组织挂钩）", len(discovered)))
		}
	}
	opts := crawler.CrawlOptions{
		Query:   query,
		Repos:   repos,
		Limit:   limit,
		PerPage: 10,
		// 每个仓库处理完回调：进度映射到 10% ~ 80% 区间
		OnProgress: func(done, total int) {
			p := 10
			if total > 0 {
				p = 10 + int(float64(done)/float64(total)*70)
			}
			if p > 80 {
				p = 80
			}
			_ = s.repo.UpdateExecution(execID, map[string]any{"progress": p})
			s.publishExec(execID, domain.ExecEvent{Type: "progress", ExecID: execID, Progress: p, Step: fmt.Sprintf("爬取中 %d/%d", done, total)})
		},
	}
	appendLog("info", "开始抓取 GitHub 仓库")
	skills, failures, err := client.CrawlDetailed(opts)
	if err != nil {
		// 手动停止：爬虫客户端被取消，标记为已停止（而非失败）
		if ctx.Err() != nil || client.IsCancelled() {
			finish(domain.TaskStopped, "warn", "任务已被手动停止")
			_ = s.repo.UpdateTask(taskID, map[string]any{"status": domain.TaskWaiting})
			return
		}
		finish(domain.TaskFailed, "err", "任务执行失败："+err.Error())
		_ = s.repo.UpdateTask(taskID, map[string]any{"status": domain.TaskFailed, "fail_count": increment(s, taskID, "fail_count")})
		return
	}

	progress(85, "统计并写入数据库")

	// 统计与入库
	var stats domain.ExecutionStats
	stats.Pages = len(skills) + len(failures)
	stats.Fetched = len(skills)
	stats.Failed = len(failures)

	// 区分官方 / 社区（个人）来源
	officialN, communityN := 0, 0
	for _, sk := range skills {
		if sk.IsOfficial {
			officialN++
		} else {
			communityN++
		}
	}
	appendLog("info", fmt.Sprintf("本次抓取：官方 %d 个，社区/个人 %d 个", officialN, communityN))

	// 组装待入库候选（标题/描述翻译成中文，供官网中英文切换）
	appendLog("info", "翻译标题与描述为中文…")
	candidates := make([]InsertSkill, 0, len(skills))
	for _, sk := range skills {
		candidates = append(candidates, InsertSkill{
			ID: sk.ID, Name: sk.Name, NameZh: s.translator.Translate(sk.Name, "zh-CN"),
			Author: sk.Author, Description: sk.Description,
			DescriptionZh: s.translator.Translate(sk.Description, "zh-CN"),
			Category:      sk.Category, DownloadURL: sk.DownloadURL, IsOfficial: sk.IsOfficial,
			GithubURL: sk.GithubURL, GithubStars: sk.GithubStars, License: sk.License,
			SkillPath: sk.SkillPath,
			Tags:      sk.Tags, Content: toContentSections(sk.Content),
		})
	}

	// 批量判重：一次查出数据库中已存在的 ID 与同源同名指纹（避免重复入库）
	ids := make([]string, 0, len(candidates))
	sources := make([][3]string, 0, len(candidates))
	for _, sk := range candidates {
		ids = append(ids, sk.ID)
		sources = append(sources, [3]string{sk.Name, sk.Author, sk.GithubURL})
	}
	existIDs := map[string]bool{}
	existSources := map[string]bool{}
	if existIDs, err = s.repo.ListExistingSkillIDs(ids); err != nil {
		appendLog("warn", "技能判重查询失败: "+err.Error())
	}
	if existSources, err = s.repo.ListExistingSkillSources(sources); err != nil {
		appendLog("warn", "技能判重查询失败: "+err.Error())
	}

	// 过滤重复（批内 + 数据库已存在），只保留新增
	insert, dup := filterDuplicates(candidates, existIDs, existSources)
	stats.NewData = len(insert)
	stats.Duplicate = dup

	if len(insert) > 0 {
		if err := s.repo.InsertCrawledSkills(insert); err != nil {
			appendLog("warn", "写入数据失败: "+err.Error())
		} else {
			appendLog("ok", fmt.Sprintf("新增 %d 条数据（待审核），跳过重复 %d 条", len(insert), dup))
		}
	} else if dup > 0 {
		appendLog("info", fmt.Sprintf("本次无新增：%d 条均为重复数据，已跳过", dup))
	}

	// 失败任务落库
	for _, f := range failures {
		_ = s.repo.CreateFailure(&domain.FailureRecord{
			ID: newID("fail"), TaskID: taskID, TaskName: taskName,
			URL: f.Repo, Reason: f.Reason, Error: f.Error,
			RetryCount: cfg.RetryCount, FailedAt: now(),
		})
	}
	if len(failures) > 0 {
		appendLog("err", fmt.Sprintf("发现 %d 个失败页面", len(failures)))
	}

	finish(domain.TaskSuccess, "ok", fmt.Sprintf("任务执行完成，共 %d 条数据", len(skills)), map[string]any{
		"pages": stats.Pages, "fetched": stats.Fetched, "failed": stats.Failed,
		"new_data": stats.NewData, "duplicate": stats.Duplicate,
	})
	_ = s.repo.UpdateTask(taskID, map[string]any{
		"status": domain.TaskSuccess, "last_run_at": now()[:16],
		"last_duration": time.Since(start).Round(time.Second).String(),
		"run_count":     increment(s, taskID, "run_count"),
		"success_count": increment(s, taskID, "success_count"),
	})
}

// filterDuplicates 过滤重复技能，返回待入库列表与重复数量。
// 判重规则（任一命中即视为重复，不入库）：
//  1. 批内指纹重复（同 name + author + githubUrl 只保留第一条）
//  2. 数据库中已存在相同 ID（existIDs）
//  3. 数据库中已存在同源同名（existSources，name|author|githubUrl）
func filterDuplicates(skills []InsertSkill, existIDs, existSources map[string]bool) ([]InsertSkill, int) {
	if existIDs == nil {
		existIDs = map[string]bool{}
	}
	if existSources == nil {
		existSources = map[string]bool{}
	}
	seen := make(map[string]bool, len(skills))
	out := make([]InsertSkill, 0, len(skills))
	dup := 0
	for _, sk := range skills {
		sourceKey := sk.Name + "|" + sk.Author + "|" + sk.GithubURL
		if seen[sourceKey] || existIDs[sk.ID] || existSources[sourceKey] {
			dup++
			continue
		}
		seen[sourceKey] = true
		out = append(out, sk)
	}
	return out, dup
}

// getLogs 读取当前执行记录的日志。
func getLogs(s *Service, execID string) []domain.LogLine {
	e, err := s.repo.GetExecution(execID)
	if err != nil {
		return nil
	}
	return e.Logs
}

// increment 对任务某计数列 +1。
func increment(s *Service, taskID, column string) int {
	t, err := s.repo.GetTask(taskID)
	if err != nil {
		return 0
	}
	switch column {
	case "run_count":
		return t.RunCount + 1
	case "success_count":
		return t.SuccessCount + 1
	case "fail_count":
		return t.FailCount + 1
	}
	return 0
}

// ---------- 执行记录 ----------

func (s *Service) ListExecutions() ([]domain.ExecutionRecord, error) {
	return s.repo.ListExecutions()
}

func (s *Service) GetExecution(id string) (domain.ExecutionRecord, error) {
	return s.repo.GetExecution(id)
}

// ---------- 失败任务 ----------

func (s *Service) ListFailures() ([]domain.FailureRecord, error) {
	return s.repo.ListFailures()
}

// RetryFailure 真正重新执行失败任务：找到对应的爬虫任务并触发运行，
// 成功后再移除失败记录；任务不存在则仅删除记录。
func (s *Service) RetryFailure(id string) error {
	failures, err := s.repo.ListFailures()
	if err == nil {
		for _, f := range failures {
			if f.ID == id && f.TaskID != "" {
				if _, err := s.RunTask(f.TaskID); err == nil {
					_ = s.repo.DeleteFailure(id)
					return nil
				}
				_ = s.repo.DeleteFailure(id)
				return nil
			}
		}
	}
	return s.repo.DeleteFailure(id)
}

func (s *Service) IgnoreFailure(id string) error {
	return s.repo.DeleteFailure(id)
}

// ---------- 爬虫配置 ----------

func (s *Service) GetConfig() (domain.CrawlerConfig, error) {
	return s.repo.GetConfig()
}

func (s *Service) SaveConfig(c domain.CrawlerConfig) error {
	if err := s.repo.SaveConfig(c); err != nil {
		return err
	}
	// 配置变更后刷新 token 池，立即对后续爬虫生效
	s.RefreshTokenPool()
	return nil
}

// ---------- 翻译管理 ----------

// ScanUntranslated 扫描未汉化技能（标题或描述任一不含中文）。
// 返回列表 + 未汉化总数，供翻译页面展示。
func (s *Service) ScanUntranslated(limit int) ([]domain.TranslationItem, int, error) {
	items, err := s.repo.ListUntranslatedSkills(limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountUntranslatedSkills()
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// TranslateSkill 翻译单条技能：标题与描述分别翻译为中文并写库。
// 返回更新后的条目（含翻译前后状态）。
func (s *Service) TranslateSkill(id string) (domain.TranslationItem, error) {
	items, err := s.repo.ListUntranslatedSkills(0)
	if err != nil {
		return domain.TranslationItem{}, err
	}
	var target *domain.TranslationItem
	for i := range items {
		if items[i].ID == id {
			target = &items[i]
			break
		}
	}
	if target == nil {
		return domain.TranslationItem{}, fmt.Errorf("技能不存在或已汉化")
	}
	return s.translateAndSave(target)
}

// TranslateAllUntranslated 批量翻译所有未汉化技能（标题+描述），返回翻译成功条数。
func (s *Service) TranslateAllUntranslated() (int, error) {
	items, err := s.repo.ListUntranslatedSkills(0)
	if err != nil {
		return 0, err
	}
	done := 0
	for i := range items {
		it := items[i]
		// 标题与描述都已汉化的跳过（并发/重复扫描兜底）
		if containsCJK(it.NameZh) && containsCJK(it.DescriptionZh) {
			continue
		}
		updated, err := s.translateAndSave(&it)
		if err != nil {
			continue // 单条翻译失败不中断批量
		}
		// 只有翻译后实际产生中文的才算成功（品牌名/无需翻译的不计入）
		if containsCJK(updated.NameZh) || containsCJK(updated.DescriptionZh) {
			done++
		}
	}
	return done, nil
}

// translateAndSave 将条目标题与描述翻译为中文并写库（缺啥补啥，已汉化的字段跳过）。
// 品牌名/专有名词等翻译后仍是原文（不含中文）的字段，视为"无需翻译"并同样写库标记，
// 使后续扫描不再把这类条目列为待翻译。
func (s *Service) translateAndSave(it *domain.TranslationItem) (domain.TranslationItem, error) {
	nameZh := it.NameZh
	descZh := it.DescriptionZh
	var err error
	// 标题未汉化才翻译；已汉化保留原文，避免覆盖人工修正
	if !containsCJK(nameZh) && it.Name != "" {
		nameZh, err = s.translator.TranslateStrict(it.Name, "zh-CN")
		if err != nil {
			return *it, fmt.Errorf("标题翻译失败: %w", err)
		}
	}
	if !containsCJK(descZh) && it.Description != "" {
		descZh, err = s.translator.TranslateStrict(it.Description, "zh-CN")
		if err != nil {
			return *it, fmt.Errorf("描述翻译失败: %w", err)
		}
	}
	if nameZh == "" && descZh == "" {
		return *it, nil
	}
	if err := s.repo.UpdateSkillTranslation(it.ID, nameZh, descZh); err != nil {
		return *it, err
	}
	it.NameZh = nameZh
	it.DescriptionZh = descZh
	it.TitleTranslated = containsCJK(nameZh)
	it.DescTranslated = containsCJK(descZh)
	return *it, nil
}

// ---------- 翻译配置管理 ----------

// provider 中文名映射（后台展示）。
var providerNames = map[string]string{
	"tencent": "腾讯云",
	"baidu":   "百度",
	"google":  "Google",
	"deepl":   "DeepL",
}

// TranslateStatus 翻译器当前状态（后台管理页）。
type TranslateStatus struct {
	Providers    []string       `json:"providers"`    // 当前生效通道链（按优先级）
	Primary      string         `json:"primary"`      // 后台配置的主通道（'' = 环境变量默认）
	Configured   map[string]bool `json:"configured"`  // 各通道是否已配置密钥
	LastSuccess  string         `json:"lastSuccess"`  // 最近一次翻译成功的通道
	Enabled      bool           `json:"enabled"`
	ProviderName map[string]string `json:"providerName"` // 通道中文名
}

// applyTranslateProvider 启动时从 DB 读取后台配置的主翻译通道并应用到翻译器。
func (s *Service) applyTranslateProvider() {
	p, err := s.repo.GetTranslateProvider()
	if err != nil || p == "" {
		return // 未配置或读取失败，沿用环境变量默认
	}
	s.setTranslatorProvider(p)
}

// setTranslatorProvider 把后台主通道配置应用到翻译器（运行时立即生效）。
// provider=auto 时使用默认优先级（tencent>baidu>google>deepl）。
func (s *Service) setTranslatorProvider(p string) {
	p = strings.ToLower(strings.TrimSpace(p))
	switch p {
	case "", "auto":
		// 保持翻译器默认链（New 时已按环境变量+密钥过滤）
		return
	case "tencent", "baidu", "google", "deepl":
		order := []string{p}
		for _, q := range []string{"tencent", "baidu", "google", "deepl"} {
			if q != p {
				order = append(order, q)
			}
		}
		s.translator.SetProviders(order)
	default:
		return
	}
}

// GetTranslateStatus 返回翻译器当前状态（通道链 / 配置情况 / 最近成功通道）。
func (s *Service) GetTranslateStatus() TranslateStatus {
	p, _ := s.repo.GetTranslateProvider()
	last := s.translator.LastSuccess()
	return TranslateStatus{
		Providers:    s.translator.Providers(),
		Primary:      p,
		Configured:   s.translator.ProviderStatus(),
		LastSuccess:  last,
		Enabled:      s.translator.Enabled(),
		ProviderName: providerNames,
	}
}

// SetTranslateProvider 设置主翻译通道（写库 + 运行时生效）。
// provider: auto / tencent / baidu / google / deepl
func (s *Service) SetTranslateProvider(provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	switch provider {
	case "auto", "tencent", "baidu", "google", "deepl":
	default:
		return fmt.Errorf("无效的翻译通道: %s（可选 auto/tencent/baidu/google/deepl）", provider)
	}
	if err := s.repo.SaveTranslateProvider(provider); err != nil {
		return err
	}
	s.setTranslatorProvider(provider)
	return nil
}

// TestTranslateProvider 测试指定通道的翻译连通性（不改变熔断/缓存状态）。
// provider=all 时逐个测试全部已配置通道。返回各通道测试结果。
func (s *Service) TestTranslateProvider(provider string) ([]map[string]any, error) {
	configured := s.translator.ProviderStatus()
	var targets []string
	if provider == "all" || provider == "" {
		order := []string{"tencent", "baidu", "google", "deepl"}
		for _, p := range order {
			if configured[p] {
				targets = append(targets, p)
			}
		}
	} else {
		provider = strings.ToLower(strings.TrimSpace(provider))
		switch provider {
		case "tencent", "baidu", "google", "deepl":
			targets = []string{provider}
		default:
			return nil, fmt.Errorf("无效的翻译通道: %s", provider)
		}
	}

	testText := "Machine Learning Engineering is a collection of curated resources."
	results := make([]map[string]any, 0, len(targets))
	for _, p := range targets {
		res := map[string]any{"provider": p, "name": providerNames[p]}
		out, elapsed, err := s.translator.TestProvider(p, testText, "zh-CN")
		if err != nil {
			res["ok"] = false
			res["error"] = err.Error()
		} else {
			res["ok"] = true
			res["elapsed"] = elapsed.Round(time.Millisecond).String()
			res["output"] = out
		}
		results = append(results, res)
	}
	return results, nil
}

// ---------- 抓取数据（数据审核） ----------

func (s *Service) ListData(f DataFilter) ([]DataItem, error) {
	return s.repo.ListData(f)
}

func (s *Service) UpdateDataStatus(id, status string) error {
	if !isValidDataStatus(status) {
		return fmt.Errorf("无效的数据状态")
	}
	return s.repo.UpdateDataStatus(id, status)
}

// UpdateDataStatusBatch 批量更新数据状态（审核页全选后一键通过/忽略）。
func (s *Service) UpdateDataStatusBatch(ids []string, status string) error {
	if len(ids) == 0 {
		return fmt.Errorf("未选择任何数据")
	}
	if !isValidDataStatus(status) {
		return fmt.Errorf("无效的数据状态")
	}
	return s.repo.UpdateDataStatusBatch(ids, status)
}

// AutoAuditPending 机器人自动审核：内容完整规范且无重复的直接通过，有问题的留人工。
func (s *Service) AutoAuditPending() (AutoAuditResult, error) {
	return s.repo.AutoAuditPending()
}

// PublishAllApproved 一键发布全部已审核（approved）数据到官网。
func (s *Service) PublishAllApproved() (int64, error) {
	return s.repo.PublishAllApproved()
}

func (s *Service) DeleteData(id string) error {
	return s.repo.DeleteData(id)
}

func isValidDataStatus(s string) bool {
	switch s {
	case string(domain.DataPending), string(domain.DataApproved), string(domain.DataPublished), string(domain.DataIgnored):
		return true
	}
	return false
}

// ---------- 内容 ----------

func (s *Service) ListArticles() ([]domain.Article, error) {
	return s.repo.ListArticles()
}

func (s *Service) GetArticle(id string) (domain.Article, error) {
	return s.repo.GetArticle(id)
}

func (s *Service) CreateArticle(input domain.Article) (domain.Article, error) {
	status := input.Status
	if status != "published" {
		status = "draft"
	}
	a := domain.Article{
		ID: newID("art"), Title: input.Title, Status: status,
		Category: input.Category, Author: "admin", UpdatedAt: today(),
		Content: input.Content,
	}
	if a.Category == "" {
		a.Category = "教程"
	}
	return a, s.repo.CreateArticle(&a)
}

func (s *Service) UpdateArticle(id string, input domain.Article) (domain.Article, error) {
	if id == "" {
		return domain.Article{}, fmt.Errorf("文章 ID 为空")
	}
	a := domain.Article{
		ID:        id,
		Title:     input.Title,
		Status:    input.Status,
		Category:  input.Category,
		Author:    input.Author,
		UpdatedAt: today(),
		Content:   input.Content,
	}
	if a.Category == "" {
		a.Category = "教程"
	}
	if a.Status == "" {
		a.Status = "draft"
	}
	if a.Author == "" {
		a.Author = "admin"
	}
	return a, s.repo.UpdateArticle(id, &a)
}

func (s *Service) DeleteArticle(id string) error {
	return s.repo.DeleteArticle(id)
}

// ---------- 赞助商 ----------

func (s *Service) ListSponsors() ([]domain.Sponsor, error) {
	return s.repo.ListSponsors()
}

func (s *Service) CreateSponsor(input domain.Sponsor) (domain.Sponsor, error) {
	sp := domain.Sponsor{
		ID:            newID("spn"),
		Name:          input.Name,
		Logo:          input.Logo,
		DescriptionZh: input.DescriptionZh,
		DescriptionEn: input.DescriptionEn,
		URL:           input.URL,
		Position:      normalizePosition(input.Position),
		Enabled:       input.Enabled,
		SortOrder:     input.SortOrder,
	}
	return sp, s.repo.CreateSponsor(&sp)
}

func (s *Service) UpdateSponsor(id string, input domain.Sponsor) (domain.Sponsor, error) {
	if id == "" {
		return domain.Sponsor{}, fmt.Errorf("赞助商 ID 为空")
	}
	sp := domain.Sponsor{
		ID:            id,
		Name:          input.Name,
		Logo:          input.Logo,
		DescriptionZh: input.DescriptionZh,
		DescriptionEn: input.DescriptionEn,
		URL:           input.URL,
		Position:      normalizePosition(input.Position),
		Enabled:       input.Enabled,
		SortOrder:     input.SortOrder,
	}
	return sp, s.repo.UpdateSponsor(id, &sp)
}

func (s *Service) DeleteSponsor(id string) error {
	return s.repo.DeleteSponsor(id)
}

func (s *Service) IncrSponsorClicks(id string) error {
	return s.repo.IncrSponsorClicks(id)
}

// normalizePosition 归一化展示位置（home / sidebar / both，非法值回退 home）。
func normalizePosition(p string) string {
	switch p {
	case "sidebar", "both":
		return p
	default:
		return "home"
	}
}

func (s *Service) GetSeo() (domain.SeoConfig, error)         { return s.repo.GetSeo() }
func (s *Service) SaveSeo(c domain.SeoConfig) error          { return s.repo.SaveSeo(c) }
func (s *Service) GetSiteConfig() (domain.SiteConfig, error) { return s.repo.GetSiteConfig() }
func (s *Service) SaveSiteConfig(c domain.SiteConfig) error  { return s.repo.SaveSiteConfig(c) }

// ---------- 工作台统计 ----------

func (s *Service) Stats() (domain.AdminStats, error) {
	return s.repo.Stats()
}

func splitRepos(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// discoverOfficialRepos 为每个官方组织查找其技能仓库（并发，限流），返回 fullName 列表。
// 两轮筛选：①按 star 排行取名称含 skill/agent/mcp 的候选仓库；②只保留具备 skill 结构
// （根 SKILL.md 或 skills/skillsets 目录）的仓库，避免爬 agent/mcp 代码仓库浪费资源。
// 总上限 15 个，去重。使爬虫能主动抓取官方组织的技能仓库，爬出的技能自动标记为官方。
func discoverOfficialRepos(client *crawler.Client, owners []string) []string {
	// 第一轮：候选仓库（每个官方组织 star 最高的技能类仓库）
	var candidates []string
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)
	for _, owner := range owners {
		wg.Add(1)
		sem <- struct{}{}
		go func(o string) {
			defer wg.Done()
			defer func() { <-sem }()
			repos, err := client.ListOrgRepos(o, 10)
			if err != nil || len(repos) == 0 {
				return
			}
			picked := ""
			bestStars := -1
			for _, r := range repos {
				n := strings.ToLower(r.FullName)
				if strings.Contains(n, "skill") || strings.Contains(n, "agent") || strings.Contains(n, "mcp") {
					if r.Stars > bestStars {
						picked = r.FullName
						bestStars = r.Stars
					}
				}
			}
			if picked != "" {
				mu.Lock()
				candidates = append(candidates, picked)
				mu.Unlock()
			}
		}(owner)
	}
	wg.Wait()

	// 第二轮：只保留具备 skill 结构的仓库（能真正产出技能）
	var out []string
	var mu2 sync.Mutex
	var wg2 sync.WaitGroup
	sem2 := make(chan struct{}, 5)
	for _, fullName := range candidates {
		wg2.Add(1)
		sem2 <- struct{}{}
		go func(fn string) {
			defer wg2.Done()
			defer func() { <-sem2 }()
			if !client.HasSkillStructure(fn) {
				return
			}
			mu2.Lock()
			if len(out) < 15 {
				out = append(out, fn)
			}
			mu2.Unlock()
		}(fullName)
	}
	wg2.Wait()
	return out
}

// toContentSections 将爬虫内容区块转换为领域模型。
func toContentSections(in []crawler.ContentSection) []domain.ContentSection {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.ContentSection, 0, len(in))
	for _, s := range in {
		out = append(out, domain.ContentSection{Heading: s.Heading, Body: s.Body})
	}
	return out
}
