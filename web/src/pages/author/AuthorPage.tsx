import { useParams, Link } from "react-router-dom";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useAuthor } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n } from "../../i18n";

export default function AuthorPage() {
  const { authorSlug } = useParams<{ authorSlug: string }>();
  const { data: detail, loading } = useAuthor(authorSlug);
  const author = detail?.author ?? null;
  const skills = detail?.skills ?? [];
  const { t } = useI18n();

  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [authorSlug, skills.length]);

  usePageMeta({
    title: author ? `${author.name} — ${t("brand.name")}` : t("brand.title"),
    description: author ? t("author.desc", { name: author.name }) : undefined,
  });

  if (loading) {
    return <PageLoading />;
  }

  if (!author) {
    return (
      <div className="detail-not-found">
        <h2>{t("author.notFound")}</h2>
        <Link to="/">{t("author.backHome")}</Link>
      </div>
    );
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: author.name }]} />

      <div style={{ display: "flex", alignItems: "center", gap: 20, marginBottom: 36 }}>
        <span style={{ fontSize: 48, background: "var(--color-primary-50)", borderRadius: 16, width: 80, height: 80, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
          {author.avatar}
        </span>
        <div>
          <h1 style={{ fontSize: 32, fontWeight: 800, margin: "0 0 4px", color: "var(--color-text)" }}>{author.name}</h1>
          <p style={{ fontSize: 16, color: "var(--color-text-secondary)", margin: 0 }}>
            {t("author.skillCount", { n: skills.length })}
          </p>
        </div>
      </div>

      {skills.length > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
          {skills.map((s) => (
            <SkillCard key={s.id} skill={s} />
          ))}
        </div>
      ) : (
        <p style={{ color: "var(--color-text-muted)", fontSize: 14 }}>{t("author.empty")}</p>
      )}
    </PageContainer>
  );
}
