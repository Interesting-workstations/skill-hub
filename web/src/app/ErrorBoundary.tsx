import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
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
            出错了
          </h1>
          <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 24px" }}>
            页面加载出现异常，请刷新重试。
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
            刷新页面
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
