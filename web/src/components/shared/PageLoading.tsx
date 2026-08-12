import "./PageLoading.css";

/** 全局品牌加载指示器（用于 Suspense fallback / 数据加载态） */
export default function PageLoading() {
  return (
    <div className="page-loading" role="status" aria-label="页面加载中">
      <span className="page-loading-logo" aria-hidden="true">
        <span className="page-loading-logo-line" />
        <span className="page-loading-logo-line" />
        <span className="page-loading-logo-line" />
      </span>
      <span className="page-loading-brand">Agent Skills</span>
      <span className="page-loading-dots" aria-hidden="true">
        <span className="page-loading-dot" />
        <span className="page-loading-dot" />
        <span className="page-loading-dot" />
      </span>
    </div>
  );
}
