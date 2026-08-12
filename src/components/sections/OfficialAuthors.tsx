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
  return (
    <section className="official-section">
      <div className="official-header">
        <Link to="/official" className="official-title-link">
          <h2 className="official-title">官方技能</h2>
          <span className="official-arrow">→</span>
        </Link>
      </div>
      <div className="official-grid">
        {authors.map((author) => (
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
              <p className="author-count">{author.skillCount} 个技能</p>
            </div>
          </Link>
        ))}
      </div>
    </section>
  );
}
