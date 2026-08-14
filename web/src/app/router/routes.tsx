// 本文件是路由配置模块（非组件文件），lazy 常量 + 路由表均为导出项
/* oxlint-disable react/only-export-components */
import { lazy, type ReactElement } from "react";

// 页面组件按需加载（lazy），实现代码分割
const HomePage = lazy(() => import("../../pages/home/HomePage"));
const SkillDetailPage = lazy(() => import("../../pages/skill/SkillDetailPage"));
const AuthorPage = lazy(() => import("../../pages/author/AuthorPage"));
const CategoryPage = lazy(() => import("../../pages/category/CategoryPage"));
const CategoriesPage = lazy(() => import("../../pages/categories/CategoriesPage"));
const OfficialPage = lazy(() => import("../../pages/official/OfficialPage"));
const FeaturedPage = lazy(() => import("../../pages/featured/FeaturedPage"));
const ArticlesPage = lazy(() => import("../../pages/articles/ArticlesPage"));
const ArticleDetailPage = lazy(() => import("../../pages/articles/ArticleDetailPage"));
const SubmitPage = lazy(() => import("../../pages/submit/SubmitPage"));
const NotFoundPage = lazy(() => import("../../pages/not-found/NotFoundPage"));

export interface RouteConfig {
  path: string;
  element: ReactElement;
}

/** 集中式路由表：新增页面只需在此注册 */
export const routes: RouteConfig[] = [
  { path: "/", element: <HomePage /> },
  { path: "/skill/:skillId", element: <SkillDetailPage /> },
  { path: "/author/:authorSlug", element: <AuthorPage /> },
  { path: "/category/:slug", element: <CategoryPage /> },
  { path: "/categories", element: <CategoriesPage /> },
  { path: "/official", element: <OfficialPage /> },
  { path: "/featured", element: <FeaturedPage /> },
  { path: "/articles", element: <ArticlesPage /> },
  { path: "/articles/:id", element: <ArticleDetailPage /> },
  { path: "/submit", element: <SubmitPage /> },
  { path: "*", element: <NotFoundPage /> },
];
