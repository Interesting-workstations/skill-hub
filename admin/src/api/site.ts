/** 官网数据 API —— 对接 skill-hub 后端（/api/v1）。 */

import { http } from "../core/http";
import { site } from "../config/site";
import type { Skill, Category, Author, Stats } from "../types";

const base = site.apiBase;

export const siteApi = {
  /** 站点聚合统计 */
  stats(): Promise<Stats> {
    return http.get<Stats>(`${base}/stats`);
  },

  /** 技能列表（支持 featured / official / category / author 筛选） */
  skills(params: {
    featured?: boolean;
    official?: boolean;
    category?: string;
    author?: string;
  } = {}): Promise<Skill[]> {
    const qs = new URLSearchParams();
    if (params.featured) qs.set("featured", "true");
    if (params.official) qs.set("official", "true");
    if (params.category) qs.set("category", params.category);
    if (params.author) qs.set("author", params.author);
    const q = qs.toString();
    return http.get<Skill[]>(`${base}/skills${q ? `?${q}` : ""}`);
  },

  /** 分类列表（含真实数量） */
  categories(): Promise<Category[]> {
    return http.get<Category[]>(`${base}/categories`);
  },

  /** 作者列表（含官方技能数） */
  authors(): Promise<Author[]> {
    return http.get<Author[]>(`${base}/authors`);
  },
};
