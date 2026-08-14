// Package admin 管理后台业务层：任务/执行/失败/配置/内容/认证。
package admin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
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

	blacklist     map[string]int64        // access token hash → 过期时间（主动登出）
	refreshTokens map[string]refreshEntry // refresh token hash → 条目
	tokenVersions map[string]int          // username → token version（改密后 +1 全失效）
	loginAttempts map[string][]int64      // ip → 登录尝试时间戳（滑动窗口限流）

	execSubs  map[string]map[chan domain.ExecEvent]struct{} // execID → 订阅者（执行进度/日志推送）
	wsTickets map[string]wsTicket                           // ticket → 一次性 WS 票据
}

// NewService 创建后台管理服务。
func NewService(repo Repository) *Service {
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)
	return &Service{
		repo:          repo,
		client:        crawler.NewClient(os.Getenv("GITHUB_TOKEN")),
		token:         secret,
		blacklist:     make(map[string]int64),
		refreshTokens: make(map[string]refreshEntry),
		tokenVersions: make(map[string]int),
		loginAttempts: make(map[string][]int64),
		execSubs:      make(map[string]map[chan domain.ExecEvent]struct{}),
		wsTickets:     make(map[string]wsTicket),
	}
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

	go s.executeTask(task.ID, task.Name, task.Query, record.ID)
	return record, nil
}

// StopTask 停止任务（对等待中的任务生效）。
func (s *Service) StopTask(id string) error {
	return s.repo.UpdateTask(id, map[string]any{"status": domain.TaskStopped})
}

// executeTask 后台执行爬虫并更新记录与统计，同时通过 WebSocket 实时推送进度与日志。
func (s *Service) executeTask(taskID, taskName, query, execID string) {
	start := time.Now()

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
		s.client.SetOfficialOrgs(clientOrgs)
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

	repos := splitRepos(cfg.OfficialRepos)
	// 自动发现官方组织的技能仓库，与官方组织挂钩（名称含 skill/agent/mcp 优先，否则取高星仓库）
	if len(orgOwners) > 0 {
		discovered := discoverOfficialRepos(s.client, orgOwners)
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
	skills, failures, err := s.client.CrawlDetailed(opts)
	if err != nil {
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

	// 组装待入库候选
	candidates := make([]InsertSkill, 0, len(skills))
	for _, sk := range skills {
		candidates = append(candidates, InsertSkill{
			ID: sk.ID, Name: sk.Name, Author: sk.Author, Description: sk.Description,
			Category: sk.Category, DownloadURL: sk.DownloadURL, IsOfficial: sk.IsOfficial,
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
	return s.repo.SaveConfig(c)
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
// 只选名称含 skill/agent/mcp 的仓库（取 star 最高者），无技能仓库的组织跳过；
// 总上限 15 个，去重。使爬虫能主动抓取官方组织的技能仓库，爬出的技能自动标记为官方（与官方组织挂钩）。
func discoverOfficialRepos(client *crawler.Client, owners []string) []string {
	seen := map[string]bool{}
	var out []string
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
			if picked == "" {
				return // 该组织没有明显的技能仓库，跳过
			}
			mu.Lock()
			if !seen[picked] && len(out) < 20 {
				seen[picked] = true
				out = append(out, picked)
			}
			mu.Unlock()
		}(owner)
	}
	wg.Wait()
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
