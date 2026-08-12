import { useState } from "react";

const FORMATS = [
  { id: "json", name: "JSON", desc: "结构化数据，适合二次处理" },
  { id: "csv", name: "CSV", desc: "表格格式，适合 Excel 打开" },
  { id: "markdown", name: "Markdown", desc: "适合直接作为官网内容" },
];

export default function ExportPage() {
  const [format, setFormat] = useState("json");
  const [scope, setScope] = useState("published");
  const [exporting, setExporting] = useState(false);
  const [done, setDone] = useState(false);

  const handleExport = () => {
    setExporting(true);
    setDone(false);
    setTimeout(() => {
      setExporting(false);
      setDone(true);
    }, 800);
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
            {done && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 导出成功，文件已生成（演示）</span>}
          </div>
        </div>
      </div>
    </div>
  );
}
