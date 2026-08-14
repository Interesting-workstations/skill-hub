/** 爬虫管理 API —— 对接 skill-hub 后端（/api/v1/admin），数据全部由 Go 提供。 */

import { http } from "../core/http";
import type { AdminStats, CrawlTask, CrawlerConfig, ExecutionRecord, FailureRecord, OfficialOrg, OrgVerifyResult } from "../types";

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

  /** 官方组织列表 */
  listOfficialOrgs(): Promise<OfficialOrg[]> {
    return http.get<OfficialOrg[]>(`${base}/official-orgs`);
  },

  /** 新增官方组织 */
  createOfficialOrg(input: Pick<OfficialOrg, "owner" | "displayName" | "avatar" | "logoUrl" | "sortOrder" | "enabled">): Promise<OfficialOrg> {
    return http.post<OfficialOrg>(`${base}/official-orgs`, input);
  },

  /** 更新官方组织 */
  updateOfficialOrg(owner: string, patch: Pick<OfficialOrg, "displayName" | "avatar" | "logoUrl" | "sortOrder" | "enabled">): Promise<void> {
    return http.put<void>(`${base}/official-orgs/${encodeURIComponent(owner)}`, patch);
  },

  /** 删除官方组织 */
  deleteOfficialOrg(owner: string): Promise<void> {
    return http.delete<void>(`${base}/official-orgs/${encodeURIComponent(owner)}`);
  },

  /** 一键校验所有官方组织的 GitHub 类型与头像有效性 */
  verifyOfficialOrgs(): Promise<OrgVerifyResult[]> {
    return http.get<OrgVerifyResult[]>(`${base}/official-orgs/verify`);
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

  /** 停止执行记录（取消运行中的爬虫） */
  stopExecution(id: string): Promise<void> {
    return http.post<void>(`${base}/executions/${encodeURIComponent(id)}/stop`);
  },

  /** 删除执行记录 */
  deleteExecution(id: string): Promise<void> {
    return http.delete<void>(`${base}/executions/${encodeURIComponent(id)}`);
  },

  /** 获取 WebSocket 长连接的一次性票据（避免把 Token 放进 URL） */
  wsTicket(execId: string): Promise<{ ticket: string }> {
    return http.post<{ ticket: string }>(`${base}/ws-ticket`, { execId });
  },

  /** 失败任务列表 */
  listFailures(): Promise<FailureRecord[]> {
    return http.get<FailureRecord[]>(`${base}/failures`);
  },

  /** 处理失败任务 */
  /** 重新执行失败任务（真正触发任务运行） */
  retryFailure(id: string): Promise<void> {
    return http.post<void>(`${base}/failures/${encodeURIComponent(id)}/retry`);
  },
  /** 忽略失败任务（仅删除记录） */
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
