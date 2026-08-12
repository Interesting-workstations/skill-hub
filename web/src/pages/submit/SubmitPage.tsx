import { useState } from "react";
import { Link } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { site } from "../../config/site";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";

export default function SubmitPage() {
  const pageRef = usePageAnimation();
  const [submitted, setSubmitted] = useState(false);

  usePageMeta({ title: `提交技能 — ${site.name}` });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitted(true);
  };

  return (
    <PageContainer ref={pageRef} maxWidth={640}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "提交技能" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>提交技能</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 32px", lineHeight: 1.6 }}>
        分享你创建的 Agent Skill，帮助更多开发者提升 AI 编程体验。
      </p>

      {submitted ? (
        <div
          style={{
            padding: "40px 32px",
            background: "var(--color-success-bg)",
            border: "1px solid var(--color-success-border)",
            borderRadius: 12,
            textAlign: "center",
          }}
        >
          <span style={{ fontSize: 40, display: "block", marginBottom: 12 }}>🎉</span>
          <h2 style={{ fontSize: 20, fontWeight: 700, color: "var(--color-success-text-strong)", margin: "0 0 8px" }}>
            提交成功！
          </h2>
          <p style={{ fontSize: 14, color: "var(--color-success-text)", margin: "0 0 20px" }}>
            我们会尽快审核你的技能。审核通过后将出现在资源库中。
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
        </div>
      ) : (
        <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              技能名称 <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <input
              type="text"
              required
              placeholder="例如：Frontend Design"
              style={{
                width: "100%",
                padding: "12px 16px",
                border: "1px solid var(--color-border-input)",
                borderRadius: 8,
                fontSize: 14,
                color: "var(--color-text)",
                outline: "none",
                boxSizing: "border-box",
              }}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              GitHub 仓库地址 <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <input
              type="url"
              required
              placeholder="https://github.com/your-username/your-skill"
              style={{
                width: "100%",
                padding: "12px 16px",
                border: "1px solid var(--color-border-input)",
                borderRadius: 8,
                fontSize: 14,
                color: "var(--color-text)",
                outline: "none",
                boxSizing: "border-box",
              }}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              简短描述 <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <textarea
              required
              rows={3}
              placeholder="用一句话描述这个技能的功能..."
              style={{
                width: "100%",
                padding: "12px 16px",
                border: "1px solid var(--color-border-input)",
                borderRadius: 8,
                fontSize: 14,
                color: "var(--color-text)",
                outline: "none",
                resize: "vertical",
                boxSizing: "border-box",
                fontFamily: "inherit",
              }}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              分类标签
            </label>
            <input
              type="text"
              placeholder="development, testing, creative..."
              style={{
                width: "100%",
                padding: "12px 16px",
                border: "1px solid var(--color-border-input)",
                borderRadius: 8,
                fontSize: 14,
                color: "var(--color-text)",
                outline: "none",
                boxSizing: "border-box",
              }}
            />
            <p style={{ fontSize: 12, color: "var(--color-text-muted)", margin: "4px 0 0" }}>多个标签用逗号分隔</p>
          </div>

          <button
            type="submit"
            style={{
              padding: "12px 32px",
              background: "var(--color-primary)",
              color: "var(--color-text-inverse)",
              border: "none",
              borderRadius: 8,
              fontSize: 15,
              fontWeight: 600,
              cursor: "pointer",
              alignSelf: "flex-start",
              transition: "background 0.15s",
            }}
          >
            提交技能
          </button>
        </form>
      )}
    </PageContainer>
  );
}
