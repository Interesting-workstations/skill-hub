import { Link } from "react-router-dom";
import { useSiteConfig } from "../../hooks/useSiteConfig";
import "./Footer.css";

export default function Footer() {
  const { siteName, slogan, icp } = useSiteConfig();
  return (
    <footer className="footer">
      <div className="footer-inner">
        <div className="footer-brand">
          <span className="footer-logo-text">{siteName}</span>
          <p className="footer-tagline">{slogan}</p>
        </div>
        <div className="footer-links">
          <div className="footer-col">
            <h4>浏览</h4>
            <Link to="/official">官方技能</Link>
            <Link to="/featured">精选技能</Link>
            <Link to="/categories">全部分类</Link>
            <Link to="/articles">文章与教程</Link>
          </div>
          <div className="footer-col">
            <h4>贡献</h4>
            <Link to="/submit">提交技能</Link>
            <a href="https://github.com" target="_blank" rel="noopener noreferrer">
              GitHub
            </a>
          </div>
        </div>
      </div>
      <div className="footer-bottom">
        <p>© 2026 {siteName}. Built with ❤️{icp ? <span style={{ marginLeft: 12 }}>{icp}</span> : null}</p>
      </div>
    </footer>
  );
}
