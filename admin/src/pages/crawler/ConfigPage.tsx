import { useEffect, useState, type FormEvent } from "react";
import { crawlerApi } from "../../api/crawler";
import type { CrawlerConfig } from "../../types";

export default function ConfigPage() {
  const [config, setConfig] = useState<CrawlerConfig | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    crawlerApi.getConfig().then(setConfig);
  }, []);

  if (!config) {
    return (
      <div className="page">
        <div className="loading"><span className="spin" />加载中…</div>
      </div>
    );
  }

  const update = <K extends keyof CrawlerConfig>(key: K, value: CrawlerConfig[K]) => {
    setConfig({ ...config, [key]: value });
    setSaved(false);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await crawlerApi.saveConfig(config);
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>爬虫配置</h1>
          <div className="sub">全局爬虫运行参数，保存后对后续任务生效（GitHub Token 池请在「Token 池」页管理）</div>
        </div>
      </div>

      <form className="card card-pad form" onSubmit={handleSubmit}>
        <div className="card-title">运行参数</div>
        <div className="form-grid">
          <div className="form-item">
            <label>并发数</label>
            <input className="input" type="number" min={1} max={16}
              value={config.concurrency}
              onChange={(e) => update("concurrency", Number(e.target.value))} />
          </div>
          <div className="form-item">
            <label>请求超时（秒）</label>
            <input className="input" type="number" min={5}
              value={config.timeout}
              onChange={(e) => update("timeout", Number(e.target.value))} />
          </div>
          <div className="form-item">
            <label>失败重试次数</label>
            <input className="input" type="number" min={0} max={10}
              value={config.retryCount}
              onChange={(e) => update("retryCount", Number(e.target.value))} />
          </div>
          <div className="form-item">
            <label>请求间隔（ms）</label>
            <input className="input" type="number" min={100}
              value={config.requestInterval}
              onChange={(e) => update("requestInterval", Number(e.target.value))} />
          </div>
          <div className="form-item">
            <label>单次最大抓取页数</label>
            <input className="input" type="number" min={50}
              value={config.maxPagesPerRun}
              onChange={(e) => update("maxPagesPerRun", Number(e.target.value))} />
          </div>
          <div className="form-item">
            <label>默认搜索关键词</label>
            <input className="input" value={config.defaultQuery}
              onChange={(e) => update("defaultQuery", e.target.value)} />
          </div>
        </div>

        <div className="card-title" style={{ marginTop: 8 }}>官方仓库</div>
        <div className="form-item">
          <label>官方仓库列表（逗号分隔）</label>
          <input className="input" value={config.officialRepos}
            onChange={(e) => update("officialRepos", e.target.value)} />
          <span className="hint">来自这些仓库的技能自动标记为官方：anthropics/skills,openai/codex,vercel/ai…</span>
        </div>

        <div className="form-item">
          <label>User-Agent</label>
          <input className="input" value={config.userAgent}
            onChange={(e) => update("userAgent", e.target.value)} />
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button className="btn btn-primary" type="submit">保存配置</button>
          {saved && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 已保存</span>}
        </div>
      </form>
    </div>
  );
}
