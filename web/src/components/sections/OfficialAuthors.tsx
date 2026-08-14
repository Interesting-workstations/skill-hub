import { Link } from "react-router-dom";
import type { OfficialOrgSummary } from "../../data/types";
import { cardHoverEnter, cardHoverLeave } from "../../animations";
import { useI18n } from "../../i18n";
import OrgAvatar from "../shared/OrgAvatar";
import "./OfficialAuthors.css";

interface Props {
  orgs: OfficialOrgSummary[];
}

const handleMouseEnter = (e: React.MouseEvent<HTMLAnchorElement>) => {
  cardHoverEnter(e.currentTarget);
};
const handleMouseLeave = (e: React.MouseEvent<HTMLAnchorElement>) => {
  cardHoverLeave(e.currentTarget);
};

export default function OfficialAuthors({ orgs }: Props) {
  const { t } = useI18n();
  // 只展示发布了官方技能的组织，按官方技能数降序（与官方组织表统一）
  const official = orgs
    .filter((o) => o.officialCount > 0)
    .sort((a, b) => b.officialCount - a.officialCount);
  if (official.length === 0) {
    return null;
  }
  return (
    <section className="official-section">
      <div className="official-header">
        <Link to="/official" className="official-title-link">
          <h2 className="official-title">{t("official.title")}</h2>
          <span className="official-arrow">→</span>
        </Link>
      </div>
      <div className="official-grid">
        {official.map((org) => (
          <Link
            key={org.owner}
            to={`/official?org=${encodeURIComponent(org.owner)}`}
            className="author-card"
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
          >
            <OrgAvatar
              src={org.logoUrl}
              fallback={org.avatar}
              size={44}
              className="author-avatar"
            />
            <div className="author-info">
              <h3 className="author-name">{org.displayName}</h3>
              <p className="author-count">{t("official.skillCount", { n: org.officialCount })}</p>
            </div>
          </Link>
        ))}
      </div>
      <div className="official-more">
        <Link to="/orgs" className="official-more-link">
          {t("orgs.viewAll", { n: orgs.length })} <span className="official-arrow">→</span>
        </Link>
      </div>
    </section>
  );
}
