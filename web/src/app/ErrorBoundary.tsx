import { Component, type ReactNode } from "react";
import { useI18n } from "../i18n";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
}

/** 全局错误边界的兜底 UI（函数组件，可访问 i18n） */
function ErrorFallback() {
  const { t } = useI18n();
  return (
    <div
      style={{
        maxWidth: 640,
        margin: "0 auto",
        padding: "80px 24px",
        textAlign: "center",
      }}
    >
      <div style={{ fontSize: 48, marginBottom: 12 }}>⚠️</div>
      <h1 style={{ fontSize: 24, fontWeight: 700, color: "var(--color-text)", margin: "0 0 8px" }}>
        {t("error.title")}
      </h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 24px" }}>
        {t("error.desc")}
      </p>
      <button
        onClick={() => window.location.reload()}
        style={{
          padding: "10px 24px",
          background: "var(--color-primary)",
          color: "var(--color-text-inverse)",
          border: "none",
          borderRadius: 8,
          fontSize: 14,
          fontWeight: 500,
          cursor: "pointer",
        }}
      >
        {t("error.reload")}
      </button>
    </div>
  );
}

/** 全局错误边界：捕获渲染错误，避免整页白屏 */
export default class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: unknown) {
    console.error("ErrorBoundary caught:", error);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback />;
    }
    return this.props.children;
  }
}
