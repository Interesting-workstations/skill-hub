/** 内容 / 数据管理 API —— 对接 skill-hub 后端（/api/v1/admin），全部由 Go 提供。 */

import { http } from "../core/http";
import type { Article, CrawledDataItem, DataStatus, SeoConfig, SiteConfig } from "../types";

/** 导出结果（后端生成文本内容） */
export interface ExportResult {
  filename: string;
  content: string;
  format: string;
  count: number;
}

const base = "/api/v1/admin";

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
  listData(status?: DataStatus): Promise<CrawledDataItem[]> {
    return http.get<CrawledDataItem[]>(`${base}/data${status ? `?status=${status}` : ""}`);
  },
  updateDataStatus(id: string, status: DataStatus): Promise<void> {
    return http.put<void>(`${base}/data/${encodeURIComponent(id)}/status`, { status });
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
};
