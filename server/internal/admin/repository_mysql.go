// Package admin 是管理后台业务模块（Handler → Service → Repository）。
// 数据全部持久化到 MySQL：爬虫任务、执行记录、失败任务、配置、内容、认证。
package admin

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Interesting-workstations/skill-hub/server/internal/domain"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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

// DataFilter 抓取数据列表的筛选条件（审核页）。
// Source：official=官方来源 / community=社区个人；Sort：stars=高星优先 / name / newest。
type DataFilter struct {
	Status   string
	Source   string
	Category string
	Author   string
	Query    string
	Sort     string
}

// AutoAuditResult 机器人自动审核结果。
type AutoAuditResult struct {
	Total    int `json:"total"`
	Approved int `json:"approved"`
	Manual   int `json:"manual"`
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
	DeleteExecution(id string) error
	RecoverStale() error

	ListFailures() ([]domain.FailureRecord, error)
	CreateFailure(f *domain.FailureRecord) error
	DeleteFailure(id string) error

	GetConfig() (domain.CrawlerConfig, error)
	SaveConfig(c domain.CrawlerConfig) error

	// GitHub Token 池（独立表，每行一个 token）
	ListGitHubTokens() ([]domain.GitHubToken, error)
	CreateGitHubToken(t *domain.GitHubToken) error
	UpdateGitHubToken(id string, fields map[string]any) error
	DeleteGitHubToken(id string) error

	ListData(f DataFilter) ([]DataItem, error)
	UpdateDataStatus(id, status string) error
	UpdateDataStatusBatch(ids []string, status string) error
	DeleteData(id string) error
	AutoAuditPending() (AutoAuditResult, error)
	PublishAllApproved() (int64, error)

	// 翻译管理：扫描未汉化技能 + 更新翻译结果
	ListUntranslatedSkills(limit int) ([]domain.TranslationItem, error)
	CountUntranslatedSkills() (int, error)
	UpdateSkillTranslation(id, nameZh, descZh string) error

	// 翻译配置：主翻译通道（'' = 用环境变量默认）
	GetTranslateProvider() (string, error)
	SaveTranslateProvider(provider string) error

	ListArticles() ([]domain.Article, error)
	GetArticle(id string) (domain.Article, error)
	CreateArticle(a *domain.Article) error
	UpdateArticle(id string, a *domain.Article) error
	DeleteArticle(id string) error

	// ListExportSkills 返回全部技能完整数据（供导出 JSON/CSV/Markdown）。
	ListExportSkills() ([]domain.Skill, error)

	// 赞助商
	ListSponsors() ([]domain.Sponsor, error)
	CreateSponsor(s *domain.Sponsor) error
	UpdateSponsor(id string, s *domain.Sponsor) error
	DeleteSponsor(id string) error
	IncrSponsorClicks(id string) error

	GetSeo() (domain.SeoConfig, error)
	SaveSeo(s domain.SeoConfig) error

	GetSiteConfig() (domain.SiteConfig, error)
	SaveSiteConfig(s domain.SiteConfig) error

	GetAdmin(username string) (AdminRecord, error)
	UpdateAdminPassword(username, hash string) error

	// 认证安全扩展
	GetAdminAuth(username string) (AdminRecord, error)
	UpdateAdminLogin(username, ip, ua, at string) error
	IncAdminFail(username string) error
	ResetAdminFail(username string) error
	SetAdminLocked(username, until string) error
	InsertLoginLog(l domain.AdminLoginLog) error
	GetLoginLogs(limit int) ([]domain.AdminLoginLog, error)

	SkillExists(id string) (bool, error)
	InsertCrawledSkills(skills []InsertSkill) error
	// 批量判重：返回已存在的 ID 集合，与已存在的 (name, author, githubUrl) 指纹集合
	ListExistingSkillIDs(ids []string) (map[string]bool, error)
	ListExistingSkillSources(pairs [][3]string) (map[string]bool, error)
	CountSkills() (int, error)
	CountOfficial() (int, error)
	CountAuthors() (int, error)
	Stats() (domain.AdminStats, error)

	// 官方组织（动态管理）
	ListOfficialOrgs() ([]domain.OfficialOrg, error)
	CreateOfficialOrg(o *domain.OfficialOrg) error
	UpdateOfficialOrg(owner string, fields map[string]any) error
	DeleteOfficialOrg(owner string) error
}

// AdminRecord 管理员账号记录（含安全相关字段；密码哈希不可逆存储）。
type AdminRecord struct {
	Username     string
	PasswordHash string
	DisplayName  string
	LastLoginAt  string
	LastLoginIP  string
	LastLoginUA  string
	FailCount    int
	LockedUntil  string
}

