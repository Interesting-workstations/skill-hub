// Package admin 是管理后台业务模块（Handler → Service → Repository）。
// 数据全部持久化到 MySQL：爬虫任务、执行记录、失败任务、配置、内容、认证。
package admin

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/go-sql-driver/mysql"
)

// DataItem 抓取数据（含审核状态）。
type DataItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Category    string `json:"category"`
	GithubStars string `json:"githubStars"`
	IsOfficial  bool   `json:"isOfficial"`
	Source      string `json:"source"`
	FetchedAt   string `json:"fetchedAt"`
	Status      string `json:"status"`
}

// Repository 定义后台管理数据访问接口。
type Repository interface {
	ListTasks() ([]domain.CrawlTask, error)
	GetTask(id string) (domain.CrawlTask, error)
	CreateTask(t *domain.CrawlTask) error
	UpdateTask(id string, fields map[string]any) error
	DeleteTask(id string) error

	ListExecutions() ([]domain.ExecutionRecord, error)
	GetExecution(id string) (domain.ExecutionRecord, error)
	CreateExecution(e *domain.ExecutionRecord) error
	UpdateExecution(id string, fields map[string]any) error

	ListFailures() ([]domain.FailureRecord, error)
	CreateFailure(f *domain.FailureRecord) error
	DeleteFailure(id string) error

	GetConfig() (domain.CrawlerConfig, error)
	SaveConfig(c domain.CrawlerConfig) error

	ListData(status string) ([]DataItem, error)
	UpdateDataStatus(id, status string) error
	DeleteData(id string) error

	ListArticles() ([]domain.Article, error)
	CreateArticle(a *domain.Article) error
	DeleteArticle(id string) error

	GetSeo() (domain.SeoConfig, error)
	SaveSeo(s domain.SeoConfig) error

	GetSiteConfig() (domain.SiteConfig, error)
	SaveSiteConfig(s domain.SiteConfig) error

	GetAdmin(username string) (AdminRecord, error)
	UpdateAdminPassword(username, hash string) error

	SkillExists(id string) (bool, error)
	InsertCrawledSkills(skills []InsertSkill) error
	CountSkills() (int, error)
	CountOfficial() (int, error)
	CountAuthors() (int, error)
	Stats() (domain.AdminStats, error)
}

// AdminRecord 管理员账号记录（含密码哈希）。
type AdminRecord struct {
	Username     string
	PasswordHash string
	DisplayName  string
}

// InsertSkill 爬虫结果写入 skills 表所需的字段。
type InsertSkill struct {
	ID          string
	Name        string
	Author      string
	Description string
	Category    string
	DownloadURL string
	IsOfficial  bool
	GithubURL   string
	GithubStars string
	License     string
	Tags        []string
	Content     []domain.ContentSection
}

type mysqlRepo struct {
	db *sql.DB
}

// NewMySQLRepository 连接 MySQL 并确保后台表结构就绪。
func NewMySQLRepository(dsn string) (Repository, error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析 DSN 失败: %w", err)
	}
	if cfg.DBName == "" {
		return nil, fmt.Errorf("DSN 必须指定数据库名")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	repo := &mysqlRepo{db: db}
	if err := repo.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := repo.seed(); err != nil {
		db.Close()
		return nil, err
	}
	return repo, nil
}

