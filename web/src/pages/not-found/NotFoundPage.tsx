import { Link } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import PageContainer from "../../components/shared/PageContainer";

export default function NotFoundPage() {
  usePageMeta({ title: "404 — 页面未找到" });
  const pageRef = usePageAnimation();

  return (
    <PageContainer
      ref={pageRef}
      maxWidth={640}
      padding="80px 24px 60px"
      style={{ textAlign: "center" }}
    >
      <div style={{ fontSize: 72, fontWeight: 800, color: "var(--color-primary)", marginBottom: 8 }}>
        404
      </div>
      <h1 style={{ fontSize: 24, fontWeight: 700, color: "var(--color-text)", margin: "0 0 8px" }}>
        页面未找到
      </h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        你访问的页面不存在或已被移除。
      </p>
      <Link
        to="/"
        style={{
          display: "inline-block",
          padding: "10px 24px",
          background: "var(--color-primary)",
          color: "var(--color-text-inverse)",
          borderRadius: 8,
          textDecoration: "none",
          fontWeight: 500,
          fontSize: 14,
        }}
      >
        返回首页
      </Link>
    </PageContainer>
  );
}
