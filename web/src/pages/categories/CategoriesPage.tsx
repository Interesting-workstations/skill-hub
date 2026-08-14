import { Link } from "react-router-dom";
import { useCategories, useStats } from "../../hooks/useSkillData";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { site } from "../../config/site";

export default function CategoriesPage() {
  const { data: categories, loading } = useCategories();
  const { data: stats } = useStats();
  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".category-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [categories?.length]);

  usePageMeta({ title: `全部分类 — ${site.name}` });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "全部分类" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 28px" }}>全部分类</h1>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: 16 }}>
        {/* Featured category */}
        <Link
          to="/featured"
          className="category-card"
          style={{
            padding: "24px",
            border: "1px solid var(--color-border)",
            borderRadius: 12,
            textDecoration: "none",
            color: "inherit",
            transition: "border-color 0.15s, box-shadow 0.15s",
            background: "linear-gradient(135deg, var(--color-primary-50), var(--color-primary-100))",
          }}
        >
          <span style={{ fontSize: 28, display: "block", marginBottom: 8 }}>⭐</span>
          <h3 style={{ fontSize: 17, fontWeight: 600, color: "var(--color-text)", margin: "0 0 4px" }}>精选技能</h3>
          <p style={{ fontSize: 13, color: "var(--color-text-secondary)", margin: 0 }}>{stats?.featuredSkills ?? 0} 个技能</p>
        </Link>

        {(categories ?? []).map((cat) => (
          <Link
            key={cat.slug}
            to={`/category/${cat.slug}`}
            className="category-card"
            style={{
              padding: "24px",
              border: "1px solid var(--color-border)",
              borderRadius: 12,
              textDecoration: "none",
              color: "inherit",
              transition: "border-color 0.15s, box-shadow 0.15s",
              background: "var(--color-surface)",
            }}
          >
            <span style={{ fontSize: 28, display: "block", marginBottom: 8 }}>
              {cat.slug === "document" ? "📄" : cat.slug === "browser-automation" ? "🌐" :
               cat.slug === "database" ? "🗄️" : cat.slug === "development" ? "💻" :
               cat.slug === "creative" ? "🎨" : cat.slug === "media" ? "🎬" :
               cat.slug === "productivity" ? "⚡" : "📦"}
            </span>
            <h3 style={{ fontSize: 17, fontWeight: 600, color: "var(--color-text)", margin: "0 0 4px" }}>{cat.name}</h3>
            <p style={{ fontSize: 13, color: "var(--color-text-secondary)", margin: 0 }}>{cat.count} 个技能</p>
          </Link>
        ))}
      </div>
    </PageContainer>
  );
}
