import { useParams, Link } from "react-router-dom";
import { useEffect, useRef, useCallback } from "react";
import { getAllSkills } from "../data/skills";
import SkillCard from "../components/SkillCard";
import {
  pageEnter,
  createAnimationContext,
  sectionEnter,
  buttonClick,
  panelEnterRight,
} from "../animations";
import "./SkillDetailPage.css";

export default function SkillDetailPage() {
  const { skillId } = useParams<{ skillId: string }>();
  const allSkills = getAllSkills();
  const skill = allSkills.find((s) => s.id === skillId);
  const pageRef = useRef<HTMLDivElement>(null);
  const sidebarRef = useRef<HTMLElement>(null);
  const copyBtnRef = useRef<HTMLButtonElement>(null);
  const ctx = useRef(createAnimationContext());

  const handleCopy = useCallback(() => {
    if (skill?.installCommand) {
      navigator.clipboard.writeText(skill.installCommand);
    }
    if (copyBtnRef.current) {
      buttonClick(copyBtnRef.current);
    }
  }, [skill]);

  useEffect(() => {
    if (!pageRef.current) return;
    ctx.current.killAll();

    // 页面进入
    const tl = pageEnter(pageRef.current);
    ctx.current.add(tl);

    // 内容区块 stagger
    const sections = pageRef.current.querySelectorAll(".detail-section");
    if (sections.length > 0) {
      const st = sectionEnter(sections, { fromY: 12 });
      ctx.current.add(st);
    }

    // 侧边栏面板滑入
    if (sidebarRef.current) {
      const panel = panelEnterRight(sidebarRef.current);
      ctx.current.add(panel);
    }

    return () => {
      ctx.current.killAll();
    };
  }, [skillId]);

  if (!skill) {
    return (
      <div className="detail-not-found">
        <h2>技能未找到</h2>
        <Link to="/">返回首页</Link>
      </div>
    );
  }

  const authorSkills = allSkills.filter(
    (s) => s.author === skill.author && s.id !== skill.id
  );

  return (
    <div className="detail-page" ref={pageRef}>
      {/* Breadcrumb */}
      <nav className="detail-breadcrumb" aria-label="breadcrumb">
        <Link to="/">Agent Skills 资源库</Link>
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M5 3l4 4-4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
        </svg>
        <span>{skill.name}</span>
      </nav>

      <div className="detail-layout">
        {/* Main content */}
        <article className="detail-main">
          {/* Header */}
          <div className="detail-header">
            <div className="detail-header-top">
              <span className="detail-icon">{skill.isOfficial ? "⭐" : "📦"}</span>
              <div className="detail-meta">
                <h1>{skill.name}</h1>
                <span className="detail-author">
                  作者：<Link to={`/author/${skill.author.toLowerCase()}`}>{skill.author}</Link>
                </span>
              </div>
            </div>
            <p className="detail-desc">{skill.description}</p>

            {/* Install command */}
            {skill.installCommand && (
              <div className="detail-install">
                <code>{skill.installCommand}</code>
                <button
                  ref={copyBtnRef}
                  className="btn-copy"
                  onClick={handleCopy}
                  title="复制安装命令"
                >
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                    <rect x="5" y="5" width="9" height="9" rx="1.5" stroke="currentColor" strokeWidth="1.5" />
                    <path d="M3 11V3h8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                  </svg>
                </button>
              </div>
            )}

            {/* Action buttons */}
            <div className="detail-actions">
              <a href={skill.downloadUrl} className="btn-download">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M8 2v8M5 8l3 3 3-3M3 12h10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                下载 ZIP
              </a>
              {skill.githubUrl && (
                <a href={skill.githubUrl} className="btn-github" target="_blank" rel="noopener noreferrer">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
                  </svg>
                  GitHub {skill.githubStars}
                </a>
              )}
            </div>
          </div>

          {/* Content sections */}
          {skill.content?.map((section, i) => (
            <div key={i} className="detail-section">
              <h2>{section.heading}</h2>
              {section.body.map((p, j) => (
                <p key={j}>{p}</p>
              ))}
            </div>
          ))}

          {/* Tags */}
          <div className="detail-tags">
            {skill.tags.map((tag) => (
              <Link key={tag} to={`/category/${tag}`} className="detail-tag">
                {tag}
              </Link>
            ))}
          </div>

          {/* License */}
          {skill.license && <p className="detail-license">许可：{skill.license}</p>}
        </article>

        {/* Sidebar */}
        <aside className="detail-sidebar" ref={sidebarRef}>
          {/* Ad placeholder */}
          <div className="sidebar-ad">
            <p className="sidebar-ad-text">
              投放广告？<a href="#">联系我们 →</a>
            </p>
          </div>

          {/* More from author */}
          {authorSkills.length > 0 && (
            <div className="sidebar-more">
              <h3>来自 {skill.author} 的更多技能</h3>
              <div className="sidebar-skills">
                {authorSkills.slice(0, 4).map((s) => (
                  <SkillCard key={s.id} skill={s} />
                ))}
              </div>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}
