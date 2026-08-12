import "./Footer.css";

export default function Footer() {
  return (
    <footer className="footer">
      <div className="footer-inner">
        <div className="footer-brand">
          <span className="footer-logo-text">Agent Skills</span>
          <p className="footer-tagline">
            AI 编程助手的可复用技能资源库
          </p>
        </div>
        <div className="footer-links">
          <div className="footer-col">
            <h4>浏览</h4>
            <a href="/official">官方技能</a>
            <a href="/featured">精选技能</a>
            <a href="/categories">全部分类</a>
          </div>
          <div className="footer-col">
            <h4>贡献</h4>
            <a href="/submit">提交技能</a>
            <a href="https://github.com" target="_blank" rel="noopener noreferrer">
              GitHub
            </a>
          </div>
        </div>
      </div>
      <div className="footer-bottom">
        <p>© 2026 Agent Skills Hub. Built with ❤️</p>
      </div>
    </footer>
  );
}
