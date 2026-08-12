/** 爬虫模块 Mock 数据 —— 后端爬虫任务 API 尚未实现，先用本地数据模拟。 */

import type {
  CrawlTask,
  ExecutionRecord,
  FailureRecord,
  CrawlerConfig,
  LogLine,
} from "../../types";

export const mockCrawlTasks: CrawlTask[] = [
  {
    id: "task-1",
    name: "官方技能采集",
    type: "skill",
    query: "anthropics/skills, openai/codex",
    status: "success",
    schedule: "每天 02:00",
    lastRunAt: "2026-08-12 02:00",
    lastDuration: "4m 12s",
    runCount: 26,
    successCount: 24,
    failCount: 2,
    createdAt: "2026-07-01",
  },
  {
    id: "task-2",
    name: "社区技能搜索",
    type: "skill",
    query: "claude skills",
    status: "running",
    schedule: "每小时",
    lastRunAt: "2026-08-12 18:00",
    lastDuration: "--",
    runCount: 132,
    successCount: 128,
    failCount: 4,
    createdAt: "2026-07-05",
  },
  {
    id: "task-3",
    name: "星标榜热更",
    type: "info",
    query: "agent skills stars:>100",
    status: "failed",
    schedule: "每 6 小时",
    lastRunAt: "2026-08-12 12:00",
    lastDuration: "1m 02s",
    runCount: 58,
    successCount: 51,
    failCount: 7,
    createdAt: "2026-07-10",
  },
  {
    id: "task-4",
    name: "新分类探测",
    type: "data",
    query: "skillsets",
    status: "waiting",
    schedule: "手动",
    lastRunAt: "--",
    lastDuration: "--",
    runCount: 0,
    successCount: 0,
    failCount: 0,
    createdAt: "2026-08-11",
  },
  {
    id: "task-5",
    name: "媒体类技能采集",
    type: "media",
    query: "image video skill",
    status: "stopped",
    schedule: "每周一 03:00",
    lastRunAt: "2026-08-05 03:00",
    lastDuration: "2m 40s",
    runCount: 9,
    successCount: 6,
    failCount: 3,
    createdAt: "2026-07-15",
  },
];

function buildLogs(taskName: string, status: string): LogLine[] {
  const logs: LogLine[] = [
    { time: "18:00:01", level: "info", text: `开始执行任务「${taskName}」` },
    { time: "18:00:03", level: "info", text: "初始化爬虫客户端" },
    { time: "18:00:05", level: "info", text: "开始抓取目标仓库" },
    { time: "18:02:10", level: "ok", text: "已抓取 100 条数据" },
    { time: "18:03:40", level: "ok", text: "已抓取 200 条数据" },
    { time: "18:04:55", level: "warn", text: "2 个页面请求超时，已自动重试" },
  ];
  if (status === "running") {
    logs.push({ time: "18:05:30", level: "info", text: "抓取进行中…" });
  } else if (status === "success") {
    logs.push(
      { time: "18:05:58", level: "ok", text: "任务执行完成，共 287 条数据" },
    );
  } else {
    logs.push(
      { time: "18:05:40", level: "err", text: "检测到 4 个失败页面（连接被拒绝）" },
      { time: "18:05:58", level: "err", text: "任务执行失败，已记录失败详情" },
    );
  }
  return logs;
}

export const mockExecutions: ExecutionRecord[] = [
  {
    id: "exec-1",
    taskId: "task-2",
    taskName: "社区技能搜索",
    status: "running",
    startTime: "2026-08-12 18:00:01",
    endTime: "--",
    duration: "进行中",
    progress: 68,
    stats: { pages: 320, fetched: 218, failed: 0, newData: 176, updated: 42, duplicate: 0 },
    logs: buildLogs("社区技能搜索", "running"),
  },
  {
    id: "exec-2",
    taskId: "task-1",
    taskName: "官方技能采集",
    status: "success",
    startTime: "2026-08-12 02:00:01",
    endTime: "2026-08-12 02:04:12",
    duration: "4m 11s",
    progress: 100,
    stats: { pages: 64, fetched: 64, failed: 0, newData: 12, updated: 52, duplicate: 8 },
    logs: buildLogs("官方技能采集", "success"),
  },
  {
    id: "exec-3",
    taskId: "task-3",
    taskName: "星标榜热更",
    status: "failed",
    startTime: "2026-08-12 12:00:03",
    endTime: "2026-08-12 12:01:05",
    duration: "1m 02s",
    progress: 100,
    stats: { pages: 40, fetched: 36, failed: 4, newData: 18, updated: 0, duplicate: 18 },
    logs: buildLogs("星标榜热更", "failed"),
  },
  {
    id: "exec-4",
    taskId: "task-5",
    taskName: "媒体类技能采集",
    status: "stopped",
    startTime: "2026-08-05 03:00:02",
    endTime: "2026-08-05 03:02:42",
    duration: "2m 40s",
    progress: 100,
    stats: { pages: 90, fetched: 52, failed: 3, newData: 30, updated: 10, duplicate: 12 },
    logs: buildLogs("媒体类技能采集", "stopped"),
  },
];

export const mockFailures: FailureRecord[] = [
  {
    id: "fail-1",
    taskId: "task-3",
    taskName: "星标榜热更",
    url: "https://github.com/octocat/Hello-World",
    reason: "连接被拒绝",
    error: "Get \"https://api.github.com/repos/octocat/Hello-World\": dial tcp: connection refused",
    retryCount: 3,
    failedAt: "2026-08-12 12:01:03",
  },
  {
    id: "fail-2",
    taskId: "task-3",
    taskName: "星标榜热更",
    url: "https://github.com/org/repo-with-404",
    reason: "仓库不存在 (404)",
    error: "GitHub API /repos/org/repo-with-404: 404 Not Found",
    retryCount: 1,
    failedAt: "2026-08-12 12:00:58",
  },
  {
    id: "fail-3",
    taskId: "task-5",
    taskName: "媒体类技能采集",
    url: "https://github.com/org/media-skill",
    reason: "SKILL.md 解析失败",
    error: "unexpected end of JSON input while parsing frontmatter",
    retryCount: 0,
    failedAt: "2026-08-05 03:02:30",
  },
  {
    id: "fail-4",
    taskId: "task-2",
    taskName: "社区技能搜索",
    url: "https://github.com/org/timeout-repo",
    reason: "请求超时",
    error: "context deadline exceeded (Client.Timeout exceeded while awaiting headers)",
    retryCount: 2,
    failedAt: "2026-08-12 18:04:55",
  },
];

export const mockCrawlerConfig: CrawlerConfig = {
  concurrency: 4,
  timeout: 20,
  retryCount: 3,
  userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) SkillHubBot/1.0",
  requestInterval: 500,
  maxPagesPerRun: 500,
  officialRepos: "anthropics/skills,openai/codex,vercel/ai",
  defaultQuery: "claude skills",
  enabled: true,
};
