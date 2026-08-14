/**
 * 中英文切换 —— 基于官方 i18next + react-i18next。
 *
 * - 语言包：public/locales/{zh,en}/translation.json（可独立下载 / 更新）
 * - 加载：官方 i18next-http-backend 按需 HTTP 下载语言包
 * - 检测 / 持久化：官方 i18next-browser-languagedetector
 *   （优先读 localStorage key=skillhub-lang，其次浏览器语言；切换后写回 localStorage）
 * - 插值使用单花括号 {n}/{author}（init 时配置 interpolation prefix/suffix）
 *
 * 用法（组件 API 不变）：
 *   const { lang, setLang, toggleLang, t } = useI18n();
 *   t("nav.submit")                    // 普通文案
 *   t("detail.moreFrom", { author })   // {author} 插值
 */
import type { ReactNode } from "react";
import i18n from "i18next";
import { initReactI18next, useTranslation } from "react-i18next";
import HttpBackend from "i18next-http-backend";
import LanguageDetector from "i18next-browser-languagedetector";

export type Lang = "zh" | "en";

const STORAGE_KEY = "skillhub-lang";

export interface I18nValue {
  lang: Lang;
  /** 切换到指定语言 */
  setLang: (l: Lang) => void;
  /** 切换语言（zh ⇄ en） */
  toggleLang: () => void;
  /** 翻译：key 不存在时原样返回 key；支持 {var} 插值 */
  t: (key: string, vars?: Record<string, string | number>) => string;
}

// 初始化官方 i18next：HTTP 下载语言包 + 语言检测/持久化
void i18n
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    supportedLngs: ["zh", "en"],
    fallbackLng: "zh",
    // 语言包下载路径：public/locales/{lng}/translation.json
    backend: {
      loadPath: "/locales/{{lng}}/translation.json",
    },
    detection: {
      // 先读 localStorage 已选语言，其次浏览器语言；切换后写回 localStorage
      order: ["localStorage", "navigator"],
      caches: ["localStorage"],
      lookupLocalStorage: STORAGE_KEY,
    },
    // 与既有字典兼容：使用单花括号插值（官方默认是 {{var}}）
    interpolation: {
      escapeValue: false,
      prefix: "{",
      suffix: "}",
    },
    // 语言包异步下载期间不启用 Suspense，避免首屏白屏
    react: {
      useSuspense: false,
    },
  });

// 语言变化：同步 <html lang>（localStorage 持久化由 LanguageDetector 负责）
i18n.on("languageChanged", (lng: string) => {
  document.documentElement.lang = lng;
});

export function I18nProvider({ children }: { children: ReactNode }) {
  return <>{children}</>;
}

/** 读取当前语言的便捷 hook（封装官方 useTranslation，保持既有 API） */
export function useI18n(): I18nValue {
  const { t, i18n: i18nInstance } = useTranslation();
  const lang: Lang = i18nInstance.resolvedLanguage === "en" ? "en" : "zh";

  return {
    lang,
    setLang: (l: Lang) => {
      void i18nInstance.changeLanguage(l);
    },
    toggleLang: () => {
      void i18nInstance.changeLanguage(lang === "zh" ? "en" : "zh");
    },
    t: ((key: string, vars?: Record<string, string | number>) =>
      t(key, vars) as string) as I18nValue["t"],
  };
}
