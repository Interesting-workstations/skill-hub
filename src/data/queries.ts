import type { Skill } from "./types";
import { featuredSkills, skillCategories } from "./skills";

/** Get all skills as a flat list for detail page lookup */
export function getAllSkills(): Skill[] {
  return [...featuredSkills, ...skillCategories.flatMap((c) => c.skills)];
}

/** Get skills by author name */
export function getSkillsByAuthor(authorName: string): Skill[] {
  return getAllSkills().filter((s) => s.author === authorName);
}

/** Get skills by category slug */
export function getSkillsByCategory(slug: string): Skill[] {
  return getAllSkills().filter(
    (s) => s.category === slug || s.tags.includes(slug)
  );
}

/** Get official skills */
export function getOfficialSkills(): Skill[] {
  return getAllSkills().filter((s) => s.isOfficial);
}