// InsertSkill 爬虫结果写入 skills 表所需的字段。
type InsertSkill struct {
	ID            string
	Name          string
	NameZh        string
	Author        string
	Description   string
	DescriptionZh string
	Category      string
	DownloadURL   string
	IsOfficial    bool
	GithubURL     string
	GithubStars   string
	License       string
	SkillPath     string
	Tags          []string
	Content       []domain.ContentSection
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
	if err := repo.seedOfficialOrgs(); err != nil {
		db.Close()
		return nil, err
	}
	if err := repo.migrateOfficialOrgLogos(); err != nil {
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
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			github_tokens TEXT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS articles (
			id VARCHAR(64) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			status VARCHAR(16) NOT NULL DEFAULT 'draft',
			category VARCHAR(64) NOT NULL DEFAULT '教程',
			author VARCHAR(64) NOT NULL DEFAULT 'admin',
			views INT NOT NULL DEFAULT 0,
			updated_at VARCHAR(16) NOT NULL,
			content MEDIUMTEXT NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS sponsors (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(128) NOT NULL,
			logo VARCHAR(255) NOT NULL DEFAULT '',
			description_zh TEXT NOT NULL,
			description_en TEXT NOT NULL,
			url VARCHAR(500) NOT NULL DEFAULT '',
			position VARCHAR(16) NOT NULL DEFAULT 'home',
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			sort_order INT NOT NULL DEFAULT 0,
			clicks INT NOT NULL DEFAULT 0,
			created_at VARCHAR(32) NOT NULL
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
		`CREATE TABLE IF NOT EXISTS admin_login_logs (
			id VARCHAR(64) PRIMARY KEY,
			username VARCHAR(64) NOT NULL DEFAULT '',
			action VARCHAR(16) NOT NULL,
			ip VARCHAR(64) NOT NULL DEFAULT '',
			user_agent VARCHAR(255) NOT NULL DEFAULT '',
			created_at VARCHAR(32) NOT NULL,
			KEY idx_login_username (username),
			KEY idx_login_created (created_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS official_orgs (
			owner VARCHAR(100) PRIMARY KEY,
			display_name VARCHAR(100) NOT NULL DEFAULT '',
			avatar VARCHAR(32) NOT NULL DEFAULT '',
			logo_url VARCHAR(255) NOT NULL DEFAULT '',
			sort_order INT NOT NULL DEFAULT 0,
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			created_at VARCHAR(16) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS translate_config (
			id TINYINT PRIMARY KEY,
			provider VARCHAR(32) NOT NULL DEFAULT '',
			updated_at VARCHAR(32) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
		`CREATE TABLE IF NOT EXISTS github_tokens (
			id VARCHAR(64) PRIMARY KEY,
			token TEXT NOT NULL,
			remark VARCHAR(255) NOT NULL DEFAULT '',
			enabled TINYINT(1) NOT NULL DEFAULT 1,
			created_at VARCHAR(16) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}
	for _, stmt := range stmts {
		if _, err := r.db.Exec(stmt); err != nil {
			return err
		}
	}
	if err := r.ensureAdminColumns(); err != nil {
		return err
	}
	if err := r.ensureDataStatusColumn(); err != nil {
		return err
	}
	if err := r.ensureSkillPathColumn(); err != nil {
		return err
	}
	if err := r.ensureLogoURLColumn(); err != nil {
		return err
	}
	if err := r.ensureGitHubTokensColumn(); err != nil {
		return err
	}
	if err := r.migrateLegacyGitHubTokens(); err != nil {
		return err
	}
	return r.ensureArticleContentColumn()
}

// ensureGitHubTokensColumn 给已存在的 crawler_config 表补充 github_tokens（Token 池）列。
func (r *mysqlRepo) ensureGitHubTokensColumn() error {
	if _, err := r.db.Exec(`ALTER TABLE crawler_config ADD COLUMN github_tokens TEXT NOT NULL`); err != nil {
		if strings.Contains(err.Error(), "Duplicate column") {
			return nil
		}
		return err
	}
	return nil
}

// migrateLegacyGitHubTokens 把旧版 crawler_config.github_tokens（逗号分隔）迁移到独立表 github_tokens。
// 仅当 github_tokens 表为空时执行一次，避免重复导入。
func (r *mysqlRepo) migrateLegacyGitHubTokens() error {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM github_tokens`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil // 已有数据，跳过
	}
	var legacy string
	if err := r.db.QueryRow(`SELECT github_tokens FROM crawler_config WHERE id = 1`).Scan(&legacy); err != nil {
		// 表不存在或列不存在（旧版本）忽略
		if strings.Contains(err.Error(), "no such column") || strings.Contains(err.Error(), "Unknown column") {
			return nil
		}
		return err
	}
	for _, t := range strings.Split(legacy, ",") {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, err := r.db.Exec(
			`INSERT INTO github_tokens(id, token, remark, enabled, created_at) VALUES(?, ?, '', 1, ?)`,
			newID("tok"), t, today(),
		); err != nil {
			return err
		}
	}
	return nil
}

// ensureLogoURLColumn 给已存在的 official_orgs 表补充 logo_url（logo 图片）列。
func (r *mysqlRepo) ensureLogoURLColumn() error {
	if _, err := r.db.Exec(`ALTER TABLE official_orgs ADD COLUMN logo_url VARCHAR(255) NOT NULL DEFAULT ''`); err != nil {
		if strings.Contains(err.Error(), "Duplicate column") {
			return nil
		}
		return err
	}
	return nil
}

// ensureArticleContentColumn 给已存在的 articles 表补充 content（正文）列。
func (r *mysqlRepo) ensureArticleContentColumn() error {
	if _, err := r.db.Exec(`ALTER TABLE articles ADD COLUMN content MEDIUMTEXT NOT NULL`); err != nil {
		// 列已存在或表不存在均忽略
		if strings.Contains(err.Error(), "Duplicate column") || strings.Contains(err.Error(), "check that column/key exists") {
			return nil
		}
		return err
	}
	return nil
}

// adminExtraColumns 需保证存在的 admin_users 安全相关列（已有表增量迁移）。
var adminExtraColumns = []string{
	"last_login_at", "last_login_ip", "last_login_ua", "fail_count", "locked_until",
}

// ensureAdminColumns 给已存在的 admin_users 表补充安全相关列（缺列则 ALTER）。
func (r *mysqlRepo) ensureAdminColumns() error {
	existing := map[string]bool{}
	rows, err := r.db.Query(`SELECT COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_users'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		existing[name] = true
	}
	for _, col := range adminExtraColumns {
		if existing[col] {
			continue
		}
		var def string
		switch col {
		case "last_login_at":
			def = "VARCHAR(32) NOT NULL DEFAULT ''"
		case "last_login_ip":
			def = "VARCHAR(64) NOT NULL DEFAULT ''"
		case "last_login_ua":
			def = "VARCHAR(255) NOT NULL DEFAULT ''"
		case "fail_count":
			def = "INT NOT NULL DEFAULT 0"
		case "locked_until":
			def = "VARCHAR(32) NOT NULL DEFAULT ''"
		}
		if _, err := r.db.Exec(`ALTER TABLE admin_users ADD COLUMN ` + col + ` ` + def); err != nil {
			// 并发/竞态下列可能刚被创建，忽略 Duplicate column
			if strings.Contains(err.Error(), "Duplicate column") {
				continue
			}
			return err
		}
	}
	return nil
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

// ensureSkillPathColumn 给 skills 表补充 skill_path（技能目录）列（已存在则忽略）。
func (r *mysqlRepo) ensureSkillPathColumn() error {
	_, err := r.db.Exec(`ALTER TABLE skills ADD COLUMN skill_path VARCHAR(500) NOT NULL DEFAULT ''`)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate column") || strings.Contains(err.Error(), "check that column/key exists") {
			return nil
		}
		return err
	}
	return nil
}

// randomPassword 生成随机初始密码（首次部署未配置 ADMIN_INIT_PASSWORD 时使用，避免弱密码）。
func randomPassword(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "SkillHubInit@2026"
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// seed 初始化后台种子数据（管理员、爬虫配置、SEO、站点、示例任务）。
func (r *mysqlRepo) seed() error {
	// 管理员
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM admin_users`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		// 初始管理员密码：优先环境变量 ADMIN_INIT_PASSWORD，否则生成随机强密码
		initPwd := os.Getenv("ADMIN_INIT_PASSWORD")
		generated := false
		if initPwd == "" {
			initPwd = randomPassword(16)
			generated = true
		}
		hash, err := HashPassword(initPwd)
		if err != nil {
			return err
		}
		if _, err := r.db.Exec(
			`INSERT INTO admin_users(username, password_hash, display_name) VALUES(?,?,?)`,
			"admin", hash, "管理员",
		); err != nil {
			return err
		}
		if generated {
			log.Printf("⚠️ 已创建初始管理员 admin，初始密码：%s（请登录后立即修改）", initPwd)
		} else {
			log.Printf("⚠️ 已创建初始管理员 admin（密码来自 ADMIN_INIT_PASSWORD 环境变量），请及时修改")
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
				request_interval, max_pages, official_repos, default_query, enabled, github_tokens)
			 VALUES(1, 4, 20, 3, 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) SkillHubBot/1.0',
				500, 500, 'anthropics/skills,openai/codex,vercel/ai', 'claude skills', 1, '')`,
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
			{ID: "task-5", Name: "MCP 官方服务器", Type: "skill", Query: "modelcontextprotocol/servers", Status: domain.TaskWaiting, Schedule: "每天 04:00", RunCount: 0, CreatedAt: "2026-08-15 14:30"},
			{ID: "task-6", Name: "头部公司技能库", Type: "skill", Query: "replicate/skills, SAP/ai-skills-library, databricks-agent-skills, elevenlabs/skills", Status: domain.TaskWaiting, Schedule: "每天 06:00", RunCount: 0, CreatedAt: "2026-08-15 14:30"},
			{ID: "task-7", Name: "高星技能挖掘", Type: "skill", Query: "claude skills stars:>300", Status: domain.TaskWaiting, Schedule: "每 8 小时", RunCount: 0, CreatedAt: "2026-08-15 14:30"},
			{ID: "task-8", Name: "MCP 生态搜索", Type: "skill", Query: "mcp server", Status: domain.TaskWaiting, Schedule: "每 12 小时", RunCount: 0, CreatedAt: "2026-08-15 14:30"},
			{ID: "task-9", Name: "官方 Agent 工具", Type: "skill", Query: "aws/agent-toolkit-for-aws, cloudflare/security-audit-skill, intel/skills", Status: domain.TaskWaiting, Schedule: "每天 08:00", RunCount: 0, CreatedAt: "2026-08-15 14:30"},
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

// HashPassword 生成密码哈希（bcrypt，不可逆 + 自带安全比较，符合密码安全要求）。
func HashPassword(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword 安全校验密码（恒定时间比较，避免时序攻击）。
// 返回 (是否匹配, 是否旧格式需迁移重哈希)。
// 旧格式为历史遗留的 SHA-256 + 固定盐，检测到且校验通过时由调用方重哈希为 bcrypt。
func VerifyPassword(hash, pwd string) (ok, legacy bool) {
	if strings.HasPrefix(hash, "$2") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
		return err == nil, false
	}
	sum := sha256.Sum256([]byte("skillhub-admin-salt:" + pwd))
	expected := hex.EncodeToString(sum[:])
	return hmac.Equal([]byte(expected), []byte(hash)), true
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

func (r *mysqlRepo) DeleteExecution(id string) error {
	_, err := r.db.Exec(`DELETE FROM crawl_executions WHERE id = ?`, id)
	return err
}

// RecoverStale 服务启动时清理残留状态：
// 上次进程退出/崩溃留下的 running 执行记录标记为“中断”，对应 running 任务恢复为待运行。
func (r *mysqlRepo) RecoverStale() error {
	_, err := r.db.Exec(`UPDATE crawl_executions SET status='failed', progress=100,
		end_time=?, duration='服务重启中断' WHERE status='running'`, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`UPDATE crawl_tasks SET status='waiting' WHERE status='running'`)
	return err
}

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
		request_interval, max_pages, official_repos, default_query, enabled, github_tokens
		FROM crawler_config WHERE id = 1`).Scan(
		&c.Concurrency, &c.Timeout, &c.RetryCount, &c.UserAgent,
		&c.RequestInterval, &c.MaxPagesPerRun, &c.OfficialRepos, &c.DefaultQuery, &enabled, &c.GitHubTokens)
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
			request_interval=?, max_pages=?, official_repos=?, default_query=?, enabled=?, github_tokens=?
		 WHERE id = 1`,
		c.Concurrency, c.Timeout, c.RetryCount, c.UserAgent,
		c.RequestInterval, c.MaxPagesPerRun, c.OfficialRepos, c.DefaultQuery, enabled, c.GitHubTokens)
	return err
}

// ---------- GitHub Token 池 ----------

func (r *mysqlRepo) ListGitHubTokens() ([]domain.GitHubToken, error) {
	rows, err := r.db.Query(`SELECT id, token, remark, enabled, created_at FROM github_tokens ORDER BY enabled DESC, created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.GitHubToken
	for rows.Next() {
		var t domain.GitHubToken
		var enabled int
		if err := rows.Scan(&t.ID, &t.Token, &t.Remark, &enabled, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Enabled = enabled == 1
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateGitHubToken(t *domain.GitHubToken) error {
	enabled := 0
	if t.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(
		`INSERT INTO github_tokens(id, token, remark, enabled, created_at) VALUES(?, ?, ?, ?, ?)`,
		t.ID, t.Token, t.Remark, enabled, t.CreatedAt)
	return err
}

func (r *mysqlRepo) UpdateGitHubToken(id string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	var sets []string
	var args []any
	if v, ok := fields["token"]; ok {
		sets = append(sets, "token = ?")
		args = append(args, v)
	}
	if v, ok := fields["remark"]; ok {
		sets = append(sets, "remark = ?")
		args = append(args, v)
	}
	if v, ok := fields["enabled"]; ok {
		sets = append(sets, "enabled = ?")
		args = append(args, v)
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	_, err := r.db.Exec(`UPDATE github_tokens SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...)
	return err
}

func (r *mysqlRepo) DeleteGitHubToken(id string) error {
	_, err := r.db.Exec(`DELETE FROM github_tokens WHERE id = ?`, id)
	return err
}

// ---------- 抓取数据（数据审核） ----------

func (r *mysqlRepo) ListData(f DataFilter) ([]DataItem, error) {
	query := `SELECT id, name, author, category, github_stars, is_official, github_url, data_status
		FROM skills`
	var conds []string
	var args []any
	if f.Status != "" {
		conds = append(conds, "data_status = ?")
		args = append(args, f.Status)
	}
	if f.Source == "official" {
		conds = append(conds, "is_official = 1")
	} else if f.Source == "community" {
		conds = append(conds, "is_official = 0")
	}
	if f.Category != "" {
		conds = append(conds, "category = ?")
		args = append(args, f.Category)
	}
	if f.Author != "" {
		conds = append(conds, "author = ?")
		args = append(args, f.Author)
	}
	if f.Query != "" {
		conds = append(conds, "(name LIKE ? OR description LIKE ? OR author LIKE ?)")
		kw := "%" + f.Query + "%"
		args = append(args, kw, kw, kw)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	// 高星优先：github_stars 为文本列（可能带 k/m 后缀，如 271.9k），解析为数值排序
	switch f.Sort {
	case "stars":
		query += ` ORDER BY CASE
			WHEN github_stars LIKE '%k' OR github_stars LIKE '%K' THEN CAST(REPLACE(LOWER(github_stars), 'k', '') AS DECIMAL(10,2)) * 1000
			WHEN github_stars LIKE '%m' OR github_stars LIKE '%M' THEN CAST(REPLACE(LOWER(github_stars), 'm', '') AS DECIMAL(10,2)) * 1000000
			ELSE CAST(NULLIF(github_stars, '') AS DECIMAL(10,2))
		END DESC, id DESC`
	case "name":
		query += ` ORDER BY name ASC, id DESC`
	default:
		query += ` ORDER BY id DESC`
	}
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

func (r *mysqlRepo) UpdateDataStatusBatch(ids []string, status string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, 0, len(ids)+1)
	args = append(args, status)
	for _, id := range ids {
		args = append(args, id)
	}
	_, err := r.db.Exec(`UPDATE skills SET data_status = ? WHERE id IN (`+placeholders+`)`, args...)
	return err
}

// PublishAllApproved 一键发布全部已审核（approved）的数据到官网（published）。
// 用于清理历史积压：机器人审核旧逻辑产生的 approved 数据一键上线。
func (r *mysqlRepo) PublishAllApproved() (int64, error) {
	res, err := r.db.Exec(`UPDATE skills SET data_status = 'published' WHERE data_status = 'approved'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *mysqlRepo) DeleteData(id string) error {
	_, err := r.db.Exec(`DELETE FROM skills WHERE id = ?`, id)
	return err
}

// ---------- 翻译管理 ----------

// ListUntranslatedSkills 扫描未汉化技能：标题（name_zh）或描述（description_zh）
// 任一不含中文（CJK 字符）即视为未汉化。limit<=0 表示不限制。
// 注意：旧版假翻译可能把英文原文写进 name_zh（如 Homelab Pihole DNS），
// 这类仍需翻译，因此不做 name_zh==name 排除；纯品牌名（DOCX/PDF 等）
// 在翻译时若结果无变化由调用方标记跳过。
// 返回字段含中英文原文与翻译状态，供翻译页面展示与批量翻译。
func (r *mysqlRepo) ListUntranslatedSkills(limit int) ([]domain.TranslationItem, error) {
	query := `SELECT id, name, name_zh, author, description, description_zh, category
		FROM skills
		WHERE (name_zh = '' OR name_zh NOT REGEXP '[一-龥]')
		   OR (description_zh IS NULL OR description_zh = '' OR description_zh NOT REGEXP '[一-龥]')
		ORDER BY is_official DESC, github_stars DESC, id DESC`
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 用 make 初始化保证无数据时返回空数组 [] 而非 null（前端 map 不报错）
	out := make([]domain.TranslationItem, 0)
	for rows.Next() {
		var it domain.TranslationItem
		var descZh sql.NullString
		if err := rows.Scan(&it.ID, &it.Name, &it.NameZh, &it.Author, &it.Description, &descZh, &it.Category); err != nil {
			return nil, err
		}
		it.DescriptionZh = descZh.String
		it.TitleTranslated = containsCJK(it.NameZh)
		it.DescTranslated = containsCJK(it.DescriptionZh)
		out = append(out, it)
	}
	return out, rows.Err()
}

// CountUntranslatedSkills 统计未汉化技能数量。
func (r *mysqlRepo) CountUntranslatedSkills() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM skills
		WHERE (name_zh = '' OR name_zh NOT REGEXP '[一-龥]')
		   OR (description_zh IS NULL OR description_zh = '' OR description_zh NOT REGEXP '[一-龥]')`).Scan(&n)
	return n, err
}

// UpdateSkillTranslation 保存技能的中文标题与描述翻译结果。
func (r *mysqlRepo) UpdateSkillTranslation(id, nameZh, descZh string) error {
	_, err := r.db.Exec(`UPDATE skills SET name_zh = ?, description_zh = ? WHERE id = ?`, nameZh, descZh, id)
	return err
}

// GetTranslateProvider 读取后台配置的主翻译通道（” 表示用环境变量默认）。
func (r *mysqlRepo) GetTranslateProvider() (string, error) {
	var p string
	err := r.db.QueryRow(`SELECT provider FROM translate_config WHERE id = 1`).Scan(&p)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return p, err
}

// SaveTranslateProvider 保存主翻译通道（upsert 单行）。
func (r *mysqlRepo) SaveTranslateProvider(provider string) error {
	_, err := r.db.Exec(
		`INSERT INTO translate_config (id, provider, updated_at) VALUES (1, ?, ?)
		 ON DUPLICATE KEY UPDATE provider = VALUES(provider), updated_at = VALUES(updated_at)`,
		provider, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

// containsCJK 判断字符串是否包含中文字符（CJK 统一表意文字）。
func containsCJK(s string) bool {
	for _, r := range s {
		if r >= 0x4e00 && r <= 0x9fff {
			return true
		}
	}
	return false
}

// AutoAuditPending 机器人自动审核：对待审核技能按规则判定，
// 内容完整规范且无重复的直接通过并发布（published），有问题的（内容不足/描述缺失/重复/名称异常）留给人工。
func (r *mysqlRepo) AutoAuditPending() (AutoAuditResult, error) {
	// 已通过/已发布集合（重复检测：同 author+name 已存在 → 转人工）
	existing := map[string]bool{}
	erows, err := r.db.Query(`SELECT author, name FROM skills WHERE data_status IN ('published','approved')`)
	if err != nil {
		return AutoAuditResult{}, err
	}
	for erows.Next() {
		var a, n string
		if erows.Scan(&a, &n) == nil {
			existing[strings.ToLower(a+"\x00"+n)] = true
		}
	}
	erows.Close()

	rows, err := r.db.Query(`SELECT id, name, author, description, content FROM skills WHERE data_status = 'pending'`)
	if err != nil {
		return AutoAuditResult{}, err
	}
	defer rows.Close()

	var res AutoAuditResult
	var publishIDs []string
	for rows.Next() {
		var id, name, author, desc, content string
		if err := rows.Scan(&id, &name, &author, &desc, &content); err != nil {
			continue
		}
		res.Total++
		if skillAutoPass(name, desc, content) && !existing[strings.ToLower(author+"\x00"+name)] {
			publishIDs = append(publishIDs, id)
		}
	}
	if err := rows.Err(); err != nil {
		return res, err
	}
	// 通过即发布（published），直接上线官网
	if len(publishIDs) > 0 {
		if err := r.UpdateDataStatusBatch(publishIDs, "published"); err != nil {
			return res, err
		}
	}
	res.Approved = len(publishIDs)
	res.Manual = res.Total - res.Approved
	return res, nil
}

// skillAutoPass 机器人审核通过规则（全部满足才自动通过）：
// ①名称长度 2-100 ②描述至少 20 字符 ③正文（content JSON 文本）至少 800 字符。
// 内容不足/描述缺失/名称异常 → 返回 false，转人工。
func skillAutoPass(name, desc, content string) bool {
	name = strings.TrimSpace(name)
	desc = strings.TrimSpace(desc)
	content = strings.TrimSpace(content)
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	if len(desc) < 20 {
		return false
	}
	if len(content) < 800 {
		return false
	}
	return true
}

// ---------- 文章 ----------

func (r *mysqlRepo) ListArticles() ([]domain.Article, error) {
	rows, err := r.db.Query(`SELECT id, title, status, category, author, views, updated_at, content
		FROM articles ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Article, 0)
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Status, &a.Category, &a.Author, &a.Views, &a.UpdatedAt, &a.Content); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) GetArticle(id string) (domain.Article, error) {
	var a domain.Article
	err := r.db.QueryRow(`SELECT id, title, status, category, author, views, updated_at, content
		FROM articles WHERE id = ?`, id).
		Scan(&a.ID, &a.Title, &a.Status, &a.Category, &a.Author, &a.Views, &a.UpdatedAt, &a.Content)
	return a, err
}

func (r *mysqlRepo) CreateArticle(a *domain.Article) error {
	_, err := r.db.Exec(
		`INSERT INTO articles(id, title, status, category, author, views, updated_at, content)
		 VALUES(?,?,?,?,?,0,?,?)`,
		a.ID, a.Title, a.Status, a.Category, a.Author, a.UpdatedAt, a.Content)
	return err
}

func (r *mysqlRepo) UpdateArticle(id string, a *domain.Article) error {
	_, err := r.db.Exec(
		`UPDATE articles SET title=?, status=?, category=?, author=?, updated_at=?, content=?
		 WHERE id = ?`,
		a.Title, a.Status, a.Category, a.Author, a.UpdatedAt, a.Content, id)
	return err
}

func (r *mysqlRepo) DeleteArticle(id string) error {
	_, err := r.db.Exec(`DELETE FROM articles WHERE id = ?`, id)
	return err
}

// ---------- 赞助商 ----------

func (r *mysqlRepo) ListSponsors() ([]domain.Sponsor, error) {
	rows, err := r.db.Query(`SELECT id, name, logo, description_zh, description_en, url,
		position, enabled, sort_order, clicks, created_at
		FROM sponsors ORDER BY sort_order ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Sponsor, 0)
	for rows.Next() {
		var s domain.Sponsor
		var enabled int
		if err := rows.Scan(&s.ID, &s.Name, &s.Logo, &s.DescriptionZh, &s.DescriptionEn,
			&s.URL, &s.Position, &enabled, &s.SortOrder, &s.Clicks, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Enabled = enabled == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateSponsor(s *domain.Sponsor) error {
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	if s.CreatedAt == "" {
		s.CreatedAt = today()
	}
	_, err := r.db.Exec(
		`INSERT INTO sponsors(id, name, logo, description_zh, description_en, url,
			position, enabled, sort_order, clicks, created_at)
		 VALUES(?,?,?,?,?,?,?,?,?,0,?)`,
		s.ID, s.Name, s.Logo, s.DescriptionZh, s.DescriptionEn, s.URL,
		s.Position, enabled, s.SortOrder, s.CreatedAt)
	return err
}

func (r *mysqlRepo) UpdateSponsor(id string, s *domain.Sponsor) error {
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(
		`UPDATE sponsors SET name=?, logo=?, description_zh=?, description_en=?, url=?,
			position=?, enabled=?, sort_order=? WHERE id = ?`,
		s.Name, s.Logo, s.DescriptionZh, s.DescriptionEn, s.URL,
		s.Position, enabled, s.SortOrder, id)
	return err
}

func (r *mysqlRepo) DeleteSponsor(id string) error {
	_, err := r.db.Exec(`DELETE FROM sponsors WHERE id = ?`, id)
	return err
}

func (r *mysqlRepo) IncrSponsorClicks(id string) error {
	_, err := r.db.Exec(`UPDATE sponsors SET clicks = clicks + 1 WHERE id = ?`, id)
	return err
}

// ListExportSkills 返回全部技能完整字段（供 JSON/CSV/Markdown 导出）。
func (r *mysqlRepo) ListExportSkills() ([]domain.Skill, error) {
	rows, err := r.db.Query(`SELECT id, name, author, description, category,
		download_url, is_official, is_featured, install_command, github_url,
		github_stars, license, skill_path, tags, content FROM skills ORDER BY author, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Skill, 0, 64)
	for rows.Next() {
		var s domain.Skill
		var official, featured int
		var tagsRaw, contentRaw string
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Author, &s.Description, &s.Category,
			&s.DownloadURL, &official, &featured, &s.InstallCommand, &s.GithubURL,
			&s.GithubStars, &s.License, &s.SkillPath, &tagsRaw, &contentRaw,
		); err != nil {
			return nil, err
		}
		s.IsOfficial = official == 1
		s.IsFeatured = featured == 1
		_ = json.Unmarshal([]byte(tagsRaw), &s.Tags)
		_ = json.Unmarshal([]byte(contentRaw), &s.Content)
		out = append(out, s)
	}
	return out, rows.Err()
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

func (r *mysqlRepo) GetAdminAuth(username string) (AdminRecord, error) {
	var a AdminRecord
	err := r.db.QueryRow(`SELECT username, password_hash, display_name,
		last_login_at, last_login_ip, last_login_ua, fail_count, locked_until
		FROM admin_users WHERE username = ?`, username).
		Scan(&a.Username, &a.PasswordHash, &a.DisplayName,
			&a.LastLoginAt, &a.LastLoginIP, &a.LastLoginUA, &a.FailCount, &a.LockedUntil)
	return a, err
}

func (r *mysqlRepo) UpdateAdminPassword(username, hash string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET password_hash = ? WHERE username = ?`, hash, username)
	return err
}

// UpdateAdminLogin 更新最后登录信息（时间 / IP / UA）。
func (r *mysqlRepo) UpdateAdminLogin(username, ip, ua, at string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET last_login_at = ?, last_login_ip = ?, last_login_ua = ? WHERE username = ?`,
		at, ip, ua, username)
	return err
}

func (r *mysqlRepo) IncAdminFail(username string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET fail_count = fail_count + 1 WHERE username = ?`, username)
	return err
}

func (r *mysqlRepo) ResetAdminFail(username string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET fail_count = 0 WHERE username = ?`, username)
	return err
}

func (r *mysqlRepo) SetAdminLocked(username, until string) error {
	_, err := r.db.Exec(`UPDATE admin_users SET locked_until = ? WHERE username = ?`, until, username)
	return err
}

// InsertLoginLog 写入一条登录审计日志（不含任何敏感信息）。
func (r *mysqlRepo) InsertLoginLog(l domain.AdminLoginLog) error {
	_, err := r.db.Exec(`INSERT INTO admin_login_logs(id, username, action, ip, user_agent, created_at)
		VALUES(?,?,?,?,?,?)`,
		l.ID, l.Username, l.Action, l.IP, l.UserAgent, l.CreatedAt)
	return err
}

// GetLoginLogs 查询最近的登录审计日志。
func (r *mysqlRepo) GetLoginLogs(limit int) ([]domain.AdminLoginLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Query(`SELECT id, username, action, ip, user_agent, created_at
		FROM admin_login_logs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.AdminLoginLog, 0)
	for rows.Next() {
		var l domain.AdminLoginLog
		if err := rows.Scan(&l.ID, &l.Username, &l.Action, &l.IP, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// ---------- 执行辅助 ----------

func (r *mysqlRepo) SkillExists(id string) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE id = ?`, id).Scan(&n)
	return n > 0, err
}

// ListExistingSkillIDs 批量查询已存在的技能 ID（返回存在的 ID 集合）。
func (r *mysqlRepo) ListExistingSkillIDs(ids []string) (map[string]bool, error) {
	out := make(map[string]bool)
	if len(ids) == 0 {
		return out, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := r.db.Query(`SELECT id FROM skills WHERE id IN (`+placeholders+`)`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return out, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// ListExistingSkillSources 批量查询已存在的 (name, author, githubUrl) 指纹。
// 返回 "name|author|githubUrl" 集合；同源同名视为同一技能，用于防重复入库。
func (r *mysqlRepo) ListExistingSkillSources(pairs [][3]string) (map[string]bool, error) {
	out := make(map[string]bool)
	if len(pairs) == 0 {
		return out, nil
	}
	conds := make([]string, 0, len(pairs))
	args := make([]any, 0, len(pairs)*3)
	for _, p := range pairs {
		conds = append(conds, "(name = ? AND author = ? AND github_url = ?)")
		args = append(args, p[0], p[1], p[2])
	}
	rows, err := r.db.Query(`SELECT name, author, github_url FROM skills WHERE `+strings.Join(conds, " OR "), args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var name, author, url string
		if err := rows.Scan(&name, &author, &url); err != nil {
			return out, err
		}
		out[name+"|"+author+"|"+url] = true
	}
	return out, rows.Err()
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
		// 官方来源自动通过（直接发布），社区来源进入待审核（人工审核）
		dataStatus := "pending"
		if s.IsOfficial {
			dataStatus = "published"
		}
		if _, err := tx.Exec(
			`INSERT INTO skills(id, name, name_zh, author, description, description_zh, category, download_url,
				is_official, is_featured, install_command, github_url, github_stars, license, skill_path, tags, content, data_status)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			s.ID, s.Name, s.NameZh, s.Author, s.Description, s.DescriptionZh, s.Category, s.DownloadURL,
			official, 0, "", s.GithubURL, s.GithubStars, s.License, s.SkillPath, string(tags), string(content), dataStatus,
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
	// 数据状态分布（审核看板）
	st.StatusDist = map[string]int{}
	if rows, err := r.db.Query(`SELECT data_status, COUNT(*) FROM skills GROUP BY data_status`); err == nil {
		for rows.Next() {
			var k string
			var n int
			if rows.Scan(&k, &n) == nil {
				st.StatusDist[k] = n
			}
		}
		rows.Close()
	}
	// 来源分布（官方 / 社区）
	st.TypeDist = map[string]int{"official": 0, "community": 0}
	var offN int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM skills WHERE is_official = 1`).Scan(&offN)
	st.TypeDist["official"] = offN
	st.TypeDist["community"] = st.TotalSkills - offN
	st.Trend = r.execTrend(7)
	return st, nil
}

// execTrend 统计近 n 天（含今天）每天的执行次数与成功次数。
// 按 crawl_executions.start_time 的前 10 位（YYYY-MM-DD）聚合，无记录的天补 0。
func (r *mysqlRepo) execTrend(n int) []domain.TrendPoint {
	if n <= 0 {
		n = 7
	}
	days := make([]string, 0, n)
	now := time.Now()
	for i := n - 1; i >= 0; i-- {
		days = append(days, now.AddDate(0, 0, -i).Format("2006-01-02"))
	}
	counts := map[string]int{}
	successes := map[string]int{}
	rows, err := r.db.Query(`SELECT LEFT(start_time, 10) AS day, COUNT(*),
		COALESCE(SUM(status = 'success'), 0)
		FROM crawl_executions
		WHERE start_time <> '' AND start_time >= ?
		GROUP BY day`, days[0])
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day string
			var count, success int
			if err := rows.Scan(&day, &count, &success); err == nil {
				counts[day] = count
				successes[day] = success
			}
		}
	}
	out := make([]domain.TrendPoint, 0, n)
	for _, day := range days {
		out = append(out, domain.TrendPoint{
			Day:     day[5:], // MM-DD
			Count:   counts[day],
			Success: successes[day],
		})
	}
	return out
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

// ---------- 官方组织（动态管理） ----------

// seedOfficialOrgs 表为空时预置默认官方组织列表（含大量知名组织）。
// officialLogoOwner 官方组织 owner → 正确的 GitHub 组织名（用于 logo 头像 URL）。
// 部分 owner 是别名/占位，GitHub 上对应的是错误账号（如 anthropic 是个人用户），需修正。
var officialLogoOwner = map[string]string{
	"anthropic":  "anthropics",
	"cohere":     "cohere-ai",
	"pinecone":   "pinecone-io",
	"snowflake":  "snowflakedb",
	"scaleai":    "scaleapi",
	"amazon":     "aws",
	"confluent":  "confluentinc",
	"redhat":     "redhat-developer",
	"samba-nova": "sambanova",
	"slack":      "slackhq",
	"xai":        "xai-org",
	"meta":       "facebook",
}

// officialLogoURL 返回组织 logo 的 GitHub 头像地址（使用修正后的正确组织名）。
func officialLogoURL(owner string) string {
	gh := officialLogoOwner[owner]
	if gh == "" {
		gh = owner
	}
	return fmt.Sprintf("https://github.com/%s.png", gh)
}

// officialLogoOverride 无有效 GitHub 头像的组织 → 使用官网品牌 logo（经 /img-proxy 白名单代理）。
// 这些组织的 GitHub 组织未设置头像（默认 identicon / 纯色块），必须用官网 logo 才能显示正确。
var officialLogoOverride = map[string]string{
	"sst":           "https://sst.dev/favicon.svg",
	"perplexity-ai": "https://www.perplexity.ai/favicon.svg",
	"zhipuai":       "https://www.zhipuai.cn/favicon.png",
	"nomic-ai":      "https://www.nomic.ai/favicon-512x512.png",
}

// officialLogo 计算组织 logo 地址：官网覆盖优先，否则 GitHub 头像。
func officialLogo(owner string) string {
	if ov, ok := officialLogoOverride[owner]; ok {
		return ov
	}
	return officialLogoURL(owner)
}

func (r *mysqlRepo) seedOfficialOrgs() error {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM official_orgs`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	// owner, displayName, avatar, sortOrder
	orgs := [][4]any{
		{"anthropics", "Anthropic", "🅰️", 1},
		{"openai", "OpenAI", "🤖", 2},
		{"microsoft", "Microsoft", "🪟", 3},
		{"google", "Google", "🇬", 4},
		{"googlecloudplatform", "Google", "🇬", 4},
		{"deepmind", "DeepMind", "🇬", 4},
		{"vercel", "Vercel", "▲", 5},
		{"github", "GitHub", "🐙", 6},
		{"cloudflare", "Cloudflare", "☁️", 7},
		{"figma", "Figma", "🎨", 8},
		{"notion", "Notion", "📝", 9},
		{"stripe", "Stripe", "💳", 10},
		{"aws", "Amazon", "☁️", 11},
		{"aws-samples", "AWS", "☁️", 11},
		{"sst", "SST", "▲", 12},
		{"facebook", "Meta", "🔵", 13},
		{"huggingface", "Hugging Face", "🤗", 14},
		{"ibm", "IBM", "🔷", 15},
		{"oracle", "Oracle", "🟠", 16},
		{"apple", "Apple", "🍎", 17},
		{"netflix", "Netflix", "🎬", 18},
		{"linkedin", "LinkedIn", "💼", 19},
		{"alibaba", "Alibaba", "🅰️", 21},
		{"tencent", "Tencent", "🐧", 22},
		{"baidu", "Baidu", "🐻", 23},
		{"xai-org", "xAI", "🐦", 24},
		{"mistralai", "Mistral AI", "🌀", 25},
		{"cohere-ai", "Cohere", "🌊", 26},
		{"databricks", "Databricks", "🔥", 27},
		{"snowflakedb", "Snowflake", "❄️", 28},
		{"nvidia", "NVIDIA", "🟢", 29},
		{"intel", "Intel", "🔵", 30},
		{"amd", "AMD", "🔴", 31},
		{"samsung", "Samsung", "🔷", 32},
		{"bytedance", "ByteDance", "🎵", 33},
		{"tiktok", "TikTok", "🎵", 33},
		{"redhat-developer", "Red Hat", "🧢", 34},
		{"docker", "Docker", "🐳", 35},
		{"vmware", "VMware", "🟠", 36},
		{"salesforce", "Salesforce", "☁️", 37},
		{"sap", "SAP", "🟡", 38},
		{"uber", "Uber", "🚗", 39},
		{"shopify", "Shopify", "🛍️", 40},
		{"spotify", "Spotify", "🎧", 41},
		{"airbnb", "Airbnb", "🏠", 42},
		{"dropbox", "Dropbox", "📦", 43},
		{"slackhq", "Slack", "💬", 44},
		{"zoom", "Zoom", "📹", 45},
		{"atlassian", "Atlassian", "🧩", 46},
		{"gitlab", "GitLab", "🦊", 47},
		{"hashicorp", "HashiCorp", "🏗️", 48},
		{"mongodb", "MongoDB", "🍃", 49},
		{"redis", "Redis", "🔴", 50},
		{"neo4j", "Neo4j", "🟢", 51},
		{"elastic", "Elastic", "🟡", 52},
		{"datadog", "Datadog", "🐕", 53},
		{"confluentinc", "Confluent", "🌊", 54},
		{"pinecone-io", "Pinecone", "🌲", 55},
		{"replicate", "Replicate", "🌀", 56},
		{"langchain-ai", "LangChain", "🦜", 57},
		{"wandb", "Weights & Biases", "📈", 58},
		{"perplexity-ai", "Perplexity", "🔮", 59},
		{"elevenlabs", "ElevenLabs", "🎙️", 60},
		{"deepseek-ai", "DeepSeek", "🐋", 61},
		{"qwenlm", "Qwen", "🐉", 62},
		{"zhipuai", "Zhipu AI", "🧠", 63},
		{"moonshotai", "Moonshot AI", "🌙", 64},
		{"ai21labs", "AI21", "🌈", 65},
		{"tiiuae", "TII", "🕌", 66},
		{"scaleapi", "Scale AI", "📐", 67},
		{"run-ai", "Run.ai", "🏃", 68},
		{"sambanova", "SambaNova", "🦙", 69},
		{"cerebras", "Cerebras", "🧠", 70},
		{"groq", "Groq", "⚡", 71},
		{"nomic-ai", "Nomic", "🧭", 72},
	}
	for _, o := range orgs {
		owner := o[0].(string)
		if _, err := r.db.Exec(
			`INSERT INTO official_orgs(owner, display_name, avatar, sort_order, enabled, logo_url, created_at) VALUES(?,?,?,?,1,?,?)`,
			o[0], o[1], o[2], o[3], officialLogo(owner), today(),
		); err != nil {
			return err
		}
	}
	return nil
}

// migrateOfficialOrgLogos 为已有官方组织补充 logo_url（用修正后的正确 GitHub 组织名）。
func (r *mysqlRepo) migrateOfficialOrgLogos() error {
	rows, err := r.db.Query(`SELECT owner, logo_url FROM official_orgs`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var pending []string
	for rows.Next() {
		var owner, logo string
		if err := rows.Scan(&owner, &logo); err != nil {
			continue
		}
		if logo == "" {
			pending = append(pending, owner)
		}
	}
	for _, owner := range pending {
		if _, err := r.db.Exec(`UPDATE official_orgs SET logo_url = ? WHERE owner = ?`, officialLogo(owner), owner); err != nil {
			return err
		}
	}
	return nil
}

func (r *mysqlRepo) ListOfficialOrgs() ([]domain.OfficialOrg, error) {
	rows, err := r.db.Query(`SELECT owner, display_name, avatar, logo_url, sort_order, enabled, created_at
		FROM official_orgs ORDER BY sort_order, owner`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.OfficialOrg, 0)
	for rows.Next() {
		var o domain.OfficialOrg
		var enabled int
		if err := rows.Scan(&o.Owner, &o.DisplayName, &o.Avatar, &o.LogoURL, &o.SortOrder, &enabled, &o.CreatedAt); err != nil {
			return nil, err
		}
		o.Enabled = enabled == 1
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *mysqlRepo) CreateOfficialOrg(o *domain.OfficialOrg) error {
	enabled := 0
	if o.Enabled {
		enabled = 1
	}
	if o.CreatedAt == "" {
		o.CreatedAt = today()
	}
	if o.LogoURL == "" {
		o.LogoURL = officialLogoURL(o.Owner)
	}
	_, err := r.db.Exec(
		`INSERT INTO official_orgs(owner, display_name, avatar, logo_url, sort_order, enabled, created_at) VALUES(?,?,?,?,?,?,?)`,
		o.Owner, o.DisplayName, o.Avatar, o.LogoURL, o.SortOrder, enabled, o.CreatedAt)
	return err
}

func (r *mysqlRepo) UpdateOfficialOrg(owner string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	cols := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields))
	for k, v := range fields {
		cols = append(cols, k+" = ?")
		args = append(args, v)
	}
	args = append(args, owner)
	_, err := r.db.Exec(fmt.Sprintf("UPDATE official_orgs SET %s WHERE owner = ?", strings.Join(cols, ", ")), args...)
	return err
}

func (r *mysqlRepo) DeleteOfficialOrg(owner string) error {
	_, err := r.db.Exec(`DELETE FROM official_orgs WHERE owner = ?`, owner)
	return err
}
