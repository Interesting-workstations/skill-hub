import { useEffect, useState, type FormEvent } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { contentApi } from "../../api/content";
import type { Article } from "../../types";

export default function ArticlePage() {
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState<Article | null>(null);

  const [title, setTitle] = useState("");
  const [category, setCategory] = useState("教程");

  const load = async () => {
    setLoading(true);
    setArticles(await contentApi.listArticles());
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const doCreate = async () => {
    await contentApi.createArticle({
      title,
      category,
    });
    setCreateOpen(false);
    setTitle("");
    await load();
  };

  const handleCreate = (e: FormEvent) => {
    e.preventDefault();
    void doCreate();
  };

  const columns: Column<Article>[] = [
    {
      key: "title",
      title: "标题",
      render: (a) => <span style={{ fontWeight: 500 }}>{a.title}</span>,
    },
    {
      key: "status",
      title: "状态",
      render: (a) => (
        <span className={`badge ${a.status === "published" ? "badge-success" : "badge-waiting"}`}>
          {a.status === "published" ? "已发布" : "草稿"}
        </span>
      ),
    },
    {
      key: "category",
      title: "分类",
      render: (a) => <span className="badge badge-neutral">{a.category}</span>,
    },
    {
      key: "views",
      title: "浏览量",
      render: (a) => <span className="num">{a.views}</span>,
    },
    {
      key: "updatedAt",
      title: "更新时间",
      render: (a) => <span className="num">{a.updatedAt}</span>,
    },
    {
      key: "actions",
      title: "操作",
      width: "120px",
      render: (a) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link">编辑</button>
          <button className="btn-link danger" onClick={() => setConfirmDelete(a)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>文章管理</h1>
          <div className="sub">官网教程与公告内容</div>
        </div>
        <button className="btn btn-primary" onClick={() => setCreateOpen(true)}>+ 新建文章</button>
      </div>

      <AppTable
        columns={columns}
        data={articles}
        rowKey={(a) => a.id}
        loading={loading}
        pageSize={10}
      />

      <AppDialog
        open={createOpen}
        title="新建文章"
        onClose={() => setCreateOpen(false)}
        onConfirm={doCreate}
        confirmText="创建"
      >
        <form className="form" onSubmit={handleCreate}>
          <div className="form-item">
            <label>文章标题</label>
            <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="请输入标题" required />
          </div>
          <div className="form-item">
            <label>分类</label>
            <select className="select" value={category} onChange={(e) => setCategory(e.target.value)}>
              <option>教程</option>
              <option>官方</option>
              <option>公告</option>
            </select>
          </div>
        </form>
      </AppDialog>

      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除文章"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (confirmDelete) {
            await contentApi.deleteArticle(confirmDelete.id);
            setConfirmDelete(null);
            await load();
          }
        }}
        confirmText="确认删除"
        danger
      >
        <p>确定要删除文章「{confirmDelete?.title}」吗？</p>
      </AppDialog>
    </div>
  );
}
