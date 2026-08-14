import { useParams } from "react-router-dom";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useCategory } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n } from "../../i18n";

export default function CategoryPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: category, loading } = useCategory(slug);
  const skills = category?.skills ?? [];
  const { t } = useI18n();

  // 兜底显示名（数据未就绪时用 slug 转换）
  const fallbackName = slug
    ? slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())
    : "";
  const displayName = category?.name ?? fallbackName;

  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [slug, skills.length]);

  usePageMeta({
    title: displayName ? `${displayName} — ${t("brand.name")}` : t("brand.title"),
  });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb
        items={[
          { label: t("breadcrumb.home"), to: "/" },
          { label: t("categories.title"), to: "/categories" },
          { label: displayName },
        ]}
      />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>{displayName}</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        {t("category.skillCount", { n: skills.length })}
      </p>

      {skills.length > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
          {skills.map((s) => (
            <SkillCard key={s.id} skill={s} />
          ))}
        </div>
      ) : (
        <p style={{ color: "var(--color-text-muted)", fontSize: 14 }}>{t("category.empty")}</p>
      )}
    </PageContainer>
  );
}
