import { useEffect, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import { contentApi } from "../../api/content";
import { useToast } from "../../components/Toast";
import type { CrawledDataItem, DataStatus } from "../../types";

const STATUS_LABEL: Record<DataStatus, { text: string; cls: string }> = {
  pending: { text: "待审核", cls: "badge-warning" },
  approved: { text: "已审核", cls: "badge-running" },
  published: { text: "已发布", cls: "badge-success" },
  ignored: { text: "已忽略", cls: "badge-neutral" },
};

export default function CrawlerDataPage() {
  const toast = useToast();
  const [data, setData] = useState<CrawledDataItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<DataStatus | "">("");
  const [source, setSource] = useState<"all" | "official" | "community">("all");
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<string[]>([]);

  const load = async () => {
    setLoading(true);
    const list = await contentApi.listData(filter || undefined);
    setData(list);
    setSelected([]);
    setPage(1);
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, [filter]);

  const filtered = data.filter((d) => {
    if (keyword && !d.name.toLowerCase().includes(keyword.toLowerCase())) return false;
    if (source === "official" && !d.isOfficial) return false;
    if (source === "community" && d.isOfficial) return false;
    return true;
  });

  // 批量修改状态
  const batch = async (status: DataStatus) => {
    if (selected.length === 0) {
      toast.info("请先勾选要操作的数据");
      return;
    }
    try {
      await contentApi.batchUpdateDataStatus(selected, status);
      const label = status === "approved" ? "通过" : status === "published" ? "发布" : "忽略";
      toast.success(`已批量${label} ${selected.length} 条`);
      void load();
    } catch {
      // 错误已由全局 Toast 提示
    }
  };

  // 一键发布全部已审核数据（清理历史积压）
  const publishAll = async () => {
    if (!window.confirm("确认将全部「已审核」数据发布到官网？")) return;
    try {
      const res = await contentApi.publishAllApproved();
      toast.success(`已一键发布 ${res.published} 条已审核数据`);
      void load();
    } catch {
      // 错误已由全局 Toast 提示
    }
  };

  const columns: Column<CrawledDataItem>[] = [
    {
      key: "name",
      title: "标题",
      render: (d) => (
        <div>
          <div style={{ fontWeight: 500 }}>
            {d.name}
            {d.isOfficial ? (
              <span className="badge badge-neutral" style={{ marginLeft: 6 }}>官方</span>
            ) : (
              <span className="badge badge-neutral" style={{ marginLeft: 6, color: "var(--color-text-tertiary)" }}>社区/个人</span>
            )}
          </div>
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
      render: (d) => <span className="badge badge-neutral">{d.category}</span>,
    },
    {
      key: "stars",
      title: "星标",
      render: (d) => <span className="num">{d.githubStars ?? "--"}</span>,
    },
    {
      key: "status",
      title: "状态",
      render: (d) => {
        const s = STATUS_LABEL[d.status];
        return <span className={`badge ${s.cls}`}>{s.text}</span>;
      },
    },
    {
      key: "actions",
      title: "操作",
      width: "200px",
      render: (d) => (
        <div style={{ display: "flex", gap: 4 }}>
          {d.status === "pending" && (
            <>
              <button className="btn-link" onClick={() => contentApi.updateDataStatus(d.id, "approved").then(load)}>审核通过</button>
              <button className="btn-link danger" onClick={() => contentApi.updateDataStatus(d.id, "ignored").then(load)}>忽略</button>
            </>
          )}
          {d.status === "approved" && (
            <button className="btn-link" onClick={() => contentApi.updateDataStatus(d.id, "published").then(load)}>发布</button>
          )}
          <button className="btn-link danger" onClick={() => contentApi.deleteData(d.id).then(load)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>抓取数据</h1>
          <div className="sub">爬虫抓取到的全部数据，审核后发布到官网</div>
        </div>
      </div>

      <AppTable
        columns={columns}
        data={filtered}
        rowKey={(d) => d.id}
        loading={loading}
        page={page}
        pageSize={15}
        onPageChange={setPage}
        selectable
        selectedKeys={selected}
        onSelectionChange={setSelected}
        toolbar={
          <div className="filters" style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
            <input
              className="input"
              placeholder="搜索标题"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              style={{ width: 200 }}
            />
            <select className="select" value={source} onChange={(e) => setSource(e.target.value as "all" | "official" | "community")}>
              <option value="all">全部来源</option>
              <option value="official">官方</option>
              <option value="community">社区/个人</option>
            </select>
            <select className="select" value={filter} onChange={(e) => setFilter(e.target.value as DataStatus | "")}>
              <option value="">全部状态</option>
              <option value="pending">待审核</option>
              <option value="approved">已审核</option>
              <option value="published">已发布</option>
              <option value="ignored">已忽略</option>
            </select>
            <button className="btn btn-sm btn-primary" onClick={publishAll} title="一键发布全部已审核（approved）数据到官网">
              🚀 一键发布已审核
            </button>
            {selected.length > 0 && (
              <div style={{ display: "flex", gap: 6, alignItems: "center" }}>
                <span style={{ fontSize: 13, color: "var(--color-text-secondary)" }}>已选 {selected.length} 条</span>
                <button className="btn btn-sm btn-primary" onClick={() => batch("approved")}>批量通过</button>
                <button className="btn btn-sm btn-primary" onClick={() => batch("published")}>批量发布</button>
                <button className="btn btn-sm btn-danger" onClick={() => batch("ignored")}>批量忽略</button>
              </div>
            )}
          </div>
        }
      />
    </div>
  );
}
