import { useState } from "react";
import { Link } from "react-router-dom";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { submitSkill } from "../../services/api/skills";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import { useI18n } from "../../i18n";

type SubmitState = "idle" | "submitting" | "success" | "error";

export default function SubmitPage() {
  const pageRef = usePageAnimation();
  const { t } = useI18n();
  const [state, setState] = useState<SubmitState>("idle");
  const [error, setError] = useState("");
  const [form, setForm] = useState({
    name: "",
    author: "",
    githubUrl: "",
    description: "",
    category: "",
  });

  usePageMeta({ title: `${t("submit.title")} — ${t("brand.name")}` });

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
      setError(err instanceof Error ? err.message : t("submit.error"));
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
      <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: t("submit.title") }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>{t("submit.title")}</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 32px", lineHeight: 1.6 }}>
        {t("submit.meta")}
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
            {t("submit.successTitle")}
          </h2>
          <p style={{ fontSize: 14, color: "var(--color-success-text)", margin: "0 0 20px" }}>
            {t("submit.successDesc")}
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
            {t("submit.backHome")}
          </Link>
        </div>
      ) : (
        <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: 20 }}>
          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              {t("submit.name")} <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <input
              type="text"
              required
              maxLength={120}
              value={form.name}
              onChange={set("name")}
              placeholder={t("submit.namePlaceholder")}
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              {t("submit.author")} <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <input
              type="text"
              required
              maxLength={120}
              value={form.author}
              onChange={set("author")}
              placeholder={t("submit.authorPlaceholder")}
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              {t("submit.githubUrl")}
            </label>
            <input
              type="url"
              value={form.githubUrl}
              onChange={set("githubUrl")}
              placeholder={t("submit.githubUrlPlaceholder")}
              style={inputStyle}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              {t("submit.description")} <span style={{ color: "var(--color-danger)" }}>*</span>
            </label>
            <textarea
              required
              rows={3}
              value={form.description}
              onChange={set("description")}
              placeholder={t("submit.descriptionPlaceholder")}
              style={{ ...inputStyle, resize: "vertical", fontFamily: "inherit" }}
            />
          </div>

          <div>
            <label style={{ display: "block", fontSize: 14, fontWeight: 600, color: "var(--color-text-label)", marginBottom: 6 }}>
              {t("submit.category")}
            </label>
            <input
              type="text"
              value={form.category}
              onChange={set("category")}
              placeholder={t("submit.categoryPlaceholder")}
              style={inputStyle}
            />
            <p style={{ fontSize: 12, color: "var(--color-text-muted)", margin: "4px 0 0" }}>{t("submit.tagsHint")}</p>
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
            {state === "submitting" ? t("submit.submitting") : t("submit.title")}
          </button>
        </form>
      )}
    </PageContainer>
  );
}
