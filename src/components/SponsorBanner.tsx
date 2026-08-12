import "./SponsorBanner.css";

interface Sponsor {
  name: string;
  logo: string;
  description: string;
  url: string;
}

const sponsors: Sponsor[] = [
  {
    name: "ego lite browser",
    logo: "🌐",
    description:
      "ego lite 是AI代理运行网页自动化时最快的浏览器，可与Codex或Claude Code共享您的登录状态，零成本，零配置。",
    url: "https://lite.ego.app/",
  },
  {
    name: "Alpha Vantage MCP Server",
    logo: "📊",
    description:
      "获取金融市场数据：实时和历史股票、ETF、期权、外汇、加密货币、大宗商品、基本面、技术指标等",
    url: "https://mcp.alphavantage.co/",
  },
];

export default function SponsorBanner() {
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
              <span className="sponsor-badge">赞助</span>
            </div>
            <p className="sponsor-desc">{s.description}</p>
          </div>
        </a>
      ))}
    </div>
  );
}
