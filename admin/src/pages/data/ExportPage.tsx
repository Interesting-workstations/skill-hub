import { useState } from "react";
import { contentApi } from "../../api/content";
import type { ExportResult } from "../../api/content";

const FORMATS = [
  { id: "json", name: "JSON", desc: "结构化数据，适合二次处理", mime: "application/json" },
  { id: "csv", name: "CSV", desc: "表格格式，适合 Excel 打开", mime: "text/csv" },
  { id: "markdown", name: "Markdown", desc: "适合直接作为官网内容", mime: "text/markdown" },
] as const;

function downloadFile(result: ExportResult) {
  const fmt = FORMATS.find((f) => f.id === result.format);
  const blob = new Blob([result.content], { type: fmt?.mime ?? "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = result.filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export default function ExportPage() {
  const [format, setFormat] = useState<string>("json");
  const [scope, setScope] = useState("published");
  const [exporting, setExporting] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");

  const handleExport = async () => {
    setExporting(true);
    setDone(false);
    setError("");
    try {
      const result = await contentApi.exportData(format as "json" | "csv" | "markdown", scope);
      downloadFile(result);
      setDone(true);
    } catch (e) {
      setError(e instanceof Error ? e.message : "导出失败，请稍后重试");
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>数据导出</h1>
          <div className="sub">将官网数据导出为 JSON / CSV / Markdown</div>
        </div>
      </div>

      <div className="card card-pad">
        <div className="card-title">导出设置</div>
        <div className="form">
          <div className="form-item">
            <label>导出格式</label>
            <div style={{ display: "flex", gap: 10 }}>
              {FORMATS.map((f) => (
                <label
                  key={f.id}
                  className="card"
                  style={{
                    flex: 1,
                    padding: 14,
                    cursor: "pointer",
                    borderColor: format === f.id ? "var(--color-primary)" : "var(--color-border-light)",
                    display: "flex",
                    flexDirection: "column",
                    gap: 4,
                  }}
                >
                  <input
                    type="radio"
                    name="format"
                    value={f.id}
                    checked={format === f.id}
                    onChange={() => setFormat(f.id)}
                    style={{ display: "none" }}
                  />
                  <strong>{f.name}</strong>
                  <span style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>{f.desc}</span>
                </label>
              ))}
            </div>
          </div>

          <div className="form-item">
            <label>导出范围</label>
            <select className="select" value={scope} onChange={(e) => setScope(e.target.value)}>
              <option value="published">已发布数据</option>
              <option value="approved">已审核数据</option>
              <option value="all">全部数据</option>
            </select>
          </div>

          <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
            <button className="btn btn-primary" onClick={handleExport} disabled={exporting}>
              {exporting ? "导出中…" : "开始导出"}
            </button>
            {done && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 导出成功，已开始下载</span>}
            {error && <span style={{ color: "var(--color-danger)", fontSize: 13 }}>{error}</span>}
          </div>
        </div>
      </div>
    </div>
  );
}
