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

/** 官方组织概览（官网「官方技能 / 官方组织」统一数据源，含各组织官方技能数） */
export interface OfficialOrgSummary {
  owner: string;
  displayName: string;
  avatar: string;
  officialCount: number;
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

/** 赞助商（admin 管理，官网按位置渲染；中英描述按语言展示） */
export interface Sponsor {
  id: string;
  name: string;
  /** emoji 或图片 URL */
  logo: string;
  descriptionZh: string;
  descriptionEn: string;
  url: string;
  /** home（首页横幅）/ sidebar（详情页侧边栏）/ both */
  position: string;
  enabled: boolean;
  sortOrder: number;
  clicks: number;
  createdAt: string;
}