// migrate 创建后台相关数据表，并给 skills 表补充审核状态列。
func (r *mysqlRepo) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
			username VARCHAR(64) PRIMARY KEY,
			password_hash VARCHAR(128) NOT NULL,
			display_name VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS crawl_tasks (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			type VARCHAR(32) NOT NULL,
			query TEXT NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'waiting',
			schedule VARCHAR(64) NOT NULL DEFAULT '手动',
			last_run_at VARCHAR(32) NOT NULL DEFAULT '',
			last_duration VARCHAR(32) NOT NULL DEFAULT '',
			run_count INT NOT NULL DEFAULT 0,
			success_count INT NOT NULL DEFAULT 0,
			fail_count INT NOT NULL DEFAULT 0,
			created_at VARCHAR(16) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS crawl_executions (
			id VARCHAR(64) PRIMARY KEY,
			task_id VARCHAR(64) NOT NULL,
			task_name VARCHAR(128) NOT NULL,
			status VARCHAR(16) NOT NULL,
			start_time VARCHAR(32) NOT NULL,
			end_time VARCHAR(32) NOT NULL DEFAULT '',
			duration VARCHAR(32) NOT NULL DEFAULT '',
			progress INT NOT NULL DEFAULT 0,
			pages INT NOT NULL DEFAULT 0,
			fetched INT NOT NULL DEFAULT 0,
			failed INT NOT NULL DEFAULT 0,
			new_data INT NOT NULL DEFAULT 0,
			updated INT NOT NULL DEFAULT 0,
			duplicate INT NOT NULL DEFAULT 0,
			logs MEDIUMTEXT NOT NULL,
			KEY idx_exec_task (task_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS crawl_failures (
			id VARCHAR(64) PRIMARY KEY,
			task_id VARCHAR(64) NOT NULL,
			task_name VARCHAR(128) NOT NULL,
			url VARCHAR(500) NOT NULL,
			reason VARCHAR(128) NOT NULL,
			error TEXT NOT NULL,
			retry_count INT NOT NULL DEFAULT 0,
			failed_at VARCHAR(32) NOT NULL,
			KEY idx_fail_task (task_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS crawler_config (
			id TINYINT PRIMARY KEY,
			concurrency INT NOT NULL DEFAULT 4,
			timeout INT NOT NULL DEFAULT 20,
			retry_count INT NOT NULL DEFAULT 3,
			user_agent VARCHAR(255) NOT NULL DEFAULT '',
			request_interval INT NOT NULL DEFAULT 500,
			max_pages INT NOT NULL DEFAULT 500,
			official_repos TEXT NOT NULL,
			default_query VARCHAR(128) NOT NULL DEFAULT 'claude skills',
			enabled TINYINT(1) NOT NULL DEFAULT 1
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS articles (
			id VARCHAR(64) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'draft',
			category VARCHAR(64) NOT NULL DEFAULT '教程',
			author VARCHAR(64) NOT NULL DEFAULT 'admin',
			views INT NOT NULL DEFAULT 0,
			updated_at VARCHAR(16) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS seo_config (
			id TINYINT PRIMARY KEY,
			title VARCHAR(255) NOT NULL DEFAULT '',
			description TEXT NOT NULL,
			keywords VARCHAR(255) NOT NULL DEFAULT '',
			og_image VARCHAR(500) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS site_config (
			id TINYINT PRIMARY KEY,
			site_name VARCHAR(128) NOT NULL DEFAULT '',
			slogan VARCHAR(255) NOT NULL DEFAULT '',
			portal_url VARCHAR(255) NOT NULL DEFAULT '',
			icp VARCHAR(64) NOT NULL DEFAULT '',
			contact_email VARCHAR(128) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
	}
	return r.ensureDataStatusColumn()
}

// ensureDataStatusColumn 给 skills 表补充 data_status 列（已存在则忽略）。
func (r *mysqlRepo) ensureDataStatusColumn() error {
	_, err := r.db.Exec(`ALTER TABLE skills ADD COLUMN data_status VARCHAR(20) NOT NULL DEFAULT 'published'`)
	if err != nil {
		// 列已存在或表不存在均忽略（首次创建 skills 表时由 skill 模块负责）
		if strings.Contains(err.Error(), "Duplicate column") || strings.Contains(err.Error(), "check that column/key exists") {
			return nil
		}
		return err
	}
	return nil
}

// seed 初始化后台种子数据（管理员、爬虫配置、SEO、站点、示例任务）。
func (r *mysqlRepo) seed() error {
	// 管理员
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM admin_users`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if _, err := r.db.Exec(
			`INSERT INTO admin_users(username, password_hash, display_name) VALUES(?,?,?)`,
			"admin", HashPassword("admin123"), "管理员",
		); err != nil {
			return err
		}
	}

	// 爬虫配置（单行）
	var cfgCount int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM crawler_config`).Scan(&cfgCount); err != nil {
		return err
	}
	if cfgCount == 0 {
		if _, err := r.db.Exec(
			`INSERT INTO crawler_config(id, concurrency, timeout, retry_count, user_agent,
				request_interval, max_pages, official_repos, default_query, enabled)
			 VALUES(1, 4, 20, 3, 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) SkillHubBot/1.0',
				500, 500, 'anthropics/skills,openai/codex,vercel/ai', 'claude skills', 1)`,
		); err != nil {
			return err
		}
	}

	// SEO / 站点配置
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM seo_config`).Scan(&n); err == nil && n == 0 {
		if _, err := r.db.Exec(
			`INSERT INTO seo_config(id, title, description, keywords, og_image) VALUES(1,?,?,?,?)`,
			"Agent Skills 资源库 — AI 编程助手的可复用技能",
			"发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能。",
			"agent skills, claude code, AI 编程助手, 技能库",
			"https://example.com/og-cover.png",
		); err != nil {
			return err
		}
	}
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM site_config`).Scan(&n); err == nil && n == 0 {
		if _, err := r.db.Exec(
			`INSERT INTO site_config(id, site_name, slogan, portal_url, icp, contact_email) VALUES(1,?,?,?,?,?)`,
			"Agent Skills 资源库", "AI 编程助手的可复用技能",
			"http://localhost:5173", "京ICP备00000000号", "admin@example.com",
		); err != nil {
			return err
		}
	}

	// 示例爬虫任务
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM crawl_tasks`).Scan(&n); err == nil && n == 0 {
		tasks := []domain.CrawlTask{
			{ID: "task-1", Name: "官方技能采集", Type: "skill", Query: "anthropics/skills, openai/codex", Status: domain.TaskSuccess, Schedule: "每天 02:00", LastRunAt: "", RunCount: 26, SuccessCount: 24, FailCount: 2, CreatedAt: "2026-07-01"},
			{ID: "task-2", Name: "社区技能搜索", Type: "skill", Query: "claude skills", Status: domain.TaskSuccess, Schedule: "每小时", LastRunAt: "", RunCount: 132, SuccessCount: 128, FailCount: 4, CreatedAt: "2026-07-05"},
			{ID: "task-3", Name: "星标榜热更", Type: "info", Query: "agent skills stars:>100", Status: domain.TaskFailed, Schedule: "每 6 小时", LastRunAt: "", RunCount: 58, SuccessCount: 51, FailCount: 7, CreatedAt: "2026-07-10"},
			{ID: "task-4", Name: "新分类探测", Type: "data", Query: "skillsets", Status: domain.TaskWaiting, Schedule: "手动", RunCount: 0, CreatedAt: "2026-08-11"},
		}
		for _, t := range tasks {
			if _, err := r.db.Exec(
				`INSERT INTO crawl_tasks(id, name, type, query, status, schedule, run_count, success_count, fail_count, created_at)
				 VALUES(?,?,?,?,?,?,?,?,?,?)`,
				t.ID, t.Name, t.Type, t.Query, t.Status, t.Schedule, t.RunCount, t.SuccessCount, t.FailCount, t.CreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// HashPassword 生成密码哈希（SHA-256 + 固定盐，单管理员场景足够安全）。
func HashPassword(pwd string) string {
	sum := sha256.Sum256([]byte("skillhub-admin-salt:" + pwd))
	return hex.EncodeToString(sum[:])
}

// ---------- 爬虫任务 ----------

func (r *mysqlRepo) ListTasks() ([]domain.CrawlTask, error) {
	rows, err := r.db.Query(`SELECT id, name, type, query, status, schedule,
		last_run_at, last_duration, run_count, success_count, fail_count, created_at
		FROM crawl_tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CrawlTask, 0)
	for rows.Next() {
		var t domain.CrawlTask
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Query, &t.Status, &t.Schedule,
			&t.LastRunAt, &t.LastDuration, &t.RunCount, &t.SuccessCount, &t.FailCount, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetTask(id string) (domain.CrawlTask, error) {
	var t domain.CrawlTask
	err := r.db.QueryRow(`SELECT id, name, type, query, status, schedule,
		last_run_at, last_duration, run_count, success_count, fail_count, created_at
		FROM crawl_tasks WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Type, &t.Query, &t.Status, &t.Schedule,
		&t.LastRunAt, &t.LastDuration, &t.RunCount, &t.SuccessCount, &t.FailCount, &t.CreatedAt)
	return t, err
}

func (r *mysqlRepo) CreateTask(t *domain.CrawlTask) error {
	_, err := r.db.Exec(
		`INSERT INTO crawl_tasks(id, name, type, query, status, schedule, created_at)
		 VALUES(?,?,?,?,?,?,?)`,
		t.ID, t.Name, t.Type, t.Query, t.Status, t.Schedule, t.CreatedAt)
	return err
}

func (r *mysqlRepo) UpdateTask(id string, fields map[string]any) error {
	return r.update("crawl_tasks", id, fields)
}

func (r *mysqlRepo) DeleteTask(id string) error {
	_, err := r.db.Exec(`DELETE FROM crawl_tasks WHERE id = ?`, id)
	return err
}

// ---------- 执行记录 ----------

func (r *mysqlRepo) ListExecutions() ([]domain.ExecutionRecord, error) {
	rows, err := r.db.Query(`SELECT id, task_id, task_name, status, start_time, end_time,
		duration, progress, pages, fetched, failed, new_data, updated, duplicate, logs
		FROM crawl_executions ORDER BY start_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.ExecutionRecord, 0)
	for rows.Next() {
		var e domain.ExecutionRecord
		var logsRaw string
		if err := rows.Scan(&e.ID, &e.TaskID, &e.TaskName, &e.Status, &e.StartTime, &e.EndTime,
			&e.Duration, &e.Progress, &e.Stats.Pages, &e.Stats.Fetched, &e.Stats.Failed,
			&e.Stats.NewData, &e.Stats.Updated, &e.Stats.Duplicate, &logsRaw); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(logsRaw), &e.Logs)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetExecution(id string) (domain.ExecutionRecord, error) {
	var e domain.ExecutionRecord
	var logsRaw string
	err := r.db.QueryRow(`SELECT id, task_id, task_name, status, start_time, end_time,
		duration, progress, pages, fetched, failed, new_data, updated, duplicate, logs
		FROM crawl_executions WHERE id = ?`, id).Scan(
		&e.ID, &e.TaskID, &e.TaskName, &e.Status, &e.StartTime, &e.EndTime,
		&e.Duration, &e.Progress, &e.Stats.Pages, &e.Stats.Fetched, &e.Stats.Failed,
		&e.Stats.NewData, &e.Stats.Updated, &e.Stats.Duplicate, &logsRaw)
	if err != nil {
		return e, err
	}
	_ = json.Unmarshal([]byte(logsRaw), &e.Logs)
	return e, nil
}

func (r *mysqlRepo) CreateExecution(e *domain.ExecutionRecord) error {
	logs, err := json.Marshal(e.Logs)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(
		`INSERT INTO crawl_executions(id, task_id, task_name, status, start_time, end_time,
			duration, progress, pages, fetched, failed, new_data, updated, duplicate, logs)
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.TaskID, e.TaskName, e.Status, e.StartTime, e.EndTime,
		e.Duration, e.Progress, e.Stats.Pages, e.Stats.Fetched, e.Stats.Failed,
		e.Stats.NewData, e.Stats.Updated, e.Stats.Duplicate, string(logs))
	return err
}

func (r *mysqlRepo) UpdateExecution(id string, fields map[string]any) error {
	if logs, ok := fields["logs"]; ok {
		if l, ok2 := logs.([]domain.LogLine); ok2 {
			b, err := json.Marshal(l)
			if err != nil {
				return err
			}
			fields["logs"] = string(b)
		}
	}
	return r.update("crawl_executions", id, fields)
}

// ---------- 失败任务 ----------

func (r *mysqlRepo) ListFailures() ([]domain.FailureRecord, error) {
	rows, err := r.db.Query(`SELECT id, task_id, task_name, url, reason, error, retry_count, failed_at
		FROM crawl_failures ORDER BY failed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.FailureRecord, 0)
	for rows.Next() {
		var f domain.FailureRecord
		if err := rows.Scan(&f.ID, &f.TaskID, &f.TaskName, &f.URL, &f.Reason, &f.Error, &f.RetryCount, &f.FailedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateFailure(f *domain.FailureRecord) error {
	_, err := r.db.Exec(
		`INSERT INTO crawl_failures(id, task_id, task_name, url, reason, error, retry_count, failed_at)
		 VALUES(?,?,?,?,?,?,?,?)`,
		f.ID, f.TaskID, f.TaskName, f.URL, f.Reason, f.Error, f.RetryCount, f.FailedAt)
	return err
}

func (r *mysqlRepo) DeleteFailure(id string) error {
	_, err := r.db.Exec(`DELETE FROM crawl_failures WHERE id = ?`, id)
	return err
}

// ---------- 爬虫配置 ----------

func (r *mysqlRepo) GetConfig() (domain.CrawlerConfig, error) {
	var c domain.CrawlerConfig
	var enabled int
	err := r.db.QueryRow(`SELECT concurrency, timeout, retry_count, user_agent,
		request_interval, max_pages, official_repos, default_query, enabled
		FROM crawler_config WHERE id = 1`).Scan(
		&c.Concurrency, &c.Timeout, &c.RetryCount, &c.UserAgent,
		&c.RequestInterval, &c.MaxPagesPerRun, &c.OfficialRepos, &c.DefaultQuery, &enabled)
	c.Enabled = enabled == 1
	return c, err
}

func (r *mysqlRepo) SaveConfig(c domain.CrawlerConfig) error {
	enabled := 0
	if c.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(
		`UPDATE crawler_config SET concurrency=?, timeout=?, retry_count=?, user_agent=?,
			request_interval=?, max_pages=?, official_repos=?, default_query=?, enabled=?
		 WHERE id = 1`,
		c.Concurrency, c.Timeout, c.RetryCount, c.UserAgent,
		c.RequestInterval, c.MaxPagesPerRun, c.OfficialRepos, c.DefaultQuery, enabled)
	return err
}

// ---------- 抓取数据（数据审核） ----------

func (r *mysqlRepo) ListData(status string) ([]DataItem, error) {
	query := `SELECT id, name, author, category, github_stars, is_official, github_url, data_status
		FROM skills`
	var args []any
	if status != "" {
		query += ` WHERE data_status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY id DESC LIMIT 200`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]DataItem, 0)
	for rows.Next() {
		var d DataItem
		var official int
		if err := rows.Scan(&d.ID, &d.Name, &d.Author, &d.Category, &d.GithubStars, &official, &d.Source, &d.Status); err != nil {
			return nil, err
		}
		d.IsOfficial = official == 1
		d.FetchedAt = ""
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) UpdateDataStatus(id, status string) error {
	_, err := r.db.Exec(`UPDATE skills SET data_status = ? WHERE id = ?`, status, id)
	return err
}

func (r *mysqlRepo) DeleteData(id string) error {
	_, err := r.db.Exec(`DELETE FROM skills WHERE id = ?`, id)
	return err
}

// ---------- 文章 ----------

func (r *mysqlRepo) ListArticles() ([]domain.Article, error) {
	rows, err := r.db.Query(`SELECT id, title, status, category, author, views, updated_at
		FROM articles ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Article, 0)
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Status, &a.Category, &a.Author, &a.Views, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateArticle(a *domain.Article) error {
	_, err := r.db.Exec(
		`INSERT INTO articles(id, title, status, category, author, views, updated_at)
		 VALUES(?,?,?,?,?,0,?)`,
		a.ID, a.Title, a.Status, a.Category, a.Author, a.UpdatedAt)
	return err
}

func (r *mysqlRepo) DeleteArticle(id string) error {
	_, err := r.db.Exec(`DELETE FROM articles WHERE id = ?`, id)
	return err
}

// ---------- SEO / 站点 ----------

func (r *mysqlRepo) GetSeo() (domain.SeoConfig, error) {
	var s domain.SeoConfig
	err := r.db.QueryRow(`SELECT title, description, keywords, og_image FROM seo_config WHERE id = 1`).
		Scan(&s.Title, &s.Description, &s.Keywords, &s.OgImage)
	return s, err
}

func (r *mysqlRepo) SaveSeo(s domain.SeoConfig) error {
	_, err := r.db.Exec(`UPDATE seo_config SET title=?, description=?, keywords=?, og_image=? WHERE id = 1`,
		s.Title, s.Description, s.Keywords, s.OgImage)
	return err
}

func (r *mysqlRepo) GetSiteConfig() (domain.SiteConfig, error) {
	var s domain.SiteConfig
	err := r.db.QueryRow(`SELECT site_name, slogan, portal_url, icp, contact_email FROM site_config WHERE id = 1`).
		Scan(&s.SiteName, &s.Slogan, &s.PortalUrl, &s.ICP, &s.ContactEmail)
	return s, err
}

func (r *mysqlRepo) SaveSiteConfig(s domain.SiteConfig) error {
	_, err := r.db.Exec(`UPDATE site_config SET site_name=?, slogan=?, portal_url=?, icp=?, contact_email=? WHERE id = 1`,
		s.SiteName, s.Slogan, s.PortalUrl, s.ICP, s.ContactEmail)
	return err
}

// ---------- 认证 ----------

func (r *mysqlRepo) GetAdmin(username string) (AdminRecord, error) {
	var a AdminRecord
	err := r.db.QueryRow(`SELECT username, password_hash, display_name FROM admin_users WHERE username = ?`, username).
		Scan(&a.Username, &a.PasswordHash, &a.DisplayName)
	return a, err
}

func (r *mysqlRepo) UpdateAdminPassword(username, hash string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET password_hash = ? WHERE username = ?`, hash, username)
	return err
}

// ---------- 执行辅助 ----------

func (r *mysqlRepo) SkillExists(id string) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE id = ?`, id).Scan(&n)
	return n > 0, err
}

func (r *mysqlRepo) InsertCrawledSkills(skills []InsertSkill) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, s := range skills {
		tags, _ := json.Marshal(s.Tags)
		content, _ := json.Marshal(s.Content)
		official := 0
		if s.IsOfficial {
			official = 1
		}
		if _, err := tx.Exec(
			`INSERT INTO skills(id, name, author, description, category, download_url,
				is_official, is_featured, install_command, github_url, github_stars, license, tags, content, data_status)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			s.ID, s.Name, s.Author, s.Description, s.Category, s.DownloadURL,
			official, 0, "", s.GithubURL, s.GithubStars, s.License, string(tags), string(content), "pending",
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *mysqlRepo) CountSkills() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM skills`).Scan(&n)
	return n, err
}

func (r *mysqlRepo) CountOfficial() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE is_official = 1`).Scan(&n)
	return n, err
}

func (r *mysqlRepo) CountAuthors() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(DISTINCT author) FROM skills`).Scan(&n)
	return n, err
}

func (r *mysqlRepo) Stats() (domain.AdminStats, error) {
	var st domain.AdminStats
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM crawl_tasks WHERE status = 'running'`).Scan(&st.RunRunning)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM crawl_tasks`).Scan(&st.TodayTasks)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM crawl_failures`).Scan(&st.RunFailed)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE data_status = 'pending'`).Scan(&st.PendingData)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM skills`).Scan(&st.TotalSkills)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE is_official = 1`).Scan(&st.OfficialNums)
	_ = r.db.QueryRow(`SELECT COUNT(DISTINCT author) FROM skills`).Scan(&st.TotalAuthors)
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM categories`).Scan(&st.TotalCats)
	// 今日成功 / 新增数据（简化：执行记录中成功数）
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM crawl_executions WHERE status = 'success'`).Scan(&st.RunSuccess)
	return st, nil
}

// update 通用更新：仅更新 fields 中出现的列。
func (r *mysqlRepo) update(table, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	cols := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))
	for k, v := range fields {
		cols = append(cols, k+" = ?")
		args = append(args, v)
	}
	args = append(args, id)
	_, err := r.db.Exec(fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(cols, ", ")), args...)
	return err
}
