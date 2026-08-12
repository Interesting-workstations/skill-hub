// Package admin 管理后台业务层：任务/执行/失败/配置/内容/认证。
package admin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/crawler"
	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
)

// Service 后台管理业务逻辑。
type Service struct {
	repo   Repository
	client *crawler.Client
	token  []byte // 进程内签名密钥（重启后旧 token 失效）
	mu     sync.Mutex
}

// NewService 创建后台管理服务。
func NewService(repo Repository) *Service {
	secret := make([]byte, 32)
	_, _ = rand.Read(secret)
	return &Service{
		repo:   repo,
		client: crawler.NewClient(os.Getenv("GITHUB_TOKEN")),
		token:  secret,
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

// Login 校验管理员账号，返回 token 与用户信息。
func (s *Service) Login(username, password string) (domain.AdminUser, string, error) {
	rec, err := s.repo.GetAdmin(username)
	if err != nil {
		return domain.AdminUser{}, "", fmt.Errorf("用户名或密码错误")
	}
	if rec.PasswordHash != HashPassword(password) {
		return domain.AdminUser{}, "", fmt.Errorf("用户名或密码错误")
	}
	expiry := time.Now().Add(24 * time.Hour).Unix()
	mac := hmac.New(sha256.New, s.token)
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d", username, expiry)))
	sig := hex.EncodeToString(mac.Sum(nil))
	token := fmt.Sprintf("%s.%d.%s", username, expiry, sig)
	return domain.AdminUser{Username: rec.Username, DisplayName: rec.DisplayName, Role: "admin"}, token, nil
}

// VerifyToken 校验 token，返回用户名。
func (s *Service) VerifyToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", false
	}
	username, expiryStr, sig := parts[0], parts[1], parts[2]
	var expiry int64
	if _, err := fmt.Sscanf(expiryStr, "%d", &expiry); err != nil {
		return "", false
	}
	if time.Now().Unix() > expiry {
		return "", false
	}
	mac := hmac.New(sha256.New, s.token)
	_, _ = mac.Write([]byte(fmt.Sprintf("%s:%d", username, expiry)))
	if !hmac.Equal([]byte(sig), []byte(hex.EncodeToString(mac.Sum(nil)))) {
		return "", false
	}
	return username, true
}

// ChangePassword 修改管理员密码。
func (s *Service) ChangePassword(username, oldPwd, newPwd string) error {
	rec, err := s.repo.GetAdmin(username)
	if err != nil || rec.PasswordHash != HashPassword(oldPwd) {
		return fmt.Errorf("当前密码错误")
	}
	return s.repo.UpdateAdminPassword(username, HashPassword(newPwd))
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

// executeTask 后台执行爬虫并更新记录与统计。
func (s *Service) executeTask(taskID, taskName, query, execID string) {
	start := time.Now()
	appendLog := func(level, text string) {
		_ = s.repo.UpdateExecution(execID, map[string]any{
			"logs": append(getLogs(s, execID), domain.LogLine{Time: time.Now().Format("15:04:05"), Level: level, Text: text}),
		})
	}

	appendLog("info", "开始抓取目标数据")
	cfg, err := s.repo.GetConfig()
	if err != nil {
		cfg = domain.CrawlerConfig{}
	}
	// 单次后台执行限制搜索仓库数量，避免耗时过长
	limit := 15
	if cfg.MaxPagesPerRun > 0 && cfg.MaxPagesPerRun < limit {
		limit = cfg.MaxPagesPerRun
	}

	repos := splitRepos(cfg.OfficialRepos)
	opts := crawler.CrawlOptions{
		Query:   query,
		Repos:   repos,
		Limit:   limit,
		PerPage: 10,
	}
	skills, failures, err := s.client.CrawlDetailed(opts)
	if err != nil {
		_ = s.repo.UpdateExecution(execID, map[string]any{
			"status": domain.TaskFailed, "end_time": now(), "progress": 100,
			"logs": append(getLogs(s, execID), domain.LogLine{Time: time.Now().Format("15:04:05"), Level: "err", Text: err.Error()}),
		})
		_ = s.repo.UpdateTask(taskID, map[string]any{"status": domain.TaskFailed, "fail_count": increment(s, taskID, "fail_count")})
		return
	}

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

	var insert []InsertSkill
	for _, sk := range skills {
		exists, _ := s.repo.SkillExists(sk.ID)
		if exists {
			stats.Duplicate++
			continue
		}
		stats.NewData++
		insert = append(insert, InsertSkill{
			ID: sk.ID, Name: sk.Name, Author: sk.Author, Description: sk.Description,
			Category: sk.Category, DownloadURL: sk.DownloadURL, IsOfficial: sk.IsOfficial,
			GithubURL: sk.GithubURL, GithubStars: sk.GithubStars, License: sk.License, Tags: sk.Tags,
		})
	}
	if len(insert) > 0 {
		if err := s.repo.InsertCrawledSkills(insert); err != nil {
			appendLog("warn", "写入数据失败: "+err.Error())
		} else {
			appendLog("ok", fmt.Sprintf("新增 %d 条数据（待审核）", len(insert)))
		}
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

	appendLog("ok", fmt.Sprintf("任务执行完成，共 %d 条数据", len(skills)))
	_ = s.repo.UpdateExecution(execID, map[string]any{
		"status": domain.TaskSuccess, "end_time": now(), "progress": 100,
		"duration": time.Since(start).Round(time.Second).String(),
		"pages":    stats.Pages, "fetched": stats.Fetched, "failed": stats.Failed,
		"new_data": stats.NewData, "duplicate": stats.Duplicate,
	})
	_ = s.repo.UpdateTask(taskID, map[string]any{
		"status": domain.TaskSuccess, "last_run_at": now()[:16],
		"last_duration": time.Since(start).Round(time.Second).String(),
		"run_count":     increment(s, taskID, "run_count"),
		"success_count": increment(s, taskID, "success_count"),
	})
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

func (s *Service) RetryFailure(id string) error {
	// 简化：重试标记为已处理（删除记录，任务列表可重新执行）
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

func (s *Service) ListData(status string) ([]DataItem, error) {
	return s.repo.ListData(status)
}

func (s *Service) UpdateDataStatus(id, status string) error {
	if !isValidDataStatus(status) {
		return fmt.Errorf("无效的数据状态")
	}
	return s.repo.UpdateDataStatus(id, status)
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

func (s *Service) CreateArticle(input domain.Article) (domain.Article, error) {
	a := domain.Article{
		ID: newID("art"), Title: input.Title, Status: "draft",
		Category: input.Category, Author: "admin", UpdatedAt: today(),
	}
	if a.Category == "" {
		a.Category = "教程"
	}
	return a, s.repo.CreateArticle(&a)
}

func (s *Service) DeleteArticle(id string) error {
	return s.repo.DeleteArticle(id)
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
