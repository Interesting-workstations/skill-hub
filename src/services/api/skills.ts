// 技能资源库 API 模块：封装后端接口，组件只依赖这些函数。
import { request } from "./client";
import type { Author, AuthorDetail, Category, Skill, Stats } from "../../data/types";

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
