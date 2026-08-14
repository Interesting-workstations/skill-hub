/**
 * 中英文切换 —— 基于官方 i18next + react-i18next。
 *
 * - 语言资源：web/src/i18n/locales/{zh,en}.ts（标准 i18next 资源文件）
 * - 插值使用单花括号 {n}/{author}（init 时配置 interpolation prefix/suffix）
 * - 语言选择持久化到 localStorage（key: skillhub-lang），并同步 <html lang>
 *
 * 用法（与之前自研版本 API 完全一致，组件无需改动）：
 *   const { lang, setLang, toggleLang, t } = useI18n();
 *   t("nav.submit")                    // 普通文案
 *   t("detail.moreFrom", { author })   // {author} 插值
 */
import type { ReactNode } from "react";
import i18n from "i18next";
import { initReactI18next, useTranslation } from "react-i18next";
import zh from "./locales/zh";
import en from "./locales/en";

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

function loadInitialLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return saved === "en" ? "en" : "zh";
  } catch {
    return "zh";
  }
}

// 初始化官方 i18next（浏览器侧全局单例）
void i18n.use(initReactI18next).init({
  resources: {
    zh: { translation: zh },
    en: { translation: en },
  },
  lng: loadInitialLang(),
  fallbackLng: "zh",
  // 与既有字典兼容：使用单花括号插值（官方默认是 {{var}}）
  interpolation: {
    escapeValue: false,
    prefix: "{",
    suffix: "}",
  },
});

// 语言变化：持久化 + 同步 <html lang>
i18n.on("languageChanged", (lng: string) => {
  try {
    localStorage.setItem(STORAGE_KEY, lng);
  } catch {
    /* 忽略隐私模式等异常 */
  }
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
