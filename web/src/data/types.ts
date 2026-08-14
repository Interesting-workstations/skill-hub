export interface SkillSection {
  heading: string;
  body: string[];
}

export interface Skill {
  id: string;
  name: string;
  author: string;
  description: string;
  tags: string[];
  category: string;
  downloadUrl: string;
  isOfficial?: boolean;
  isFeatured?: boolean;
  installCommand?: string;
  githubUrl?: string;
  githubStars?: string;
  license?: string;
  skillPath?: string;
  content?: SkillSection[];
}

export interface Author {
  name: string;
  avatar: string;
  skillCount: number;
  slug: string;
  officialSkills?: number;
}

export interface Category {
  name: string;
  slug: string;
  count: number;
  skills: Skill[];
}

/** 作者详情（含其发布的技能） */
export interface AuthorDetail {
  author: Author;
  skills: Skill[];
}

/** 站点聚合统计 */
export interface Stats {
  totalSkills: number;
  totalAuthors: number;
  totalCategories: number;
  officialSkills: number;
  featuredSkills: number;
}

/** 官网文章（admin 管理，官网展示） */
export interface Article {
  id: string;
  title: string;
  status: string;
  category: string;
  author: string;
  views: number;
  updatedAt: string;
  content?: string;
}

/** 站点配置（admin 管理，官网动态读取） */
export interface SiteConfig {
  siteName: string;
  slogan: string;
  portalUrl: string;
  icp: string;
  contactEmail: string;
}

/** SEO 配置（admin 管理，官网动态读取） */
export interface SeoConfig {
  title: string;
  description: string;
  keywords: string;
  ogImage: string;
}
