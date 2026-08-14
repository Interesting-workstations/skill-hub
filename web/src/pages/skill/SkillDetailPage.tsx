import { useParams, Link } from "react-router-dom";
import { useRef, useCallback } from "react";
import SkillCard from "../../components/skill/SkillCard";
import MarkdownContent from "../../components/shared/MarkdownContent";
import { sectionEnter, buttonClick, panelEnterRight } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useSkill, useSkills, useCategories } from "../../hooks/useSkillData";
import { skillDownloadUrl } from "../../services/api/skills";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n } from "../../i18n";
import "./SkillDetailPage.css";

export default function SkillDetailPage() {
  const { skillId } = useParams<{ skillId: string }>();
  const { data: skill, loading } = useSkill(skillId);
  const { t } = useI18n();
  // 同作者的其他技能（skill 未就绪时不请求）
  const { data: authorSkillsData } = useSkills(skill ? { author: skill.author } : null);
  const authorSkills = (authorSkillsData ?? []).filter((s) => s.id !== skill?.id);
  // 全部分类（用于判断 tag 是否为真实分类，避免跳到空分类页）
  const { data: categoryList } = useCategories();
  const categorySlugs = new Set((categoryList ?? []).map((c) => c.slug));
  const sidebarRef = useRef<HTMLElement>(null);
  const copyBtnRef = useRef<HTMLButtonElement>(null);

  const handleCopy = useCallback(() => {
    if (skill?.installCommand) {
      navigator.clipboard.writeText(skill.installCommand);
    }
    if (copyBtnRef.current) {
      buttonClick(copyBtnRef.current);
    }
  }, [skill]);

  const pageRef = usePageAnimation((container, ctx) => {
    // 内容区块 stagger
    const sections = container.querySelectorAll(".detail-section");
    if (sections.length > 0) {
      ctx.add(sectionEnter(sections, { fromY: 12 }));
    }
    // 侧边栏面板滑入
    if (sidebarRef.current) {
      ctx.add(panelEnterRight(sidebarRef.current));
    }
  }, [skillId, skill]);

  usePageMeta({
    title: skill ? `${skill.name} — ${t("brand.name")}` : t("brand.title"),
    description: skill?.description,
  });

  if (loading) {
    return <PageLoading />;
  }

  if (!skill) {
    return (
      <div className="detail-not-found">
        <h2>{t("detail.notFound")}</h2>
        <Link to="/">{t("detail.backHome")}</Link>
      </div>
    );
  }

  return (
    <div className="detail-page" ref={pageRef}>
      {/* Breadcrumb */}
      <nav className="detail-breadcrumb" aria-label="breadcrumb">
        <Link to="/">{t("breadcrumb.home")}</Link>
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
                  {t("detail.authorLabel")}
                  <Link to={`/author/${skill.author.toLowerCase()}`}>{skill.author}</Link>
                </span>
              </div>
            </div>
            <MarkdownContent content={skill.description} className="detail-desc" />

            {/* Install command */}
            {skill.installCommand && (
              <div className="detail-install">
                <code>{skill.installCommand}</code>
                <button
                  ref={copyBtnRef}
                  className="btn-copy"
                  onClick={handleCopy}
                  title={t("detail.copyInstall")}
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
              <a href={skillDownloadUrl(skill.id)} className="btn-download">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                  <path d="M8 2v8M5 8l3 3 3-3M3 12h10" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                </svg>
                {t("detail.downloadZip")}
              </a>
              {skill.githubUrl && (
                <a href={skill.githubUrl} className="btn-github" target="_blank" rel="noopener noreferrer">
                  <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
                  </svg>
                  GitHub{skill.githubStars ? ` ${skill.githubStars}` : ""}
                </a>
              )}
            </div>
          </div>

          {/* Content sections：body 各行合并为完整 markdown 文档整体渲染，
              保留列表 / 代码块 / 表格等跨行结构 */}
          {skill.content?.map((section, i) => (
            <div key={i} className="detail-section">
              <h2>{section.heading === "概述" ? t("detail.overview") : section.heading}</h2>
              {section.body.length > 0 && (
                <MarkdownContent content={section.body.join("\n")} />
              )}
            </div>
          ))}

          {/* Tags */}
          <div className="detail-tags">
            {skill.tags.map((tag) =>
              categorySlugs.has(tag) ? (
                <Link key={tag} to={`/category/${tag}`} className="detail-tag">
                  {tag}
                </Link>
              ) : (
                <span key={tag} className="detail-tag">
                  {tag}
                </span>
              )
            )}
          </div>

          {/* License */}
          {skill.license && <p className="detail-license">{t("detail.license", { license: skill.license })}</p>}
        </article>

        {/* Sidebar */}
        <aside className="detail-sidebar" ref={sidebarRef}>
          {/* Ad placeholder */}
          <div className="sidebar-ad">
            <p className="sidebar-ad-text">
              {t("detail.advertise")}<a href="#">{t("detail.contactUs")}</a>
            </p>
          </div>

          {/* More from author */}
          {authorSkills.length > 0 && (
            <div className="sidebar-more">
              <h3>{t("detail.moreFrom", { author: skill.author })}</h3>
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
