import SkillCard from "../skill/SkillCard";
import { Link } from "react-router-dom";
import type { Skill } from "../../data/types";
import "./SkillSection.css";

interface Props {
  title: string;
  count: number;
  slug: string;
  skills: Skill[];
}

export default function SkillSection({ title, count, slug, skills }: Props) {
  // 首页每个模块最多显示两行（桌面 4 个 / 移动 2 个），超出通过「查看更多」查看
  const showMore = skills.length > 2;

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
      {showMore && (
        <div className="skill-section-more">
          <Link to={`/category/${slug}`} className="skill-more-btn">
            查看全部 {count} 个技能
            <span aria-hidden="true">→</span>
          </Link>
        </div>
      )}
    </section>
  );
}
