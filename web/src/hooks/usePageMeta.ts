import { useEffect } from "react";
import { site } from "../config/site";
import { fetchSeo } from "../services/api/skills";
import type { SeoConfig } from "../data/types";

interface PageMeta {
  title: string;
  description?: string;
}

/** SEO 默认值缓存：页面未显式传 description 时，用后台配置的 SEO 描述兜底。 */
let seoCache: SeoConfig | null | undefined;
function loadSeo(): SeoConfig | null {
  if (seoCache !== undefined) return seoCache;
  seoCache = null;
  fetchSeo()
    .then((s) => {
      seoCache = s;
    })
    .catch(() => {});
  return seoCache;
}

function setMeta(name: string, content: string) {
  let el = document.head.querySelector(`meta[name="${name}"]`);
  if (!el) {
    el = document.createElement("meta");
    el.setAttribute("name", name);
    document.head.appendChild(el);
  }
  el.setAttribute("content", content);
}

/**
 * 轻量 SEO：设置页面 title / description（无需第三方库）。
 * 路由切换时由下一页覆盖；卸载时恢复站点默认标题（后台 SEO 配置优先）。
 */
export function usePageMeta({ title, description }: PageMeta) {
  useEffect(() => {
    document.title = title;
    if (description) {
      setMeta("description", description);
    } else {
      const seo = loadSeo();
      if (seo?.description) setMeta("description", seo.description);
    }
    const fallbackTitle = loadSeo()?.title || site.title;
    return () => {
      document.title = fallbackTitle;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [title, description]);
}
