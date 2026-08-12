import { Link } from "react-router-dom";
import { useButtonMicro } from "../animations";
import "./SkillCard.css";
import type { Skill } from "../data/skills";

export default function SkillCard({ skill }: { skill: Skill }) {
  const btnRef = useButtonMicro();

  return (
    <Link to={`/skill/${skill.id}`} className="skill-card" data-animate="card">
      <div className="skill-card-top">
        <div className="skill-card-icon">
          {skill.isOfficial ? "⭐" : "📦"}
        </div>
        <div className="skill-card-meta">
          <span className="skill-card-name">{skill.name}</span>
          <span className="skill-card-author">{skill.author}</span>
        </div>
        <button
          ref={btnRef}
          className="skill-card-download"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
          }}
          title="下载 ZIP"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path
              d="M8 2v8M5 8l3 3 3-3M3 12h10"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </button>
      </div>
      <p className="skill-card-desc">{skill.description}</p>
      <div className="skill-card-tags">
        {skill.tags.map((tag) => (
          <span key={tag} className="skill-tag">
            {tag}
          </span>
        ))}
      </div>
    </Link>
  );
}
