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
