/**
 * 轻量 i18n：无第三方依赖，React Context + 翻译字典 + localStorage 持久化。
 * 用法：
 *   const { lang, setLang, t } = useI18n();
 *   t("nav.submit")                    // 普通文案
 *   t("detail.author", { author })     // {author} 插值
 */
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

export type Lang = "zh" | "en";

const STORAGE_KEY = "skillhub-lang";

interface I18nValue {
  lang: Lang;
  setLang: (l: Lang) => void;
  /** 切换语言（zh ⇄ en） */
  toggleLang: () => void;
  /** 翻译：key 不存在时原样返回 key；支持 {var} 插值 */
  t: (key: string, vars?: Record<string, string | number>) => string;
}

const zh: Record<string, string> = {
  // 品牌
  "brand.name": "Agent Skills 资源库",
  "brand.title": "Agent Skills 资源库 — AI 编程助手的可复用技能",
  "brand.description":
    "发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能。每个技能都是一组指令和代码包，教会你的 AI 助手执行专业任务并自动化复杂工作流。",

  // 导航
  "nav.language": "简体中文",
  "nav.zh": "简体中文",
  "nav.en": "English",
  "nav.submit": "提交",
  "nav.openMenu": "打开菜单",
  "nav.official": "官方技能",
  "nav.featured": "精选技能",
  "nav.categories": "全部分类",
  "nav.submitSkill": "提交技能",

  // 页脚
  "footer.browse": "浏览",
  "footer.official": "官方技能",
  "footer.featured": "精选技能",
  "footer.categories": "全部分类",
  "footer.articles": "文章与教程",
  "footer.contribute": "贡献",
  "footer.submitSkill": "提交技能",

  // 首页 Hero
  "hero.titleMain": "Agent Skills",
  "hero.titleAccent": " 资源库",
  "hero.description":
    "发现适用于 Claude Code、Codex 等 AI 编程助手的可复用技能。每个技能都是一组指令和代码包，教会你的 AI 助手执行专业任务并自动化复杂工作流。",
  "hero.stat.skills": "收录技能",
  "hero.stat.authors": "官方作者",
  "hero.stat.categories": "技能分类",
  "hero.stat.official": "官方技能",
  "hero.cta.browse": "浏览技能",
  "hero.cta.submit": "提交技能",
  "hero.cta.learnMore": "了解更多关于 Skills →",

  // 官方作者区块
  "official.title": "官方技能",
  "official.skillCount": "{n} 个技能",

  // 赞助横幅
  "sponsor.badge": "赞助",

  // 技能卡片
  "card.download": "下载 ZIP",

  // 技能详情
  "detail.notFound": "技能未找到",
  "detail.backHome": "返回首页",
  "detail.authorLabel": "作者：",
  "detail.copyInstall": "复制安装命令",
  "detail.downloadZip": "下载 ZIP",
  "detail.license": "许可：{license}",
  "detail.advertise": "投放广告？",
  "detail.contactUs": "联系我们 →",
  "detail.moreFrom": "来自 {author} 的更多技能",
  "detail.overview": "概述",

  // 面包屑
  "breadcrumb.home": "Agent Skills 资源库",

  // 全部分类
  "categories.title": "全部分类",
  "categories.featured": "精选技能",
  "categories.skillCount": "{n} 个技能",

  // 分类详情
  "category.skillCount": "共 {n} 个技能",
  "category.empty": "该分类暂无技能。",

  // 精选技能
  "featured.title": "精选技能",
  "featured.desc": "编辑精选的高质量技能，共 {n} 个",

  // 官方技能
  "officialPage.title": "官方技能",
  "officialPage.desc": "由官方维护和认证的高质量技能，共 {n} 个",

  // 作者详情
  "author.notFound": "作者未找到",
  "author.backHome": "返回首页",
  "author.skillCount": "共发布 {n} 个技能",
  "author.empty": "该作者暂无已收录的技能。",
  "author.desc": "浏览 {name} 发布的全部技能",

  // 文章
  "articles.title": "文章与教程",
  "articles.desc": "官方发布的技能使用教程与公告",
  "articles.empty": "暂无文章，敬请期待",
  "articles.back": "← 返回文章列表",
  "article.notFound": "文章不存在或已下线",
  "article.noContent": "暂无正文",

  // 404
  "notFound.title": "页面未找到",
  "notFound.desc": "你访问的页面不存在或已被移除。",
  "notFound.backHome": "返回首页",

  // 提交技能
  "submit.title": "提交技能",
  "submit.meta": "分享你创建的 Agent Skill，帮助更多开发者提升 AI 编程体验。提交后进入待审核，审核通过后展示在资源库。",
  "submit.successTitle": "提交成功！",
  "submit.successDesc": "你的技能已提交到待审核队列，管理员审核通过后将出现在资源库中。",
  "submit.backHome": "返回首页",
  "submit.name": "技能名称",
  "submit.namePlaceholder": "例如：Frontend Design",
  "submit.author": "作者 / 维护者",
  "submit.authorPlaceholder": "你的 GitHub 用户名或团队名",
  "submit.githubUrl": "GitHub 仓库地址",
  "submit.githubUrlPlaceholder": "https://github.com/your-username/your-skill",
  "submit.description": "简短描述",
  "submit.descriptionPlaceholder": "用一句话描述这个技能的功能...",
  "submit.category": "分类标签",
  "submit.categoryPlaceholder": "development, testing, creative...",
  "submit.tagsHint": "多个标签用逗号分隔",
  "submit.error": "提交失败，请稍后重试",
  "submit.submitting": "提交中…",

  // 错误边界
  "error.title": "出错了",
  "error.desc": "页面加载出现异常，请刷新重试。",
  "error.reload": "刷新页面",
};

