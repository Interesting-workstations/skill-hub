import { Link } from "react-router-dom";
import { useSiteConfig } from "../../hooks/useSiteConfig";
import { useI18n } from "../../i18n";
import "./Footer.css";

export default function Footer() {
  const { siteName, slogan, icp } = useSiteConfig();
  const { t, lang } = useI18n();
  // 英文模式下品牌名/标语用语言包（后端配置为中文），中文模式保留后端可自定义值
  const displaySiteName = lang === "zh" ? siteName : t("brand.name");
  const displaySlogan = lang === "zh" ? slogan : t("footer.slogan");
  return (
    <footer className="footer">
      <div className="footer-inner">
        <div className="footer-brand">
          <span className="footer-logo-text">{displaySiteName}</span>
          <p className="footer-tagline">{displaySlogan}</p>
        </div>
        <div className="footer-links">
          <div className="footer-col">
            <h4>{t("footer.browse")}</h4>
            <Link to="/official">{t("footer.official")}</Link>
            <Link to="/featured">{t("footer.featured")}</Link>
            <Link to="/categories">{t("footer.categories")}</Link>
            <Link to="/articles">{t("footer.articles")}</Link>
          </div>
          <div className="footer-col">
            <h4>{t("footer.contribute")}</h4>
            <Link to="/submit">{t("footer.submitSkill")}</Link>
            <a href="https://github.com" target="_blank" rel="noopener noreferrer">
              GitHub
            </a>
          </div>
        </div>
      </div>
      <div className="footer-bottom">
        <p>© 2026 {displaySiteName}. Built with ❤️{icp ? <span style={{ marginLeft: 12 }}>{icp}</span> : null}</p>
      </div>
    </footer>
  );
}
