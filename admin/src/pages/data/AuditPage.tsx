import { useEffect, useMemo, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import { contentApi, type DataFilter } from "../../api/content";
import { useToast } from "../../components/Toast";
import type { CrawledDataItem } from "../../types";

const STATUS_OPTIONS: { value: DataFilter["status"]; label: string }[] = [
  { value: "", label: "全部状态" },
  { value: "pending", label: "待审核" },
  { value: "approved", label: "已批准" },
  { value: "published", label: "已发布" },
  { value: "ignored", label: "已忽略" },
];

/** 解析 GitHub star 展示（支持 271.9k / 1.2m 后缀） */
function formatStars(s?: string): string {
  if (!s) return "--";
  const n = Number(s);
  if (!Number.isNaN(n)) return n.toLocaleString();
  const m = s.match(/^([\d.]+)\s*([km])?$/i);
  if (m) {
    const base = parseFloat(m[1]);
    const unit = (m[2] || "").toLowerCase();
    if (unit === "k") return Math.round(base * 1000).toLocaleString();
    if (unit === "m") return Math.round(base * 1000000).toLocaleString();
  }
  return s;
}

/** 数据审核：筛选 + 高星优先 + 全选批量通过/忽略 */
export default function AuditPage() {
  const toast = useToast();
  const [data, setData] = useState<CrawledDataItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<string[]>([]);
  const [filter, setFilter] = useState<DataFilter>({ status: "pending", sort: "stars" });
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [autoAuditing, setAutoAuditing] = useState(false);

  // 从当前结果提取分类（用于下拉筛选）
  const categories = useMemo(
    () => Array.from(new Set(data.map((d) => d.category).filter(Boolean))).sort(),
    [data]
  );

  const load = async (f: DataFilter) => {
    setLoading(true);
    setData(await contentApi.listData(f));
    setSelected([]);
    setPage(1);
    setLoading(false);
  };

  useEffect(() => {
    void load(filter);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // 关键词搜索防抖 300ms
  useEffect(() => {
    const timer = setTimeout(() => {
      if (search !== (filter.q ?? "")) {
        const next = { ...filter, q: search };
        setFilter(next);
        void load(next);
      }
    }, 300);
    return () => clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search]);

  const applyFilter = (patch: Partial<DataFilter>) => {
    const next = { ...filter, ...patch, q: search };
    setFilter(next);
    void load(next);
  };

  const batch = async (status: "approved" | "ignored" | "published") => {
    if (selected.length === 0) {
      toast.info("请先勾选要操作的数据");
      return;
    }
    try {
      await contentApi.batchUpdateDataStatus(selected, status);
      const label = status === "approved" ? "通过" : status === "published" ? "发布" : "忽略";
      toast.success(`已批量${label} ${selected.length} 条`);
      void load(filter);
    } catch {
      // 错误已由全局 Toast 提示
    }
  };

  // 机器人自动审核：内容完整规范的直接通过，有问题的留给人工
  const runAutoAudit = async () => {
    setAutoAuditing(true);
    try {
      const res = await contentApi.autoAuditData();
      if (res.manual > 0) {
        toast.success(`🤖 机器人审核完成：${res.approved} 条直接通过，${res.manual} 条转人工`);
      } else {
        toast.success(`🤖 机器人审核完成：全部 ${res.approved} 条通过`);
      }
      void load(filter);
    } catch {
      // 错误已由全局 Toast 提示
    } finally {
      setAutoAuditing(false);
    }
  };

  const columns: Column<CrawledDataItem>[] = [
    {
      key: "name",
      title: "数据",
      render: (d) => (
        <div>
          <div style={{ fontWeight: 500 }}>{d.name}</div>
          <div style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>{d.source}</div>
        </div>
      ),
    },
    {
      key: "author",
      title: "作者",
      render: (d) => <span>{d.author}</span>,
    },
    {
      key: "category",
      title: "分类",
      render: (d) => <span className="badge badge-neutral">{d.category || "--"}</span>,
    },
    {
      key: "stars",
      title: "⭐ 星标",
      render: (d) => <span className="num">{formatStars(d.githubStars)}</span>,
    },
    {
      key: "source",
      title: "来源",
      render: (d) =>
        d.isOfficial ? (
          <span className="badge badge-success">官方</span>
        ) : (
          <span className="badge badge-neutral">社区</span>
        ),
    },
    {
      key: "actions",
      title: "审核操作",
      width: "180px",
      render: (d) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button
            className="btn btn-sm btn-primary"
            onClick={() => contentApi.updateDataStatus(d.id, "approved").then(() => load(filter))}
          >
            通过
          </button>
          <button
            className="btn btn-sm btn-danger"
            onClick={() => contentApi.updateDataStatus(d.id, "ignored").then(() => load(filter))}
          >
            忽略
          </button>
        </div>
      ),
    },
  ];

  const selectStyle: React.CSSProperties = {
    height: 32,
    padding: "0 8px",
    borderRadius: 8,
    border: "1px solid var(--color-border)",
    background: "var(--color-surface)",
    color: "var(--color-text)",
    fontSize: 13,
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>数据审核</h1>
          <div className="sub">
            审核爬虫抓取的数据：官方来源自动发布，社区来源需人工审核。支持筛选、高星优先、全选批量操作。
          </div>
        </div>
        <button
          className="btn btn-primary"
          onClick={runAutoAudit}
          disabled={autoAuditing}
          title="内容完整规范且无重复的直接通过，有问题的交给人工审核"
        >
          {autoAuditing ? "🤖 审核中…" : "🤖 机器人审核"}
        </button>
      </div>

      <AppTable
        columns={columns}
        data={data}
        rowKey={(d) => d.id}
        loading={loading}
        page={page}
        pageSize={20}
        onPageChange={setPage}
        selectable
        selectedKeys={selected}
        onSelectionChange={setSelected}
        emptyText="没有匹配的数据"
        toolbar={
          <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <select
              style={selectStyle}
              value={filter.status ?? ""}
              onChange={(e) => applyFilter({ status: e.target.value as DataFilter["status"] })}
            >
              {STATUS_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
            <select
              style={selectStyle}
              value={filter.source ?? ""}
              onChange={(e) => applyFilter({ source: e.target.value as DataFilter["source"] })}
            >
              <option value="">全部来源</option>
              <option value="official">官方</option>
              <option value="community">社区</option>
            </select>
            <select
              style={selectStyle}
              value={filter.category ?? ""}
              onChange={(e) => applyFilter({ category: e.target.value })}
            >
              <option value="">全部分类</option>
              {categories.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
            <select
              style={selectStyle}
              value={filter.sort ?? "newest"}
              onChange={(e) => applyFilter({ sort: e.target.value as DataFilter["sort"] })}
            >
              <option value="newest">最新优先</option>
              <option value="stars">高星优先</option>
              <option value="name">名称排序</option>
            </select>
            <input
              className="input"
              style={{ width: 220, height: 32 }}
              placeholder="搜索名称 / 作者 / 描述"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
            {selected.length > 0 && (
              <div style={{ display: "flex", gap: 6, marginLeft: 4 }}>
                <span style={{ fontSize: 13, color: "var(--color-text-secondary)", alignSelf: "center" }}>
                  已选 {selected.length} 条
                </span>
                <button className="btn btn-sm btn-primary" onClick={() => batch("approved")}>
                  批量通过
                </button>
                <button className="btn btn-sm btn-primary" onClick={() => batch("published")}>
                  批量发布
                </button>
                <button className="btn btn-sm btn-danger" onClick={() => batch("ignored")}>
                  批量忽略
                </button>
              </div>
            )}
          </div>
        }
      />
    </div>
  );
}