const en: Record<string, string> = {
  // 品牌
  "brand.name": "Agent Skills",
  "brand.title": "Agent Skills — Reusable Skills for AI Coding Assistants",
  "brand.description":
    "Discover reusable skills for AI coding assistants like Claude Code and Codex. Each skill is a bundle of instructions and code that teaches your AI assistant to perform professional tasks and automate complex workflows.",

  // 导航
  "nav.language": "简体中文",
  "nav.zh": "简体中文",
  "nav.en": "English",
  "nav.submit": "Submit",
  "nav.openMenu": "Open menu",
  "nav.official": "Official",
  "nav.featured": "Featured",
  "nav.categories": "Categories",
  "nav.submitSkill": "Submit a Skill",

  // 页脚
  "footer.browse": "Browse",
  "footer.official": "Official Skills",
  "footer.featured": "Featured Skills",
  "footer.categories": "All Categories",
  "footer.articles": "Articles & Tutorials",
  "footer.contribute": "Contribute",
  "footer.submitSkill": "Submit a Skill",

  // 首页 Hero
  "hero.titleMain": "Agent Skills",
  "hero.titleAccent": "",
  "hero.description":
    "Discover reusable skills for AI coding assistants like Claude Code and Codex. Each skill is a bundle of instructions and code that teaches your AI assistant to perform professional tasks and automate complex workflows.",
  "hero.stat.skills": "Skills",
  "hero.stat.authors": "Official Authors",
  "hero.stat.categories": "Categories",
  "hero.stat.official": "Official Skills",
  "hero.cta.browse": "Browse Skills",
  "hero.cta.submit": "Submit a Skill",
  "hero.cta.learnMore": "Learn more about Skills →",

  // 官方作者区块
  "official.title": "Official Skills",
  "official.skillCount": "{n} skills",

  // 赞助横幅
  "sponsor.badge": "Sponsored",

  // 技能卡片
  "card.download": "Download ZIP",

  // 技能详情
  "detail.notFound": "Skill not found",
  "detail.backHome": "Back to Home",
  "detail.authorLabel": "By ",
  "detail.copyInstall": "Copy install command",
  "detail.downloadZip": "Download ZIP",
  "detail.license": "License: {license}",
  "detail.advertise": "Advertise?",
  "detail.contactUs": "Contact us →",
  "detail.moreFrom": "More skills from {author}",
  "detail.overview": "Overview",

  // 面包屑
  "breadcrumb.home": "Agent Skills",

  // 全部分类
  "categories.title": "All Categories",
  "categories.featured": "Featured",
  "categories.skillCount": "{n} skills",

  // 分类详情
  "category.skillCount": "{n} skills in total",
  "category.empty": "No skills in this category yet.",

  // 精选技能
  "featured.title": "Featured Skills",
  "featured.desc": "Editor-picked high-quality skills — {n} in total",

  // 官方技能
  "officialPage.title": "Official Skills",
  "officialPage.desc": "High-quality skills maintained and certified by official teams — {n} in total",

  // 作者详情
  "author.notFound": "Author not found",
  "author.backHome": "Back to Home",
  "author.skillCount": "Published {n} skills",
  "author.empty": "This author hasn't published any skills yet.",
  "author.desc": "Browse all skills published by {name}",

  // 文章
  "articles.title": "Articles & Tutorials",
  "articles.desc": "Official tutorials and announcements for skills",
  "articles.empty": "No articles yet. Stay tuned!",
  "articles.back": "← Back to articles",
  "article.notFound": "Article not found or removed",
  "article.noContent": "No content yet",

  // 404
  "notFound.title": "Page Not Found",
  "notFound.desc": "The page you visited does not exist or has been removed.",
  "notFound.backHome": "Back to Home",

  // 提交技能
  "submit.title": "Submit a Skill",
  "submit.meta":
    "Share an Agent Skill you created to help more developers get more out of AI coding. Submissions enter a review queue and appear in the library once approved.",
  "submit.successTitle": "Submitted!",
  "submit.successDesc":
    "Your skill has been added to the review queue. It will appear in the library once an admin approves it.",
  "submit.backHome": "Back to Home",
  "submit.name": "Skill Name",
  "submit.namePlaceholder": "e.g. Frontend Design",
  "submit.author": "Author / Maintainer",
  "submit.authorPlaceholder": "Your GitHub username or team name",
  "submit.githubUrl": "GitHub Repository URL",
  "submit.githubUrlPlaceholder": "https://github.com/your-username/your-skill",
  "submit.description": "Short Description",
  "submit.descriptionPlaceholder": "Describe what this skill does in one sentence...",
  "submit.category": "Category Tags",
  "submit.categoryPlaceholder": "development, testing, creative...",
  "submit.tagsHint": "Separate multiple tags with commas",
  "submit.error": "Submission failed. Please try again later.",
  "submit.submitting": "Submitting…",

  // 错误边界
  "error.title": "Something went wrong",
  "error.desc": "An error occurred while loading this page. Please refresh and try again.",
  "error.reload": "Reload Page",
};

