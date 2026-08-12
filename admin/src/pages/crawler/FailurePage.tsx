import { useEffect, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import { crawlerApi } from "../../api/crawler";
import type { FailureRecord } from "../../types";

export default function FailurePage() {
  const [failures, setFailures] = useState<FailureRecord[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    const data = await crawlerApi.listFailures();
    setFailures(data);
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const columns: Column<FailureRecord>[] = [
    {
      key: "taskName",
      title: "任务",
      render: (f) => <span style={{ fontWeight: 500 }}>{f.taskName}</span>,
    },
    {
      key: "url",
      title: "失败 URL",
      render: (f) => (
        <div>
          <div style={{ maxWidth: 260, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{f.url}</div>
          <div style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>{f.reason}</div>
        </div>
      ),
    },
    {
      key: "error",
      title: "错误信息",
      render: (f) => (
        <div style={{ maxWidth: 320, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", fontFamily: "var(--font-mono)", fontSize: 12 }} title={f.error}>
          {f.error}
        </div>
      ),
    },
    {
      key: "retryCount",
      title: "重试次数",
      render: (f) => (
        <span className={`badge ${f.retryCount >= 3 ? "badge-danger" : "badge-warning"}`}>
          {f.retryCount} 次
        </span>
      ),
    },
    {
      key: "failedAt",
      title: "失败时间",
      render: (f) => <span className="num">{f.failedAt}</span>,
    },
    {
      key: "actions",
      title: "操作",
      width: "160px",
      render: (f) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link" onClick={() => crawlerApi.retryFailure(f.id).then(load)}>重新执行</button>
          <button className="btn-link danger" onClick={() => crawlerApi.ignoreFailure(f.id).then(load)}>忽略</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>失败任务</h1>
          <div className="sub">
            共 {failures.length} 个失败记录 · 连续失败的爬虫会在工作台明显提示
          </div>
        </div>
      </div>
      <AppTable
        columns={columns}
        data={failures}
        rowKey={(f) => f.id}
        loading={loading}
        pageSize={10}
      />
    </div>
  );
}
