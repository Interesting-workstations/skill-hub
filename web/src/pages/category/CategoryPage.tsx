import { useParams } from "react-router-dom";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useCategory } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { site } from "../../config/site";

export default function CategoryPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: category, loading } = useCategory(slug);
  const skills = category?.skills ?? [];

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
    title: displayName ? `${displayName} — ${site.name}` : site.title,
  });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb
        items={[
          { label: "Agent Skills 资源库", to: "/" },
          { label: "全部分类", to: "/categories" },
          { label: displayName },
        ]}
      />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>{displayName}</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        共 <strong style={{ color: "var(--color-text)" }}>{skills.length}</strong> 个技能
      </p>

      {skills.length > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
          {skills.map((s) => (
            <SkillCard key={s.id} skill={s} />
          ))}
        </div>
      ) : (
        <p style={{ color: "var(--color-text-muted)", fontSize: 14 }}>该分类暂无技能。</p>
      )}
    </PageContainer>
  );
}
