import { useEffect, useState } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import { contentApi } from "../../api/content";
import type { CrawledDataItem } from "../../types";

export default function AuditPage() {
  const [data, setData] = useState<CrawledDataItem[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    setLoading(true);
    setData(await contentApi.listData("pending"));
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

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
      render: (d) => <span className="badge badge-neutral">{d.category}</span>,
    },
    {
      key: "stars",
      title: "星标",
      render: (d) => <span className="num">{d.githubStars ?? "--"}</span>,
    },
    {
      key: "actions",
      title: "审核操作",
      width: "180px",
      render: (d) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn btn-sm btn-primary" onClick={() => contentApi.updateDataStatus(d.id, "approved").then(load)}>
            通过
          </button>
          <button className="btn btn-sm btn-danger" onClick={() => contentApi.updateDataStatus(d.id, "ignored").then(load)}>
            忽略
          </button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>数据审核</h1>
          <div className="sub">审核爬虫抓取的数据：通过后进入已审核，可发布到官网</div>
        </div>
      </div>
      <AppTable
        columns={columns}
        data={data}
        rowKey={(d) => d.id}
        loading={loading}
        pageSize={10}
        emptyText="没有待审核的数据"
      />
    </div>
  );
}
