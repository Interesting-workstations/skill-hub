/** 内容 / 数据管理 API —— Mock 实现。 */

import { mockArticles, mockCrawledData, mockSeo, mockSiteConfig } from "../core/mock/content";
import type { Article, CrawledDataItem, DataStatus, SeoConfig, SiteConfig } from "../types";

const delay = (ms = 200) => new Promise((r) => setTimeout(r, ms));

let articles: Article[] = [...mockArticles];
let crawledData: CrawledDataItem[] = [...mockCrawledData];
let seo: SeoConfig = { ...mockSeo };
let siteConfig: SiteConfig = { ...mockSiteConfig };

export const contentApi = {
  /* ---- 文章 ---- */
  async listArticles(): Promise<Article[]> {
    await delay();
    return [...articles];
  },
  async createArticle(input: Omit<Article, "id" | "views" | "updatedAt">): Promise<Article> {
    await delay();
    const art: Article = { ...input, id: `art-${Date.now()}`, views: 0, updatedAt: new Date().toISOString().slice(0, 10) };
    articles = [art, ...articles];
    return art;
  },
  async deleteArticle(id: string): Promise<void> {
    await delay();
    articles = articles.filter((a) => a.id !== id);
  },

  /* ---- 抓取数据 ---- */
  async listData(status?: DataStatus): Promise<CrawledDataItem[]> {
    await delay();
    return status ? crawledData.filter((d) => d.status === status) : [...crawledData];
  },
  async updateDataStatus(id: string, status: DataStatus): Promise<void> {
    await delay();
    crawledData = crawledData.map((d) => (d.id === id ? { ...d, status } : d));
  },
  async deleteData(id: string): Promise<void> {
    await delay();
    crawledData = crawledData.filter((d) => d.id !== id);
  },

  /* ---- SEO / 站点 ---- */
  async getSeo(): Promise<SeoConfig> {
    await delay();
    return { ...seo };
  },
  async saveSeo(patch: Partial<SeoConfig>): Promise<SeoConfig> {
    await delay();
    seo = { ...seo, ...patch };
    return { ...seo };
  },
  async getSiteConfig(): Promise<SiteConfig> {
    await delay();
    return { ...siteConfig };
  },
  async saveSiteConfig(patch: Partial<SiteConfig>): Promise<SiteConfig> {
    await delay();
    siteConfig = { ...siteConfig, ...patch };
    return { ...siteConfig };
  },
};
