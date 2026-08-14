import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { sectionEnter, createAnimationContext } from "../../animations";
import { useStats } from "../../hooks/useSkillData";
import { useI18n } from "../../i18n";
import "./Hero.css";

export default function Hero() {
  const heroRef = useRef<HTMLElement>(null);
  const ctx = useRef(createAnimationContext());
  const { data: stats } = useStats();
  const { t } = useI18n();

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
          {t("hero.titleMain")}
          {t("hero.titleAccent") && <span className="hero-title-accent">{t("hero.titleAccent")}</span>}
        </h1>
        <p className="hero-desc">
          {t("hero.description")}
        </p>
      </div>

      {/* 数据统计 */}
      <div className="hero-stats">
        <div className="hero-stat">
          <span className="hero-stat-num">{allSkillsCount}+</span>
          <span className="hero-stat-label">{t("hero.stat.skills")}</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{authorsCount}+</span>
          <span className="hero-stat-label">{t("hero.stat.authors")}</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{categoryCount}+</span>
          <span className="hero-stat-label">{t("hero.stat.categories")}</span>
        </div>
        <div className="hero-stat-divider" />
        <div className="hero-stat">
          <span className="hero-stat-num">{officialCount}+</span>
          <span className="hero-stat-label">{t("hero.stat.official")}</span>
        </div>
      </div>

      {/* CTA 按钮区 */}
      <div className="hero-cta">
        <Link to="/categories" className="hero-btn-primary">
          {t("hero.cta.browse")}
        </Link>
        <Link to="/submit" className="hero-btn-secondary">
          {t("hero.cta.submit")}
        </Link>
        <a
          href="https://www.anthropic.com/news/skills"
          className="hero-btn-text"
          target="_blank"
          rel="noopener noreferrer"
        >
          {t("hero.cta.learnMore")}
        </a>
      </div>
    </section>
  );
}
