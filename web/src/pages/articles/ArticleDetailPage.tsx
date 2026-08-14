import { Link, useParams } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useAsyncData } from "../../hooks/useAsyncData";
import { fetchArticle } from "../../services/api/skills";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import MarkdownContent from "../../components/shared/MarkdownContent";
import { useI18n } from "../../i18n";

// 已读文章去重：localStorage 记录访问过的文章，同一浏览器每篇只累计一次浏览量。
const READ_KEY = "skillhub-read-articles";
function isArticleRead(id: string): boolean {
  try {
    const list: string[] = JSON.parse(localStorage.getItem(READ_KEY) || "[]");
    return list.includes(id);
  } catch {
    return false;
  }
}
function markArticleRead(id: string): void {
  try {
    const list: string[] = JSON.parse(localStorage.getItem(READ_KEY) || "[]");
    if (!list.includes(id)) {
      list.push(id);
      localStorage.setItem(READ_KEY, JSON.stringify(list));
    }
  } catch {
    /* 忽略存储异常 */
  }
}

export default function ArticleDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: article, loading } = useAsyncData(
    () => {
      if (!id) return Promise.reject(new Error("缺少文章 ID"));
      const read = isArticleRead(id);
      if (!read) markArticleRead(id);
      // 已读过的文章不再触发浏览量累加
      return fetchArticle(id, !read);
    },
    [id]
  );
  const pageRef = usePageAnimation();
  const { t } = useI18n();

  usePageMeta({
    title: article ? `${article.title} — ${t("brand.name")}` : `${t("articles.title")} — ${t("brand.name")}`,
    description: article ? article.content?.slice(0, 160).replace(/[#*`>\n]/g, " ") : undefined,
  });

  if (loading) {
    return <PageLoading />;
  }

  if (!article) {
    return (
      <PageContainer ref={pageRef} maxWidth={720}>
        <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: t("articles.title"), to: "/articles" }]} />
        <div style={{ textAlign: "center", padding: "60px 0", color: "var(--color-text-muted)" }}>
          <p style={{ fontSize: 40, marginBottom: 12 }}>📄</p>
          <p>{t("article.notFound")}</p>
          <p style={{ marginTop: 16 }}>
            <Link to="/articles" style={{ color: "var(--color-primary)", textDecoration: "none" }}>
              {t("articles.back")}
            </Link>
          </p>
        </div>
      </PageContainer>
    );
  }

  return (
    <PageContainer ref={pageRef} maxWidth={760}>
      <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: t("articles.title"), to: "/articles" }, { label: article.title }]} />

      <h1 style={{ fontSize: 30, fontWeight: 800, color: "var(--color-text)", margin: "0 0 10px", lineHeight: 1.3 }}>
        {article.title}
      </h1>
      <div style={{ display: "flex", gap: 12, alignItems: "center", fontSize: 13, color: "var(--color-text-muted)", marginBottom: 28 }}>
        <span style={{ color: "var(--color-primary)" }}>{article.category}</span>
        <span>·</span>
        <span>{article.author}</span>
        <span>·</span>
        <span>{article.updatedAt}</span>
        <span>·</span>
        <span>👁 {article.views}</span>
      </div>

      <div
        style={{
          padding: "28px 32px",
          background: "var(--color-surface)",
          border: "1px solid var(--color-border)",
          borderRadius: 12,
        }}
      >
        {article.content ? (
          <MarkdownContent content={article.content} />
        ) : (
          <p style={{ color: "var(--color-text-secondary)" }}>{t("article.noContent")}</p>
        )}
      </div>

      <div style={{ marginTop: 28 }}>
        <Link
          to="/articles"
          style={{ color: "var(--color-primary)", textDecoration: "none", fontSize: 14 }}
        >
          {t("articles.back")}
        </Link>
      </div>
    </PageContainer>
  );
}
