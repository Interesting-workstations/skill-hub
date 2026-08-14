/** 领域类型定义：官网数据（对接 skill-hub 后端）+ 后台运营数据（Mock）。 */

/* ---------- 官网数据（skill-hub 后端 API） ---------- */
export interface Skill {
  id: string;
  name: string;
  author: string;
  description: string;
  tags: string[];
  category: string;
  downloadUrl: string;
  isOfficial?: boolean;
  isFeatured?: boolean;
  installCommand?: string;
  githubUrl?: string;
  githubStars?: string;
  license?: string;
  content?: { heading: string; body: string[] }[];
}

export interface Category {
  name: string;
  slug: string;
  count: number;
  skills: Skill[];
}

export interface Author {
  name: string;
  avatar: string;
  skillCount: number;
  slug: string;
  officialSkills?: number;
}

export interface Stats {
  totalSkills: number;
  totalAuthors: number;
  totalCategories: number;
  officialSkills: number;
  featuredSkills: number;
}

/** 后台工作台聚合指标（由 Go 后端计算） */
export interface AdminStats {
  todayTasks: number;
  runSuccess: number;
  runFailed: number;
  runRunning: number;
  newData: number;
  pendingData: number;
  totalSkills: number;
  officialSkills: number;
  totalAuthors: number;
  totalCategories: number;
  /** 近 7 天执行趋势 */
  trend?: TrendPoint[];
  /** 数据状态分布（pending/approved/published/ignored） */
  statusDist?: Record<string, number>;
  /** 数据来源分布（official/community） */
  typeDist?: Record<string, number>;
}

export interface TrendPoint {
  day: string;
  count: number;
  success: number;
}

/* ---------- 爬虫任务 ---------- */
export type TaskStatus = "waiting" | "running" | "success" | "failed" | "stopped";

export interface CrawlTask {
  id: string;
  name: string;
  /** 任务类型：news/product/info/data/skill 等 */
  type: string;
  /** 搜索关键词或目标仓库 */
  query: string;
  status: TaskStatus;
  /** 计划：cron 或描述文本 */
  schedule: string;
  lastRunAt: string;
  lastDuration?: string;
  runCount: number;
  successCount: number;
  failCount: number;
  createdAt: string;
}

/* ---------- 执行记录 ---------- */
export interface ExecutionStats {
  pages: number;
  fetched: number;
  failed: number;
  newData: number;
  updated: number;
  duplicate: number;
}

export type LogLevel = "info" | "ok" | "warn" | "err";

export interface LogLine {
  time: string;
  level: LogLevel;
  text: string;
}

export interface ExecutionRecord {
  id: string;
  taskId: string;
  taskName: string;
  status: TaskStatus;
  startTime: string;
  endTime: string;
  duration: string;
  progress: number;
  stats: ExecutionStats;
  logs: LogLine[];
}

/* ---------- 失败任务 ---------- */
export interface FailureRecord {
  id: string;
  taskId: string;
  taskName: string;
  url: string;
  reason: string;
  error: string;
  retryCount: number;
  failedAt: string;
}

/* ---------- 爬虫配置 ---------- */
export interface CrawlerConfig {
  concurrency: number;
  timeout: number;
  retryCount: number;
  userAgent: string;
  requestInterval: number;
  maxPagesPerRun: number;
  officialRepos: string;
  defaultQuery: string;
  enabled: boolean;
}

/* ---------- 抓取数据（数据管理，来自 Go 后端 skills 表） ---------- */
export type DataStatus = "pending" | "approved" | "published" | "ignored";

export interface CrawledDataItem {
  id: string;
  name: string;
  author: string;
  category: string;
  githubStars?: string;
  isOfficial?: boolean;
  source: string;
  status: DataStatus;
  fetchedAt?: string;
}

/* ---------- 文章 / SEO ---------- */
export interface Article {
  id: string;
  title: string;
  status: "draft" | "published";
  category: string;
  author: string;
  views: number;
  updatedAt: string;
  /** 文章正文（Markdown） */
  content?: string;
}

export interface SeoConfig {
  title: string;
  description: string;
  keywords: string;
  ogImage: string;
}

/* ---------- 赞助商 ---------- */
export interface Sponsor {
  id: string;
  name: string;
  /** emoji 或图片 URL */
  logo: string;
  descriptionZh: string;
  descriptionEn: string;
  url: string;
  /** home（首页横幅）/ sidebar（详情页侧边栏）/ both */
  position: string;
  enabled: boolean;
  sortOrder: number;
  clicks: number;
  createdAt: string;
}

/* ---------- 系统设置 ---------- */
export interface SiteConfig {
  siteName: string;
  slogan: string;
  portalUrl: string;
  icp: string;
  contactEmail: string;
}

/* ---------- 管理员 ---------- */
export interface AdminUser {
  username: string;
  displayName: string;
  role: "admin";
}

/* ---------- 官方组织（动态管理） ---------- */
export interface OfficialOrg {
  owner: string;
  displayName: string;
  avatar: string;
  sortOrder: number;
  enabled: boolean;
  createdAt: string;
}

/* GitHub 校验结果：owner 是否为真正的组织 + 头像是否有效 */
export interface OrgVerifyResult {
  owner: string;
  displayName: string;
  /** Organization=正确组织 / User=个人账号 / NotFound=不存在 / Error=校验失败 */
  githubType: string;
  /** 组织头像是否有效（非默认 identicon / 纯色块） */
  avatarOk: boolean;
  logoUrl: string;
}
