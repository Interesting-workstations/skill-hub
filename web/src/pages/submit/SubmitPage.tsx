import { useState } from "react";
import { Link } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { site } from "../../config/site";
import { submitSkill } from "../../services/api/skills";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";

type SubmitState = "idle" | "submitting" | "success" | "error";

export default function SubmitPage() {
  const pageRef = usePageAnimation();
  const [state, setState] = useState<SubmitState>("idle");
  const [error, setError] = useState("");
  const [form, setForm] = useState({
    name: "",
    author: "",
    githubUrl: "",
    description: "",
    category: "",
  });

  usePageMeta({ title: `提交技能 — ${site.name}` });

  const set = (key: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setForm((f) => ({ ...f, [key]: e.target.value }));
    setState("idle");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (state === "submitting") return;
    setState("submitting");
    setError("");
    try {
      await submitSkill({
        name: form.name.trim(),
        author: form.author.trim(),
        description: form.description.trim(),
        category: form.category.trim() || undefined,
        githubUrl: form.githubUrl.trim() || undefined,
        tags: form.category
          .split(",")
          .map((t) => t.trim())
          .filter(Boolean),
      });
      setState("success");
    } catch (err) {
      setState("error");
      setError(err instanceof Error ? err.message : "提交失败，请稍后重试");
    }
  };

  const inputStyle: React.CSSProperties = {
    width: "100%",
    padding: "12px 16px",
    border: "1px solid var(--color-border-input)",
    borderRadius: 8,
    fontSize: 14,
    color: "var(--color-text)",
    outline: "none",
    boxSizing: "border-box",
  };

  return (
    <PageContainer ref={pageRef} maxWidth={640}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "提交技能" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>提交技能</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 32px", lineHeight: 1.6 }}>
        分享你创建的 Agent Skill，帮助更多开发者提升 AI 编程体验。提交后进入待审核，审核通过后展示在资源库。
      </p>

      {state === "success" ? (
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
            你的技能已提交到待审核队列，管理员审核通过后将出现在资源库中。
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
              maxLength={120}
              value={form.name}
              onChange={set("name")}
              placeholder="例如：Frontend Design"
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              作者 / 维护者 <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <input
              type="text"
              required
              maxLength={120}
              value={form.author}
              onChange={set("author")}
              placeholder="你的 GitHub 用户名或团队名"
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              GitHub 仓库地址
            </label>
            <input
              type="url"
              value={form.githubUrl}
              onChange={set("githubUrl")}
              placeholder="https://github.com/your-username/your-skill"
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              简短描述 <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <textarea
              required
              rows={3}
              value={form.description}
              onChange={set("description")}
              placeholder="用一句话描述这个技能的功能..."
              style={{ ...inputStyle, resize: "vertical", fontFamily: "inherit" }}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              分类标签
            </label>
            <input
              type="text"
              value={form.category}
              onChange={set("category")}
              placeholder="development, testing, creative..."
              style={inputStyle}
            />
            <p style={{ fontSize: 12, color: "var(--color-text-muted)", margin: "4px 0 0" }}>多个标签用逗号分隔</p>
          </div>

          {state === "error" && (
            <p style={{ fontSize: 13, color: "var(--color-danger)", margin: 0 }}>{error}</p>
          )}

          <button
            type="submit"
            disabled={state === "submitting"}
            style={{
              padding: "12px 32px",
              background: "var(--color-primary)",
              color: "var(--color-text-inverse)",
              border: "none",
              borderRadius: 8,
              fontSize: 15,
              fontWeight: 600,
              cursor: state === "submitting" ? "wait" : "pointer",
              alignSelf: "flex-start",
              transition: "background 0.15s",
              opacity: state === "submitting" ? 0.7 : 1,
            }}
          >
            {state === "submitting" ? "提交中…" : "提交技能"}
          </button>
        </form>
      )}
    </PageContainer>
  );
}
