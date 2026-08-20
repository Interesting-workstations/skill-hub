/** 爬虫管理 API —— 对接 skill-hub 后端（/api/v1/admin），数据全部由 Go 提供。 */

import { http, ADMIN_API_BASE } from "../core/http";
import type { AdminStats, CrawlTask, CrawlerConfig, ExecutionRecord, FailureRecord, GitHubToken, OfficialOrg, OrgLogoRefreshResult, OrgVerifyResult, ScanResult, TokenHealth, TranslateConfig, TranslateTestResult, TranslationItem } from "../types";

const base = ADMIN_API_BASE;

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

  /** 重新拉取所有官方组织 logo 图片到本地缓存（GitHub 图片不稳定时一键刷新） */
  refreshOrgLogos(): Promise<OrgLogoRefreshResult[]> {
    return http.post<OrgLogoRefreshResult[]>(`${base}/official-orgs/refresh-logos`);
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

  /** GitHub Token 池 */
  listTokens(): Promise<GitHubToken[]> {
    return http.get<GitHubToken[]>(`${base}/tokens`);
  },
  createToken(input: { token: string; remark?: string }): Promise<GitHubToken> {
    return http.post<GitHubToken>(`${base}/tokens`, input);
  },
  updateToken(id: string, patch: { token?: string; remark?: string; enabled?: boolean }): Promise<void> {
    return http.put<void>(`${base}/tokens/${encodeURIComponent(id)}`, patch);
  },
  deleteToken(id: string): Promise<void> {
    return http.delete<void>(`${base}/tokens/${encodeURIComponent(id)}`);
  },

  /** GitHub Token 池一键检测（tokens 为空时检测当前配置） */
  checkTokens(tokens?: string[]): Promise<TokenHealth[]> {
    return http.post<TokenHealth[]>(`${base}/tokens/check`, { tokens: tokens ?? [] });
  },

  /** 翻译管理：扫描未汉化技能 */
  scanUntranslated(limit?: number): Promise<ScanResult> {
    const qs = limit ? `?limit=${limit}` : "";
    return http.get<ScanResult>(`${base}/translate/scan${qs}`);
  },
  /** 翻译单条技能 */
  translateSkill(id: string): Promise<TranslationItem> {
    return http.post<TranslationItem>(`${base}/translate/${encodeURIComponent(id)}`);
  },
  /** 批量翻译所有未汉化技能 */
  translateAll(): Promise<{ translated: number }> {
    return http.post<{ translated: number }>(`${base}/translate/all`);
  },

  /** 翻译通道配置状态 */
  getTranslateConfig(): Promise<TranslateConfig> {
    return http.get<TranslateConfig>(`${base}/translate/config`);
  },
  /** 设置主翻译通道（auto/tencent/baidu/google/deepl） */
  saveTranslateConfig(provider: string): Promise<TranslateConfig> {
    return http.put<TranslateConfig>(`${base}/translate/config`, { provider });
  },
  /** 测试翻译通道连通性（tencent/baidu/google/deepl/all） */
  testTranslateProvider(provider: string): Promise<{ results: TranslateTestResult[] }> {
    return http.post<{ results: TranslateTestResult[] }>(`${base}/translate/test`, { provider });
  },
};
