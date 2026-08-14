import { useEffect, useState } from "react";
import { useI18n, type I18nValue } from "../../i18n";
import { fetchSponsors, reportSponsorClick } from "../../services/api/skills";
import type { Sponsor } from "../../data/types";
import "./SponsorBanner.css";

// 接口无数据 / 请求失败时的兜底赞助商（描述来自语言包，跟随当前语言）
function fallbackSponsors(t: I18nValue["t"]): Sponsor[] {
  return [
    {
      id: "fb-ego",
      name: "ego lite browser",
      logo: "🌐",
      descriptionZh: t("sponsor.ego"),
      descriptionEn: t("sponsor.ego"),
      url: "https://lite.ego.app/",
      position: "home",
      enabled: true,
      sortOrder: 0,
      clicks: 0,
      createdAt: "",
    },
    {
      id: "fb-alpha",
      name: "Alpha Vantage MCP Server",
      logo: "📊",
      descriptionZh: t("sponsor.alpha"),
      descriptionEn: t("sponsor.alpha"),
      url: "https://mcp.alphavantage.co/",
      position: "home",
      enabled: true,
      sortOrder: 1,
      clicks: 0,
      createdAt: "",
    },
  ];
}

export default function SponsorBanner() {
  const { t, lang } = useI18n();
  // null = 加载中/未请求；[] = 接口无数据（走回退）
  const [sponsors, setSponsors] = useState<Sponsor[] | null>(null);

  useEffect(() => {
    let alive = true;
    fetchSponsors("home")
      .then((list) => {
        if (alive) setSponsors(list);
      })
      .catch(() => {
        if (alive) setSponsors([]);
      });
    return () => {
      alive = false;
    };
  }, []);

  const list = sponsors && sponsors.length > 0 ? sponsors : fallbackSponsors(t);
  const description = (s: Sponsor) => {
    if (lang === "zh") return s.descriptionZh || s.descriptionEn;
    return s.descriptionEn || s.descriptionZh;
  };

  return (
    <div className="sponsor-banner">
      {list.map((s) => (
        <a
          key={s.id}
          href={s.url}
          className="sponsor-card"
          target="_blank"
          rel="noopener noreferrer"
          onClick={() => reportSponsorClick(s.id)}
        >
          <span className="sponsor-logo">{s.logo || "🪧"}</span>
          <div className="sponsor-info">
            <div className="sponsor-header">
              <h3 className="sponsor-name">{s.name}</h3>
              <span className="sponsor-badge">{t("sponsor.badge")}</span>
            </div>
            <p className="sponsor-desc">{description(s)}</p>
          </div>
        </a>
      ))}
    </div>
  );
}
