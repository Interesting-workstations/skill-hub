// 技能资源库数据 Hooks：页面/组件通过 Hook 获取数据，不直接调用 API。
import type {
  Author,
  AuthorDetail,
  Category,
  Skill,
  Stats,
} from "../data/types";
import {
  fetchAuthor,
  fetchAuthors,
  fetchCategories,
  fetchCategory,
  fetchSkill,
  fetchSkills,
  fetchStats,
  type SkillQuery,
} from "../services/api/skills";
import { useAsyncData, type AsyncState } from "./useAsyncData";

/** 站点统计数据 */
export function useStats(): AsyncState<Stats> {
  return useAsyncData(fetchStats, []);
}

/**
 * 技能列表（支持筛选）。
 * query 传 null 表示不请求（用于数据尚未就绪的场景）。
 */
export function useSkills(query: SkillQuery | null = {}): AsyncState<Skill[]> {
  return useAsyncData(
    () => (query ? fetchSkills(query) : Promise.resolve([])),
    [query?.category, query?.author, query?.official, query?.featured]
  );
}

/** 技能详情；id 未就绪时不请求 */
export function useSkill(id: string | undefined): AsyncState<Skill> {
  return useAsyncData(
    () => (id ? fetchSkill(id) : Promise.resolve(null as unknown as Skill)),
    [id]
  );
}

/** 全部作者 */
export function useAuthors(): AsyncState<Author[]> {
  return useAsyncData(fetchAuthors, []);
}

/** 作者详情（含技能）；slug 未就绪时不请求 */
export function useAuthor(slug: string | undefined): AsyncState<AuthorDetail> {
  return useAsyncData(
    () => (slug ? fetchAuthor(slug) : Promise.resolve(null as unknown as AuthorDetail)),
    [slug]
  );
}

/** 全部分类 */
export function useCategories(): AsyncState<Category[]> {
  return useAsyncData(fetchCategories, []);
}

/** 分类详情（含技能）；slug 未就绪时不请求 */
export function useCategory(slug: string | undefined): AsyncState<Category> {
  return useAsyncData(
    () => (slug ? fetchCategory(slug) : Promise.resolve(null as unknown as Category)),
    [slug]
  );
}
