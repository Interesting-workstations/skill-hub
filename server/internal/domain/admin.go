// Package domain 定义后台管理（admin）核心业务模型。
// 覆盖：爬虫任务、执行记录、失败任务、爬虫配置、内容管理、认证。
package domain

// TaskStatus 爬虫任务状态。
type TaskStatus string

const (
	TaskWaiting TaskStatus = "waiting"
	TaskRunning TaskStatus = "running"
	TaskSuccess TaskStatus = "success"
	TaskFailed  TaskStatus = "failed"
	TaskStopped TaskStatus = "stopped"
)

// CrawlTask 爬虫任务。
type CrawlTask struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Query        string     `json:"query"`
	Status       TaskStatus `json:"status"`
	Schedule     string     `json:"schedule"`
	LastRunAt    string     `json:"lastRunAt"`
	LastDuration string     `json:"lastDuration"`
	RunCount     int        `json:"runCount"`
	SuccessCount int        `json:"successCount"`
	FailCount    int        `json:"failCount"`
	CreatedAt    string     `json:"createdAt"`
}

// LogLine 执行日志中的一行。
type LogLine struct {
	Time  string `json:"time"`
	Level string `json:"level"`
	Text  string `json:"text"`
}

// ExecutionStats 执行结果统计。
type ExecutionStats struct {
	Pages     int `json:"pages"`
	Fetched   int `json:"fetched"`
	Failed    int `json:"failed"`
	NewData   int `json:"newData"`
	Updated   int `json:"updated"`
	Duplicate int `json:"duplicate"`
}

// ExecutionRecord 一次任务执行记录。
type ExecutionRecord struct {
	ID        string         `json:"id"`
	TaskID    string         `json:"taskId"`
	TaskName  string         `json:"taskName"`
	Status    TaskStatus     `json:"status"`
	StartTime string         `json:"startTime"`
	EndTime   string         `json:"endTime"`
	Duration  string         `json:"duration"`
	Progress  int            `json:"progress"`
	Stats     ExecutionStats `json:"stats"`
	Logs      []LogLine      `json:"logs"`
}

// FailureRecord 失败任务记录。
type FailureRecord struct {
	ID         string `json:"id"`
	TaskID     string `json:"taskId"`
	TaskName   string `json:"taskName"`
	URL        string `json:"url"`
	Reason     string `json:"reason"`
	Error      string `json:"error"`
	RetryCount int    `json:"retryCount"`
	FailedAt   string `json:"failedAt"`
}

// CrawlerConfig 全局爬虫运行配置。
type CrawlerConfig struct {
	Concurrency     int    `json:"concurrency"`
	Timeout         int    `json:"timeout"`
	RetryCount      int    `json:"retryCount"`
	UserAgent       string `json:"userAgent"`
	RequestInterval int    `json:"requestInterval"`
	MaxPagesPerRun  int    `json:"maxPagesPerRun"`
	OfficialRepos   string `json:"officialRepos"`
	DefaultQuery    string `json:"defaultQuery"`
	Enabled         bool   `json:"enabled"`
}

// Article 官网文章。
type Article struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Category  string `json:"category"`
	Author    string `json:"author"`
	Views     int    `json:"views"`
	UpdatedAt string `json:"updatedAt"`
}

// SeoConfig SEO 配置。
type SeoConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	OgImage     string `json:"ogImage"`
}

// SiteConfig 网站基础配置。
type SiteConfig struct {
	SiteName     string `json:"siteName"`
	Slogan       string `json:"slogan"`
	PortalUrl    string `json:"portalUrl"`
	ICP          string `json:"icp"`
	ContactEmail string `json:"contactEmail"`
}

// AdminUser 管理员账号（单管理员）。
type AdminUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// DataStatus 抓取数据审核状态（官网技能表的 data_status 列）。
type DataStatus string

const (
	DataPending   DataStatus = "pending"   // 待审核
	DataApproved  DataStatus = "approved"  // 已审核
	DataPublished DataStatus = "published" // 已发布（官网可见）
	DataIgnored   DataStatus = "ignored"   // 已忽略
)

// AdminStats 后台工作台聚合指标。
type AdminStats struct {
	TodayTasks   int `json:"todayTasks"`
	RunSuccess   int `json:"runSuccess"`
	RunFailed    int `json:"runFailed"`
	RunRunning   int `json:"runRunning"`
	NewData      int `json:"newData"`
	PendingData  int `json:"pendingData"`
	TotalSkills  int `json:"totalSkills"`
	OfficialNums int `json:"officialSkills"`
	TotalAuthors int `json:"totalAuthors"`
	TotalCats    int `json:"totalCategories"`
}
