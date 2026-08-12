import type { Author } from "../data/skills";
import "./OfficialAuthors.css";

interface Props {
  authors: Author[];
}

export default function OfficialAuthors({ authors }: Props) {
  return (
    <section className="official-section">
      <div className="official-header">
        <a href="/official" className="official-title-link">
          <h2 className="official-title">官方技能</h2>
          <span className="official-arrow">→</span>
        </a>
      </div>
      <div className="official-grid">
        {authors.map((author) => (
          <a
            key={author.slug}
            href={`/author/${author.slug}`}
            className="author-card"
          >
            <span className="author-avatar">{author.avatar}</span>
            <div className="author-info">
              <h3 className="author-name">{author.name}</h3>
              <p className="author-count">{author.skillCount} 个技能</p>
            </div>
          </a>
        ))}
      </div>
    </section>
  );
}
