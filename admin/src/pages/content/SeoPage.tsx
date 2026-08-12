import { useEffect, useState, type FormEvent } from "react";
import { contentApi } from "../../api/content";
import type { SeoConfig } from "../../types";

export default function SeoPage() {
  const [seo, setSeo] = useState<SeoConfig | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    contentApi.getSeo().then(setSeo);
  }, []);

  if (!seo) {
    return (
      <div className="page">
        <div className="loading"><span className="spin" />加载中…</div>
      </div>
    );
  }

  const update = <K extends keyof SeoConfig>(key: K, value: SeoConfig[K]) => {
    setSeo({ ...seo, [key]: value });
    setSaved(false);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await contentApi.saveSeo(seo);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>SEO 配置</h1>
          <div className="sub">官网页面标题、描述与关键词，影响搜索引擎收录</div>
        </div>
      </div>

      <form className="card card-pad form" onSubmit={handleSubmit}>
        <div className="form-item">
          <label>页面标题（Title）</label>
          <input className="input" value={seo.title} onChange={(e) => update("title", e.target.value)} />
          <span className="hint">建议 ≤ 60 字符，当前 {seo.title.length} 字符</span>
        </div>
        <div className="form-item">
          <label>页面描述（Description）</label>
          <textarea className="textarea" value={seo.description} onChange={(e) => update("description", e.target.value)} />
        </div>
        <div className="form-item">
          <label>关键词（Keywords）</label>
          <input className="input" value={seo.keywords} onChange={(e) => update("keywords", e.target.value)} />
          <span className="hint">逗号分隔</span>
        </div>
        <div className="form-item">
          <label>分享图（OG Image）</label>
          <input className="input" value={seo.ogImage} onChange={(e) => update("ogImage", e.target.value)} />
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button className="btn btn-primary" type="submit">保存配置</button>
          {saved && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 已保存</span>}
        </div>
      </form>
    </div>
  );
}
