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
}

export interface SkillCategory {
  name: string;
  slug: string;
  count: number;
  skills: Skill[];
}
