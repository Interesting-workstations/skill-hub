import { useEffect, useState, type FormEvent } from "react";
import { contentApi } from "../../api/content";
import type { SiteConfig } from "../../types";

export default function SiteSettingsPage() {
  const [config, setConfig] = useState<SiteConfig | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    contentApi.getSiteConfig().then(setConfig);
  }, []);

  if (!config) {
    return (
      <div className="page">
        <div className="loading"><span className="spin" />加载中…</div>
      </div>
    );
  }

  const update = <K extends keyof SiteConfig>(key: K, value: SiteConfig[K]) => {
    setConfig({ ...config, [key]: value });
    setSaved(false);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await contentApi.saveSiteConfig(config);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>网站基础配置</h1>
          <div className="sub">官网站点名称、标语与底部信息</div>
        </div>
      </div>

      <form className="card card-pad form" onSubmit={handleSubmit} style={{ maxWidth: 520 }}>
        <div className="form-item">
          <label>站点名称</label>
          <input className="input" value={config.siteName} onChange={(e) => update("siteName", e.target.value)} />
        </div>
        <div className="form-item">
          <label>站点标语</label>
          <input className="input" value={config.slogan} onChange={(e) => update("slogan", e.target.value)} />
        </div>
        <div className="form-item">
          <label>官网地址</label>
          <input className="input" value={config.portalUrl} onChange={(e) => update("portalUrl", e.target.value)} />
        </div>
        <div className="form-grid">
          <div className="form-item">
            <label>ICP 备案号</label>
            <input className="input" value={config.icp} onChange={(e) => update("icp", e.target.value)} />
          </div>
          <div className="form-item">
            <label>联系邮箱</label>
            <input className="input" value={config.contactEmail} onChange={(e) => update("contactEmail", e.target.value)} />
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button className="btn btn-primary" type="submit">保存配置</button>
          {saved && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 已保存</span>}
        </div>
      </form>
    </div>
  );
}
