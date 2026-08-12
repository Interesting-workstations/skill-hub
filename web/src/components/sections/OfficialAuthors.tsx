import { Link } from "react-router-dom";
import type { Author } from "../../data/types";
import { cardHoverEnter, cardHoverLeave } from "../../animations";
import "./OfficialAuthors.css";

interface Props {
  authors: Author[];
}

const handleMouseEnter = (e: React.MouseEvent<HTMLAnchorElement>) => {
  cardHoverEnter(e.currentTarget);
};
const handleMouseLeave = (e: React.MouseEvent<HTMLAnchorElement>) => {
  cardHoverLeave(e.currentTarget);
};

export default function OfficialAuthors({ authors }: Props) {
  // 只展示发布了官方技能的作者，数量为该作者的官方技能数
  const official = authors
    .filter((a) => (a.officialSkills ?? 0) > 0)
    .sort((a, b) => (b.officialSkills ?? 0) - (a.officialSkills ?? 0));
  return (
    <section className="official-section">
      <div className="official-header">
        <Link to="/official" className="official-title-link">
          <h2 className="official-title">官方技能</h2>
          <span className="official-arrow">→</span>
        </Link>
      </div>
      <div className="official-grid">
        {official.map((author) => (
          <Link
            key={author.slug}
            to={`/author/${author.slug}`}
            className="author-card"
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
          >
            <span className="author-avatar">{author.avatar}</span>
            <div className="author-info">
              <h3 className="author-name">{author.name}</h3>
              <p className="author-count">{author.officialSkills} 个技能</p>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
