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
	// GitHubTokens GitHub Token 池（逗号分隔多个，自动故障切换）。
	GitHubTokens string `json:"githubTokens"`
}

// GitHubToken 后台 GitHub Token 池条目（独立表存储，每行一个 token）。
type GitHubToken struct {
	ID        string `json:"id"`
	Token     string `json:"token"` // 完整 token（写入时）；读取时可脱敏
	Remark    string `json:"remark"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"createdAt"`
	// Broken 运行时是否被熔断（限流/失效冷却中）；CooldownAt 预计恢复时间。
	Broken     bool   `json:"broken"`
	CooldownAt string `json:"cooldownAt"`
}

// TranslationItem 翻译管理：待翻译/已翻译的技能条目（标题+描述双字段）。
type TranslationItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	NameZh        string `json:"nameZh"`
	Author        string `json:"author"`
	Description   string `json:"description"`
	DescriptionZh string `json:"descriptionZh"`
	Category      string `json:"category"`
	// TitleTranslated 标题是否已汉化（含中文）。
	TitleTranslated bool `json:"titleTranslated"`
	// DescTranslated 描述是否已汉化（含中文）。
	DescTranslated bool `json:"descTranslated"`
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
	// Content 文章正文（Markdown）。
	Content string `json:"content,omitempty"`
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

// Sponsor 官网赞助商 / 广告位内容。
type Sponsor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Logo 赞助商标识（emoji 或图片 URL）。
	Logo string `json:"logo"`
	// DescriptionZh / DescriptionEn 中英文描述（官网按当前语言展示对应文案）。
	DescriptionZh string `json:"descriptionZh"`
	DescriptionEn string `json:"descriptionEn"`
	// URL 点击跳转链接。
	URL string `json:"url"`
	// Position 展示位置：home（首页横幅）/ sidebar（详情页侧边栏）/ both。
	Position string `json:"position"`
	Enabled  bool   `json:"enabled"`
	// SortOrder 排序值（越小越靠前）。
	SortOrder int `json:"sortOrder"`
	// Clicks 点击次数（公开接口上报 +1）。
	Clicks    int    `json:"clicks"`
	CreatedAt string `json:"createdAt"`
}

// AdminUser 管理员账号（单管理员）。
type AdminUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

// LoginResult 登录成功返回的凭证（Access + Refresh + 用户信息）。
type LoginResult struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         AdminUser `json:"user"`
}

// AdminLoginLog 管理员登录日志（成功/失败，供审计）。
// 注意：绝不记录密码、Token 等敏感信息。
type AdminLoginLog struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Action    string `json:"action"` // success / fail / logout / refresh
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	CreatedAt string `json:"createdAt"`
}

// ExecEvent 执行记录的实时推送事件（WebSocket）。
type ExecEvent struct {
	Type     string     `json:"type"` // snapshot | log | progress | status
	ExecID   string     `json:"execId"`
	Progress int        `json:"progress,omitempty"` // 0-100
	Step     string     `json:"step,omitempty"`     // 当前步骤描述
	Status   TaskStatus `json:"status,omitempty"`
	Log      *LogLine   `json:"log,omitempty"`      // log 事件的日志行
	Logs     []LogLine  `json:"logs,omitempty"`     // snapshot 携带的历史日志
	Duration string     `json:"duration,omitempty"` // snapshot 携带
}

// DataStatus 抓取数据审核状态（官网技能表的 data_status 列）。
type DataStatus string

const (
	DataPending   DataStatus = "pending"   // 待审核
	DataApproved  DataStatus = "approved"  // 已审核
	DataPublished DataStatus = "published" // 已发布（官网可见）
	DataIgnored   DataStatus = "ignored"   // 已忽略
)

// TrendPoint 近 N 天执行趋势中的一个点。
type TrendPoint struct {
	Day     string `json:"day"`
	Count   int    `json:"count"`
	Success int    `json:"success"`
}

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
	// Trend 近 7 天执行趋势（真实统计）。
	Trend []TrendPoint `json:"trend"`
	// StatusDist 数据状态分布（pending/approved/published/ignored）。
	StatusDist map[string]int `json:"statusDist"`
	// TypeDist 数据来源分布（official/community）。
	TypeDist map[string]int `json:"typeDist"`
}

// OfficialOrg 官方组织（动态管理）。
// 爬虫识别官方来源的依据：仓库 owner 在此表内 → 官方技能。
// 可在后台增删改，无需改代码。
type OfficialOrg struct {
	Owner       string `json:"owner"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	LogoURL     string `json:"logoUrl"`
	SortOrder   int    `json:"sortOrder"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"createdAt"`
}

// OfficialOrgSummary 官网首页「官方技能/官方组织」区块的统一数据源：
// 官方组织及其官方技能数（按展示名去重）。
type OfficialOrgSummary struct {
	Owner         string `json:"owner"`
	DisplayName   string `json:"displayName"`
	Avatar        string `json:"avatar"`
	LogoURL       string `json:"logoUrl"`
	OfficialCount int    `json:"officialCount"`
}

// OrgVerifyResult 官方组织 GitHub 校验结果（后台一键校验）。
// GitHubType：Organization=正确组织 / User=个人账号（应修正或删除） / NotFound=不存在 / Error=校验失败。
// AvatarOK：Organization 且头像有效（非默认 identicon / 纯色块）时为 true；
// 头像无效说明该组织未设置品牌 logo，建议在后台 Logo URL 填写官网 logo。
type OrgVerifyResult struct {
	Owner       string `json:"owner"`
	DisplayName string `json:"displayName"`
	GitHubType  string `json:"githubType"`
	AvatarOK    bool   `json:"avatarOk"`
	LogoURL     string `json:"logoUrl"`
}
