import { useI18n } from "../../i18n";
import "./SponsorBanner.css";

interface Sponsor {
  name: string;
  logo: string;
  description: string;
  url: string;
}

export default function SponsorBanner() {
  const { t } = useI18n();
  const sponsors: Sponsor[] = [
    {
      name: "ego lite browser",
      logo: "🌐",
      description: t("sponsor.ego"),
      url: "https://lite.ego.app/",
    },
    {
      name: "Alpha Vantage MCP Server",
      logo: "📊",
      description: t("sponsor.alpha"),
      url: "https://mcp.alphavantage.co/",
    },
  ];
  return (
    <div className="sponsor-banner">
      {sponsors.map((s) => (
        <a
          key={s.name}
          href={s.url}
          className="sponsor-card"
          target="_blank"
          rel="noopener noreferrer"
        >
          <span className="sponsor-logo">{s.logo}</span>
          <div className="sponsor-info">
            <div className="sponsor-header">
              <h3 className="sponsor-name">{s.name}</h3>
              <span className="sponsor-badge">{t("sponsor.badge")}</span>
            </div>
            <p className="sponsor-desc">{s.description}</p>
          </div>
        </a>
      ))}
    </div>
  );
}
