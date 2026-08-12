import { Link } from "react-router-dom";
import { useRef } from "react";
import { cardHoverEnter, cardHoverLeave } from "../../animations";
import "./SkillCard.css";
import type { Skill } from "../../data/types";

export default function SkillCard({ skill }: { skill: Skill }) {
  const cardRef = useRef<HTMLAnchorElement>(null);

  const handleMouseEnter = () => {
    if (cardRef.current) cardHoverEnter(cardRef.current);
  };
  const handleMouseLeave = () => {
    if (cardRef.current) cardHoverLeave(cardRef.current);
  };

  return (
    <Link
      ref={cardRef}
      to={`/skill/${skill.id}`}
      className="skill-card"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <div className="skill-card-top">
        <div className="skill-card-icon">
          {skill.isOfficial ? "⭐" : "📦"}
        </div>
        <div className="skill-card-meta">
          <span className="skill-card-name">{skill.name}</span>
          <span className="skill-card-author">{skill.author}</span>
          {skill.githubStars && (
            <span className="skill-card-stars">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M12 17.27 18.18 21l-1.64-7.03L22 9.24l-7.19-.61L12 2 9.19 8.63 2 9.24l5.46 4.73L5.82 21z" />
              </svg>
              {skill.githubStars}
            </span>
          )}
        </div>
        <button
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
