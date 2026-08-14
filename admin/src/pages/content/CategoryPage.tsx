import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import AppTable, { type Column } from "../../components/AppTable";
import { siteApi } from "../../api/site";
import type { Category } from "../../types";
import { formatNumber } from "../../utils/format";

export default function CategoryPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  useEffect(() => {
    siteApi.categories()
      .then(setCategories)
      .catch(() => setCategories([]))
      .finally(() => setLoading(false));
  }, []);

  const columns: Column<Category>[] = [
    {
      key: "name",
      title: "分类名称",
      render: (c) => <span style={{ fontWeight: 500 }}>{c.name}</span>,
    },
    {
      key: "slug",
      title: "Slug",
      render: (c) => <code style={{ fontSize: 12, color: "var(--color-text-secondary)" }}>{c.slug}</code>,
    },
    {
      key: "count",
      title: "技能数量",
      render: (c) => <span className="num" style={{ fontWeight: 500 }}>{formatNumber(c.count)}</span>,
    },
    {
      key: "skills",
      title: "最近技能",
      render: (c) => (
        <div style={{ display: "flex", gap: 4, flexWrap: "wrap", maxWidth: 360 }}>
          {c.skills.slice(0, 3).map((s) => (
            <span key={s.id} className="badge badge-neutral">{s.name}</span>
          ))}
          {c.skills.length > 3 && (
            <span className="badge badge-neutral">+{c.skills.length - 3}</span>
          )}
        </div>
      ),
    },
    {
      key: "action",
      title: "官网预览",
      render: (c) => (
        <Link to={`/category/${c.slug}`} className="btn-link" target="_blank">查看 →</Link>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>分类管理</h1>
          <div className="sub">官网分类来自 skill-hub 后端，数量实时统计</div>
        </div>
      </div>
      <AppTable
        columns={columns}
        data={categories}
        rowKey={(c) => c.slug}
        loading={loading}
        page={page}
        pageSize={10}
        onPageChange={setPage}
      />
    </div>
  );
}
