/** 爬虫管理 API —— 对接 skill-hub 后端（/api/v1/admin），数据全部由 Go 提供。 */

import { http } from "../core/http";
import type { AdminStats, CrawlTask, CrawlerConfig, ExecutionRecord, FailureRecord } from "../types";

const base = "/api/v1/admin";

export const crawlerApi = {
  /** 工作台统计 */
  stats(): Promise<AdminStats> {
    return http.get<AdminStats>(`${base}/stats`);
  },

  /** 任务列表 */
  listTasks(): Promise<CrawlTask[]> {
    return http.get<CrawlTask[]>(`${base}/tasks`);
  },

  /** 新建任务 */
  createTask(input: Pick<CrawlTask, "name" | "type" | "query" | "schedule">): Promise<CrawlTask> {
    return http.post<CrawlTask>(`${base}/tasks`, input);
  },

  /** 更新任务 */
  updateTask(id: string, patch: Partial<CrawlTask>): Promise<void> {
    return http.put<void>(`${base}/tasks/${encodeURIComponent(id)}`, patch);
  },

  /** 删除任务 */
  deleteTask(id: string): Promise<void> {
    return http.delete<void>(`${base}/tasks/${encodeURIComponent(id)}`);
  },

  /** 执行任务（真实爬虫，后台异步，轮询执行记录） */
  runTask(id: string): Promise<ExecutionRecord> {
    return http.post<ExecutionRecord>(`${base}/tasks/${encodeURIComponent(id)}/run`);
  },

  /** 停止任务 */
  stopTask(id: string): Promise<void> {
    return http.post<void>(`${base}/tasks/${encodeURIComponent(id)}/stop`);
  },

  /** 执行记录列表 */
  listExecutions(): Promise<ExecutionRecord[]> {
    return http.get<ExecutionRecord[]>(`${base}/executions`);
  },

  /** 执行记录详情（轮询直到完成） */
  executionDetail(id: string): Promise<ExecutionRecord> {
    return http.get<ExecutionRecord>(`${base}/executions/${encodeURIComponent(id)}`);
  },

  /** 失败任务列表 */
  listFailures(): Promise<FailureRecord[]> {
    return http.get<FailureRecord[]>(`${base}/failures`);
  },

  /** 处理失败任务（重试 / 忽略均标记处理） */
  retryFailure(id: string): Promise<void> {
    return http.delete<void>(`${base}/failures/${encodeURIComponent(id)}`);
  },
  ignoreFailure(id: string): Promise<void> {
    return http.delete<void>(`${base}/failures/${encodeURIComponent(id)}`);
  },

  /** 爬虫配置 */
  getConfig(): Promise<CrawlerConfig> {
    return http.get<CrawlerConfig>(`${base}/config`);
  },
  saveConfig(config: CrawlerConfig): Promise<CrawlerConfig> {
    return http.put<CrawlerConfig>(`${base}/config`, config);
  },
};
