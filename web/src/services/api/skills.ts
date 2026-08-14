// 技能资源库 API 模块：封装后端接口，组件只依赖这些函数。
import { request, API_BASE_URL } from "./client";
import type { Article, Author, AuthorDetail, Category, SeoConfig, SiteConfig, Skill, Stats } from "../../data/types";

export interface SkillQuery {
  category?: string;
  author?: string;
  official?: boolean;
  featured?: boolean;
}

/** GET /api/v1/skills */
export function fetchSkills(query: SkillQuery = {}): Promise<Skill[]> {
  const qs = new URLSearchParams();
  if (query.category) qs.set("category", query.category);
  if (query.author) qs.set("author", query.author);
  if (query.official) qs.set("official", "true");
  if (query.featured) qs.set("featured", "true");
  const q = qs.toString();
  return request<Skill[]>(`/skills${q ? `?${q}` : ""}`);
}

/** GET /api/v1/skills/:id */
export function fetchSkill(id: string): Promise<Skill> {
  return request<Skill>(`/skills/${encodeURIComponent(id)}`);
}

/** GET /api/v1/skills/:id/download —— 技能 ZIP 下载地址（后端动态生成） */
export function skillDownloadUrl(id: string): string {
  return `${API_BASE_URL}/skills/${encodeURIComponent(id)}/download`;
}

/** GET /api/v1/authors */
export function fetchAuthors(): Promise<Author[]> {
  return request<Author[]>("/authors");
}

/** GET /api/v1/authors/:slug */
export function fetchAuthor(slug: string): Promise<AuthorDetail> {
  return request<AuthorDetail>(`/authors/${encodeURIComponent(slug)}`);
}

/** GET /api/v1/categories */
export function fetchCategories(): Promise<Category[]> {
  return request<Category[]>("/categories");
}

/** GET /api/v1/categories/:slug */
export function fetchCategory(slug: string): Promise<Category> {
  return request<Category>(`/categories/${encodeURIComponent(slug)}`);
}

/** GET /api/v1/stats */
export function fetchStats(): Promise<Stats> {
  return request<Stats>("/stats");
}

/* ---------- 公开内容（文章 / 站点配置 / SEO / 提交技能） ---------- */

export interface SubmitSkillInput {
  name: string;
  author: string;
  description: string;
  category?: string;
  tags?: string[];
  downloadUrl?: string;
  githubUrl?: string;
}

/** GET /api/v1/articles */
export function fetchArticles(): Promise<Article[]> {
  return request<Article[]>("/articles");
}

/** GET /api/v1/articles/:id */
export function fetchArticle(id: string): Promise<Article> {
  return request<Article>(`/articles/${encodeURIComponent(id)}`);
}

/** GET /api/v1/site-config */
export function fetchSiteConfig(): Promise<SiteConfig> {
  return request<SiteConfig>("/site-config");
}

/** GET /api/v1/seo */
export function fetchSeo(): Promise<SeoConfig> {
  return request<SeoConfig>("/seo");
}

/** POST /api/v1/skills/submit —— 提交技能 */
export function submitSkill(input: SubmitSkillInput): Promise<Skill> {
  return request<Skill>("/skills/submit", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}
