/** 内容 / 数据管理 API —— 对接 skill-hub 后端（/api/v1/admin），全部由 Go 提供。 */

import { http } from "../core/http";
import type { Article, CrawledDataItem, DataStatus, SeoConfig, SiteConfig, Sponsor } from "../types";

/** 赞助商表单输入（不含 id/clicks/createdAt 等后端生成字段） */
export type SponsorInput = Pick<
  Sponsor,
  "name" | "logo" | "descriptionZh" | "descriptionEn" | "url" | "position" | "enabled" | "sortOrder"
>;

/** 导出结果（后端生成文本内容） */
export interface ExportResult {
  filename: string;
  content: string;
  format: string;
  count: number;
}

const base = "/api/v1/admin";

/** 抓取数据筛选条件（审核页） */
export interface DataFilter {
  status?: DataStatus | "";
  /** official=官方来源 / community=社区个人 */
  source?: "official" | "community" | "";
  category?: string;
  author?: string;
  q?: string;
  /** stars=高星优先 / name / newest */
  sort?: "stars" | "name" | "newest";
}

export const contentApi = {
  /* ---- 文章 ---- */
  listArticles(): Promise<Article[]> {
    return http.get<Article[]>(`${base}/articles`);
  },
  createArticle(input: Pick<Article, "title" | "category" | "content">): Promise<Article> {
    return http.post<Article>(`${base}/articles`, input);
  },
  updateArticle(id: string, input: Pick<Article, "title" | "category" | "status" | "content">): Promise<Article> {
    return http.put<Article>(`${base}/articles/${encodeURIComponent(id)}`, input);
  },
  deleteArticle(id: string): Promise<void> {
    return http.delete<void>(`${base}/articles/${encodeURIComponent(id)}`);
  },

  /* ---- 数据导出 ---- */
  exportData(format: "json" | "csv" | "markdown", scope?: string): Promise<ExportResult> {
    const qs = new URLSearchParams({ format });
    if (scope) qs.set("scope", scope);
    return http.get<ExportResult>(`${base}/export?${qs.toString()}`);
  },

  /* ---- 抓取数据 ---- */
  listData(filter: DataFilter = {}): Promise<CrawledDataItem[]> {
    const qs = new URLSearchParams();
    if (filter.status) qs.set("status", filter.status);
    if (filter.source) qs.set("source", filter.source);
    if (filter.category) qs.set("category", filter.category);
    if (filter.author) qs.set("author", filter.author);
    if (filter.q) qs.set("q", filter.q);
    if (filter.sort) qs.set("sort", filter.sort);
    const str = qs.toString();
    return http.get<CrawledDataItem[]>(`${base}/data${str ? `?${str}` : ""}`);
  },
  updateDataStatus(id: string, status: DataStatus): Promise<void> {
    return http.put<void>(`${base}/data/${encodeURIComponent(id)}/status`, { status });
  },
  /** 批量更新状态（全选后一键通过/忽略/删除） */
  batchUpdateDataStatus(ids: string[], status: DataStatus): Promise<void> {
    return http.post<void>(`${base}/data/batch-status`, { ids, status });
  },
  /** 机器人自动审核：内容完整规范且无重复的直接通过并发布，有问题的留人工 */
  autoAuditData(): Promise<{ total: number; approved: number; manual: number }> {
    return http.post<{ total: number; approved: number; manual: number }>(`${base}/data/auto-audit`);
  },
  deleteData(id: string): Promise<void> {
    return http.delete<void>(`${base}/data/${encodeURIComponent(id)}`);
  },

  /* ---- SEO / 站点 ---- */
  getSeo(): Promise<SeoConfig> {
    return http.get<SeoConfig>(`${base}/seo`);
  },
  saveSeo(config: SeoConfig): Promise<SeoConfig> {
    return http.put<SeoConfig>(`${base}/seo`, config);
  },
  getSiteConfig(): Promise<SiteConfig> {
    return http.get<SiteConfig>(`${base}/site-config`);
  },
  saveSiteConfig(config: SiteConfig): Promise<SiteConfig> {
    return http.put<SiteConfig>(`${base}/site-config`, config);
  },

  /* ---- 赞助商 ---- */
  listSponsors(): Promise<Sponsor[]> {
    return http.get<Sponsor[]>(`${base}/sponsors`);
  },
  createSponsor(input: SponsorInput): Promise<Sponsor> {
    return http.post<Sponsor>(`${base}/sponsors`, input);
  },
  updateSponsor(id: string, input: SponsorInput): Promise<Sponsor> {
    return http.put<Sponsor>(`${base}/sponsors/${encodeURIComponent(id)}`, input);
  },
  deleteSponsor(id: string): Promise<void> {
    return http.delete<void>(`${base}/sponsors/${encodeURIComponent(id)}`);
  },
};
