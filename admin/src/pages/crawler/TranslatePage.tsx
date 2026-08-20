import { useEffect, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { crawlerApi } from "../../api/crawler";
import type { TranslateConfig, TranslateTestResult, TranslationItem } from "../../types";

export default function TranslatePage() {
  const toast = useToast();
  const [items, setItems] = useState<TranslationItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [translating, setTranslating] = useState(false);
  const [page, setPage] = useState(1);
  const [preview, setPreview] = useState<TranslationItem | null>(null);

  // 翻译通道设置
  const [cfg, setCfg] = useState<TranslateConfig | null>(null);
  const [primary, setPrimary] = useState("auto");
  const [testing, setTesting] = useState(false);
  const [testResults, setTestResults] = useState<TranslateTestResult[] | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await crawlerApi.scanUntranslated(200);
      // 后端可能返回 null（无数据），兜底为空数组避免渲染报错
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
      const cfgRes = await crawlerApi.getTranslateConfig();
      setCfg(cfgRes);
      setPrimary(cfgRes.primary || "auto");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  // 翻译单条
  const handleTranslateOne = async (it: TranslationItem) => {
    try {
      const updated = await crawlerApi.translateSkill(it.id);
      setPreview(updated);
      await load();
      toast.success("翻译完成");
    } catch {
      // 错误已由全局 Toast 自动提示
    }
  };

  // 批量翻译
  const handleTranslateAll = async () => {
    setTranslating(true);
    try {
      const res = await crawlerApi.translateAll();
      toast.success(`批量翻译完成：共 ${res.translated} 条`);
      await load();
    } catch {
      // 错误已由全局 Toast 自动提示
    } finally {
      setTranslating(false);
    }
  };

  // 保存主翻译通道
  const handleSaveProvider = async () => {
    try {
      const updated = await crawlerApi.saveTranslateConfig(primary);
      setCfg(updated);
      toast.success("翻译通道已更新，立即生效");
    } catch {
      // 全局 Toast 已提示
    }
  };

  // 测试翻译通道
  const handleTest = async () => {
    setTesting(true);
    setTestResults(null);
    try {
      const res = await crawlerApi.testTranslateProvider("all");
      setTestResults(res.results ?? []);
    } catch {
      // 全局 Toast 已提示
    } finally {
      setTesting(false);
    }
  };

  const providerLabel = (p: string) => cfg?.providerName?.[p] ?? p;
  const providerDesc: Record<string, string> = {
    tencent: "腾讯云机器翻译（每月免费额度）",
    baidu: "百度通用翻译（免费额度）",
    google: "Google 免费接口（国内服务器可能不通）",
    deepl: "DeepL（需 Key）",
  };

  const columns: Column<TranslationItem>[] = [
    {
      key: "name",
      title: "标题",
      width: "220px",
      render: (it) => (
        <div>
          <div style={{ fontWeight: 500 }}>{it.name}</div>
          {it.nameZh && (
            <div style={{ color: "var(--color-text-secondary)", fontSize: 13 }}>
              {it.nameZh}
            </div>
          )}
        </div>
      ),
    },
    {
      key: "titleTranslated",
      title: "标题状态",
      width: "110px",
      render: (it) => {
        if (it.titleTranslated) return <span className="badge badge-success">已汉化</span>;
        // 旧版假翻译写回的原文（nameZh == name）或纯品牌名：标为品牌名
        if (it.nameZh && it.nameZh.toLowerCase() === it.name.toLowerCase()) {
          return <span className="badge badge-info">品牌名</span>;
        }
        return <span className="badge badge-warning">未汉化</span>;
      },
    },
    {
      key: "description",
      title: "描述",
      width: "280px",
      render: (it) => {
        const desc = it.descriptionZh || it.description;
        return (
          <div style={{ maxWidth: 420 }}>
            <div style={{ display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden", fontSize: 13, lineHeight: 1.5 }}>
              {desc}
            </div>
          </div>
        );
      },
    },
    {
      key: "descTranslated",
      title: "描述状态",
      width: "100px",
      render: (it) => (
        <span className={`badge ${it.descTranslated ? "badge-success" : "badge-warning"}`}>
          {it.descTranslated ? "已汉化" : "未汉化"}
        </span>
      ),
    },
    {
      key: "author",
      title: "作者",
      width: "120px",
      render: (it) => <span>{it.author}</span>,
    },
    {
      key: "category",
      title: "分类",
      width: "100px",
      render: (it) => <span className="num">{it.category}</span>,
    },
    {
      key: "actions",
      title: "操作",
      width: "140px",
      render: (it) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link" onClick={() => handleTranslateOne(it)} disabled={translating}>
            翻译
          </button>
          <button className="btn-link" onClick={() => setPreview(it)}>预览</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>翻译管理</h1>
          <div className="sub">
            检测标题与描述均未汉化的技能（任一不含中文即视为未汉化），可单条或批量翻译为中文。
          </div>
        </div>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <span style={{ fontSize: 13, color: "var(--color-text-secondary)" }}>
            待翻译 <b style={{ color: "var(--color-danger)" }}>{total}</b> 条
          </span>
          <button className="btn" onClick={() => void load()} disabled={loading}>
            重新扫描
          </button>
          <button className="btn btn-primary" onClick={handleTranslateAll} disabled={translating || total === 0}>
            {translating ? "翻译中…" : "⚡ 全部翻译"}
          </button>
        </div>
      </div>

      {/* 翻译通道设置：选择主通道 + 一键测试连通性（哪个能用用哪个） */}
      <div className="card card-pad" style={{ marginBottom: 16 }}>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12 }}>
          <div style={{ fontWeight: 600, fontSize: 14 }}>翻译通道</div>
          <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
            <span style={{ fontSize: 13, color: "var(--color-text-secondary)" }}>
              当前链路：
              {(cfg?.providers ?? []).map((p, i) => (
                <span key={p}>
                  {i > 0 && <span style={{ margin: "0 4px", color: "var(--color-text-tertiary)" }}>→</span>}
                  <span className="badge badge-info" style={{ marginLeft: 2 }}>{providerLabel(p)}</span>
                </span>
              ))}
              {cfg?.lastSuccess && (
                <span style={{ marginLeft: 8 }}>最近成功：<b>{providerLabel(cfg.lastSuccess)}</b></span>
              )}
            </span>
            <button className="btn" onClick={handleTest} disabled={testing}>
              {testing ? "测试中…" : "🔍 测试全部通道"}
            </button>
          </div>
        </div>

        {/* 测试结果 */}
        {testResults && testResults.length > 0 && (
          <div style={{ marginBottom: 12, display: "flex", flexDirection: "column", gap: 6 }}>
            {testResults.map((r) => (
              <div key={r.provider} style={{ fontSize: 13, display: "flex", alignItems: "baseline", gap: 8 }}>
                <span className={`badge ${r.ok ? "badge-success" : "badge-danger"}`}>
                  {r.ok ? "✓ 可用" : "✗ 失败"}
                </span>
                <b>{r.name}</b>
                {r.ok ? (
                  <span style={{ color: "var(--color-text-secondary)" }}>
                    {r.output} <span style={{ color: "var(--color-text-tertiary)" }}>（{r.elapsed}）</span>
                  </span>
                ) : (
                  <span style={{ color: "var(--color-danger)" }}>{r.error}</span>
                )}
              </div>
            ))}
          </div>
        )}

        {/* 主通道选择 */}
        <div style={{ display: "flex", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
          <span style={{ fontSize: 13, color: "var(--color-text-secondary)" }}>主通道：</span>
          {(["auto", "tencent", "baidu", "google", "deepl"] as const).map((p) => {
            const active = primary === p;
            const configured = cfg?.configured?.[p];
            return (
              <button
                key={p}
                className={active ? "btn btn-primary" : "btn"}
                style={{ fontSize: 13 }}
                onClick={() => setPrimary(p)}
                title={providerDesc[p]}
              >
                {p === "auto" ? "自动" : providerLabel(p)}
                {p !== "auto" && (
                  <span style={{ marginLeft: 4, opacity: 0.8 }}>
                    {configured ? "·已配置" : "·未配置"}
                  </span>
                )}
              </button>
            );
          })}
          <button className="btn-link" onClick={handleSaveProvider} disabled={!cfg}>
            保存并生效
          </button>
        </div>
        <div style={{ marginTop: 8, fontSize: 12, color: "var(--color-text-tertiary)", lineHeight: 1.6 }}>
          自动：按 腾讯云 → 百度 → Google → DeepL 依次尝试，哪个能走通用哪个；选具体通道时优先用它，失败自动降级后续通道。
          密钥通过服务端环境变量配置（TX_SECRETID / TX_SECRETKEY / BAIDU_APPID / BAIDU_KEY / DEEPL_KEY）。
        </div>
      </div>

      <AppTable
        columns={columns}
        data={items}
        rowKey={(it) => it.id}
        loading={loading}
        page={page}
        pageSize={20}
        onPageChange={setPage}
      />

      {/* 预览弹窗：翻译前后对照 */}
      <AppDialog
        open={Boolean(preview)}
        title="翻译预览"
        onClose={() => setPreview(null)}
        size="lg"
      >
        {preview && (
          <div className="form" style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            <div>
              <div className="form-item"><label>英文标题</label><div style={{ padding: "6px 0" }}>{preview.name}</div></div>
              <div className="form-item"><label>中文标题</label><div style={{ padding: "6px 0", fontWeight: 500 }}>{preview.nameZh || "—"}</div></div>
            </div>
            <div>
              <div className="form-item"><label>英文描述</label>
                <div style={{ whiteSpace: "pre-wrap", fontSize: 13, lineHeight: 1.6, maxHeight: 160, overflow: "auto", padding: "6px 0" }}>{preview.description}</div>
              </div>
              <div className="form-item"><label>中文描述</label>
                <div style={{ whiteSpace: "pre-wrap", fontSize: 13, lineHeight: 1.6, maxHeight: 200, overflow: "auto", padding: "6px 0", color: "var(--color-text-secondary)" }}>
                  {preview.descriptionZh || "—"}
                </div>
              </div>
            </div>
            <div style={{ display: "flex", gap: 8 }}>
              <span className={`badge ${preview.titleTranslated ? "badge-success" : "badge-warning"}`}>
                标题 {preview.titleTranslated ? "已汉化" : "未汉化"}
              </span>
              <span className={`badge ${preview.descTranslated ? "badge-success" : "badge-warning"}`}>
                描述 {preview.descTranslated ? "已汉化" : "未汉化"}
              </span>
            </div>
          </div>
        )}
      </AppDialog>
    </div>
  );
}
