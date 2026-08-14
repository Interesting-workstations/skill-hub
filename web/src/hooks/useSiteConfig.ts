import { useEffect, useState } from "react";
import { fetchSeo, fetchSiteConfig } from "../services/api/skills";
import type { SeoConfig, SiteConfig } from "../data/types";
import { site } from "../config/site";

/** 动态站点配置（模块级缓存：一次加载，后续复用，避免每次渲染都请求） */
let cache: { siteConfig: SiteConfig | null; seo: SeoConfig | null } | null = null;
let loading: Promise<void> | null = null;

function load(): Promise<void> {
  if (cache) return Promise.resolve();
  if (loading) return loading;
  loading = (async () => {
    const [siteConfig, seo] = await Promise.all([
      fetchSiteConfig().catch(() => null),
      fetchSeo().catch(() => null),
    ]);
    cache = { siteConfig, seo };
  })();
  return loading;
}

export interface SiteMeta {
  /** 站点名（动态配置回退到内置默认） */
  siteName: string;
  /** 站点标语 */
  slogan: string;
  /** ICP 备案号 */
  icp: string;
  /** SEO 默认标题 */
  title: string;
  /** SEO 默认描述 */
  description: string;
}

const defaults: SiteMeta = {
  siteName: site.name,
  slogan: site.description,
  icp: "",
  title: site.title,
  description: site.description,
};

/**
 * 站点动态配置 Hook：首次调用触发加载，返回实时配置（合并内置默认值）。
 * 组件在配置到达后会触发一次重渲染以更新 UI。
 */
export function useSiteConfig(): SiteMeta {
  const [ready, setReady] = useState(!!cache);
  useEffect(() => {
    let alive = true;
    load().then(() => {
      if (alive) setReady(true);
    });
    return () => {
      alive = false;
    };
  }, []);

  void ready; // 缓存命中时 ready 已为 true
  const c = cache;
  if (!c) return defaults;
  const s = c.siteConfig;
  const seo = c.seo;
  return {
    siteName: s?.siteName || defaults.siteName,
    slogan: s?.slogan || defaults.slogan,
    icp: s?.icp || "",
    title: seo?.title || defaults.title,
    description: seo?.description || defaults.description,
  };
}
