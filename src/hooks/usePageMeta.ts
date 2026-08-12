import { useEffect } from "react";
import { site } from "../config/site";

interface PageMeta {
  title: string;
  description?: string;
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
 * 路由切换时由下一页覆盖；卸载时恢复站点默认标题。
 */
export function usePageMeta({ title, description }: PageMeta) {
  useEffect(() => {
    document.title = title;
    if (description) {
      setMeta("description", description);
    }
    return () => {
      document.title = site.title;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [title, description]);
}
