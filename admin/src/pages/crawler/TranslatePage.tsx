import { useEffect, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { crawlerApi } from "../../api/crawler";
import type { TranslationItem } from "../../types";

export default function TranslatePage() {
  const toast = useToast();
  const [items, setItems] = useState<TranslationItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [translating, setTranslating] = useState(false);
  const [page, setPage] = useState(1);
  const [preview, setPreview] = useState<TranslationItem | null>(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await crawlerApi.scanUntranslated(200);
      // 后端可能返回 null（无数据），兜底为空数组避免渲染报错
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
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
      width: "100px",
      render: (it) => (
        <span className={`badge ${it.titleTranslated ? "badge-success" : "badge-warning"}`}>
          {it.titleTranslated ? "已汉化" : "未汉化"}
        </span>
      ),
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
