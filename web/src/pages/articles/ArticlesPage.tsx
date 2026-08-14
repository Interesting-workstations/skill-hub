import { Link } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useAsyncData } from "../../hooks/useAsyncData";
import { fetchArticles } from "../../services/api/skills";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { site } from "../../config/site";

export default function ArticlesPage() {
  const { data: articles, loading } = useAsyncData(fetchArticles, []);
  const pageRef = usePageAnimation();

  usePageMeta({ title: `文章与教程 — ${site.name}` });

  return (
    <PageContainer ref={pageRef} maxWidth={720}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "文章与教程" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>文章与教程</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px", lineHeight: 1.6 }}>
        官方发布的技能使用教程与公告
      </p>

      {loading ? (
        <PageLoading />
      ) : !articles || articles.length === 0 ? (
        <div
          style={{
            padding: "48px 24px",
            textAlign: "center",
            color: "var(--color-text-muted)",
            border: "1px dashed var(--color-border)",
            borderRadius: 12,
            fontSize: 14,
          }}
        >
          暂无文章，敬请期待
        </div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          {articles.map((a) => (
            <Link
              key={a.id}
              to={`/articles/${a.id}`}
              style={{
                display: "block",
                padding: "18px 20px",
                background: "var(--color-surface)",
                border: "1px solid var(--color-border)",
                borderRadius: 12,
                textDecoration: "none",
                transition: "border-color 0.15s, box-shadow 0.15s",
              }}
            >
              <div style={{ fontSize: 16, fontWeight: 700, color: "var(--color-text)" }}>{a.title}</div>
              <div style={{ marginTop: 6, fontSize: 12, color: "var(--color-text-muted)" }}>
                <span style={{ display: "inline-flex", gap: 8, alignItems: "center" }}>
                  <span style={{ color: "var(--color-primary)" }}>{a.category}</span>
                  <span>·</span>
                  <span>{a.updatedAt}</span>
                  <span>·</span>
                  <span>👁 {a.views}</span>
                </span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </PageContainer>
  );
}