const dictionaries: Record<Lang, Record<string, string>> = { zh, en };

const I18nContext = createContext<I18nValue | null>(null);

function translate(lang: Lang, key: string, vars?: Record<string, string | number>): string {
  const dict = dictionaries[lang];
  let s = dict[key] ?? key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      s = s.split(`{${k}}`).join(String(v));
    }
  }
  return s;
}

function loadInitialLang(): Lang {
  try {
    const saved = localStorage.getItem(STORAGE_KEY);
    return saved === "en" ? "en" : "zh";
  } catch {
    return "zh";
  }
}

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(loadInitialLang);

  // 持久化 + 同步 <html lang>
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, lang);
    } catch {
      /* 忽略隐私模式等异常 */
    }
    document.documentElement.lang = lang;
  }, [lang]);

  const setLang = useCallback((l: Lang) => setLangState(l), []);
  const toggleLang = useCallback(() => setLangState((l) => (l === "zh" ? "en" : "zh")), []);
  const t = useCallback(
    (key: string, vars?: Record<string, string | number>) => translate(lang, key, vars),
    [lang]
  );

  const value = useMemo<I18nValue>(
    () => ({ lang, setLang, toggleLang, t }),
    [lang, setLang, toggleLang, t]
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nValue {
  const ctx = useContext(I18nContext);
  if (!ctx) {
    throw new Error("useI18n 必须在 <I18nProvider> 内使用");
  }
  return ctx;
}
