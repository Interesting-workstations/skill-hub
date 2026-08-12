/** 爬虫管理 API —— 后端爬虫任务 API 尚未实现，此处为 Mock（模拟异步）。 */

import {
  mockCrawlTasks,
  mockExecutions,
  mockFailures,
  mockCrawlerConfig,
} from "../core/mock/crawler";
import type {
  CrawlTask,
  CrawlerConfig,
  ExecutionRecord,
  FailureRecord,
  TaskStatus,
} from "../types";

const delay = (ms = 300) => new Promise((r) => setTimeout(r, ms));

/** 本地状态（模拟数据库） */
let tasks: CrawlTask[] = [...mockCrawlTasks];
let executions: ExecutionRecord[] = [...mockExecutions];
let failures: FailureRecord[] = [...mockFailures];
let config: CrawlerConfig = { ...mockCrawlerConfig };

function now(): string {
  const d = new Date();
  const p = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())} ${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}`;
}

export const crawlerApi = {
  /** 任务列表 */
  async listTasks(): Promise<CrawlTask[]> {
    await delay();
    return [...tasks];
  },

  /** 新建任务 */
  async createTask(input: Omit<CrawlTask, "id" | "createdAt" | "runCount" | "successCount" | "failCount">): Promise<CrawlTask> {
    await delay();
    const task: CrawlTask = {
      ...input,
      id: `task-${Date.now()}`,
      createdAt: now().slice(0, 10),
      runCount: 0,
      successCount: 0,
      failCount: 0,
    };
    tasks = [task, ...tasks];
    return task;
  },

  /** 更新任务 */
  async updateTask(id: string, patch: Partial<CrawlTask>): Promise<void> {
    await delay();
    tasks = tasks.map((t) => (t.id === id ? { ...t, ...patch } : t));
  },

  /** 执行任务（模拟进度推进） */
  async runTask(id: string): Promise<ExecutionRecord> {
    await delay();
    const task = tasks.find((t) => t.id === id)!;
    const record: ExecutionRecord = {
      id: `exec-${Date.now()}`,
      taskId: id,
      taskName: task.name,
      status: "running",
      startTime: now(),
      endTime: "--",
      duration: "进行中",
      progress: 0,
      stats: { pages: 0, fetched: 0, failed: 0, newData: 0, updated: 0, duplicate: 0 },
      logs: [{ time: now().slice(11), level: "info", text: `开始执行任务「${task.name}」` }],
    };
    executions = [record, ...executions];
    tasks = tasks.map((t) =>
      t.id === id ? { ...t, status: "running", lastRunAt: now().slice(0, 16) } : t,
    );
    return record;
  },

  /** 停止任务 */
  async stopTask(id: string): Promise<void> {
    await delay();
    tasks = tasks.map((t) => (t.id === id ? { ...t, status: "stopped" } : t));
  },

  /** 删除任务 */
  async deleteTask(id: string): Promise<void> {
    await delay();
    tasks = tasks.filter((t) => t.id !== id);
  },

  /** 执行记录列表 */
  async listExecutions(): Promise<ExecutionRecord[]> {
    await delay();
    return [...executions];
  },

  /** 执行记录详情 */
  async executionDetail(id: string): Promise<ExecutionRecord> {
    await delay();
    return executions.find((e) => e.id === id)!;
  },

  /** 失败任务列表 */
  async listFailures(): Promise<FailureRecord[]> {
    await delay();
    return [...failures];
  },

  /** 重试失败任务 */
  async retryFailure(id: string): Promise<void> {
    await delay();
    failures = failures.map((f) => (f.id === id ? { ...f, retryCount: f.retryCount + 1 } : f));
  },

  /** 忽略失败任务 */
  async ignoreFailure(id: string): Promise<void> {
    await delay();
    failures = failures.filter((f) => f.id !== id);
  },

  /** 获取爬虫配置 */
  async getConfig(): Promise<CrawlerConfig> {
    await delay();
    return { ...config };
  },

  /** 保存爬虫配置 */
  async saveConfig(patch: Partial<CrawlerConfig>): Promise<CrawlerConfig> {
    await delay();
    config = { ...config, ...patch };
    return { ...config };
  },

  /** 根据状态筛选任务（供 Dashboard 统计） */
  statusCounts(): { status: TaskStatus; count: number }[] {
    const counts = new Map<TaskStatus, number>();
    for (const t of tasks) {
      counts.set(t.status, (counts.get(t.status) ?? 0) + 1);
    }
    return Array.from(counts, ([status, count]) => ({ status, count }));
  },
};
