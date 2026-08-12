import SkillCard from "./SkillCard";
import type { Skill } from "../data/skills";
import "./SkillSection.css";

interface Props {
  title: string;
  count: number;
  slug: string;
  skills: Skill[];
}

export default function SkillSection({ title, count, slug, skills }: Props) {
  return (
    <section className="skill-section">
      <div className="skill-section-header">
        <a href={`/category/${slug}`} className="skill-section-title">
          <h2>{title}</h2>
          <span className="skill-section-count">{count}</span>
          <span className="skill-section-arrow">→</span>
        </a>
      </div>
      <div className="skill-section-grid">
        {skills.map((skill) => (
          <SkillCard key={skill.id} skill={skill} />
        ))}
      </div>
    </section>
  );
}
