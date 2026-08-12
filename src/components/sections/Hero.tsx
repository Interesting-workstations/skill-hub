import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { sectionEnter, createAnimationContext } from "../../animations";
import { useStats } from "../../hooks/useSkillData";
import "./Hero.css";

export default function Hero() {
  const heroRef = useRef<HTMLElement>(null);
  const ctx = useRef(createAnimationContext());
  const { data: stats } = useStats();

  const allSkillsCount = stats?.totalSkills ?? 0;
  const authorsCount = stats?.totalAuthors ?? 0;
  const categoryCount = stats?.totalCategories ?? 0;
  const officialCount = stats?.officialSkills ?? 0;

  useEffect(() => {
    if (!heroRef.current) return;
    const ctxCurrent = ctx.current;
    ctxCurrent.killAll();

    const children = heroRef.current.querySelectorAll(".hero-content > *, .hero-stats, .hero-cta");
    if (children.length > 0) {
      const st = sectionEnter(children, { fromY: 16, staggerAmount: 0.08 });
      ctxCurrent.add(st);
    }

    return () => {
      ctxCurrent.killAll();
    };
  }, []);

  return (
    <section className="hero" ref={heroRef}>
      <div className="hero-content">
        <h1 className="hero-title">
          Agent Skills
          <span className="hero-title-accent"> 资源库</span>
        </h1>
        <p className="hero-desc">
          发现适用于 Claude Code、Codex 等 AI
          编程助手的可复用技能。每个技能都是一组指令和代码包，教会你的 AI
          助手执行专业任务并自动化复杂工作流。
        </p>
      </div>

      {/* 数据统计 */}
      <div className="hero-stats">
        <div className="hero-stat">
          <span className="hero-stat-num">{allSkillsCount}+</span>
          <span className="hero-stat-label">收录技能</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{authorsCount}+</span>
          <span className="hero-stat-label">官方作者</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{categoryCount}+</span>
          <span className="hero-stat-label">技能分类</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{officialCount}+</span>
          <span className="hero-stat-label">官方技能</span>
        </div>
      </div>

      {/* CTA 按钮区 */}
      <div className="hero-cta">
        <Link to="/categories" className="hero-btn-primary">
          浏览技能
        </Link>
        <Link to="/submit" className="hero-btn-secondary">
          提交技能
        </Link>
        <a
          href="https://www.anthropic.com/news/skills"
          className="hero-btn-text"
          target="_blank"
          rel="noopener noreferrer"
        >
          了解更多关于 Skills →
        </a>
      </div>
    </section>
  );
}
