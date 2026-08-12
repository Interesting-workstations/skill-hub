import "./Hero.css";

export default function Hero() {
  return (
    <section className="hero">
      <div className="hero-content" data-animate="hero-content">
        <h1 className="hero-title">Agent Skills 资源库</h1>
        <p className="hero-desc">
          发现适用于 Claude Code、Codex 等 AI
          编程助手的可复用技能。每个技能都是一组指令和代码包，教会你的 AI
          助手执行专业任务并自动化复杂工作流。
        </p>
        <a
          href="https://www.anthropic.com/news/skills"
          className="hero-learn-more"
          target="_blank"
          rel="noopener noreferrer"
        >
          了解更多关于 Skills →
        </a>
      </div>

      <div className="hero-video">
        <div className="video-wrapper">
          <iframe
            src="https://www.youtube.com/embed/IoqpBKrNaZI"
            title="Agent Skills 介绍视频"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
            className="video-iframe"
          />
        </div>
      </div>
    </section>
  );
}
