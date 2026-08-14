import { useEffect, useState, type FormEvent } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { contentApi } from "../../api/content";
import type { Article } from "../../types";

export default function ArticlePage() {
  const toast = useToast();
  const [articles, setArticles] = useState<Article[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Article | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Article | null>(null);
  const [page, setPage] = useState(1);

  const [title, setTitle] = useState("");
  const [category, setCategory] = useState("教程");
  const [status, setStatus] = useState<"draft" | "published">("draft");
  const [content, setContent] = useState("");

  const load = async () => {
    setLoading(true);
    setArticles(await contentApi.listArticles());
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const resetForm = () => {
    setTitle("");
    setCategory("教程");
    setStatus("draft");
    setContent("");
  };

  const openCreate = () => {
    resetForm();
    setCreateOpen(true);
  };

  const openEdit = (a: Article) => {
    setTitle(a.title);
    setCategory(a.category);
    setStatus(a.status);
    setContent(a.content ?? "");
    setEditing(a);
  };

  const doSave = async () => {
    const wasEditing = Boolean(editing);
    try {
      if (editing) {
        await contentApi.updateArticle(editing.id, { title, category, status, content });
        setEditing(null);
      } else {
        await contentApi.createArticle({ title, category, content });
        setCreateOpen(false);
      }
      resetForm();
      await load();
      toast.success(wasEditing ? "文章已更新" : "文章已创建");
    } catch {
      // 错误已由全局 Toast 自动提示
    }
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    void doSave();
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
          <button className="btn-link" onClick={() => openEdit(a)}>编辑</button>
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
          <div className="sub">官网教程与公告内容（正文支持 Markdown）</div>
        </div>
        <button className="btn btn-primary" onClick={openCreate}>+ 新建文章</button>
      </div>

      <AppTable
        columns={columns}
        data={articles}
        rowKey={(a) => a.id}
        loading={loading}
        page={page}
        pageSize={10}
        onPageChange={setPage}
      />

      <AppDialog
        open={createOpen || Boolean(editing)}
        title={editing ? "编辑文章" : "新建文章"}
        onClose={() => {
          setCreateOpen(false);
          setEditing(null);
        }}
        onConfirm={doSave}
        confirmText={editing ? "保存" : "创建"}
      >
        <form className="form" onSubmit={handleSubmit}>
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
          <div className="form-item">
            <label>状态</label>
            <select className="select" value={status} onChange={(e) => setStatus(e.target.value as "draft" | "published")}>
              <option value="draft">草稿</option>
              <option value="published">已发布</option>
            </select>
          </div>
          <div className="form-item">
            <label>正文（Markdown）</label>
            <textarea
              className="textarea"
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={"支持 Markdown 语法，例如：\n## 标题\n正文段落…\n\n```bash\n命令示例\n```"}
              rows={8}
              style={{ fontFamily: "ui-monospace, SF Mono, Consolas, monospace", fontSize: 13 }}
            />
          </div>
        </form>
      </AppDialog>

      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除文章"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (!confirmDelete) return;
          try {
            await contentApi.deleteArticle(confirmDelete.id);
            setConfirmDelete(null);
            await load();
            toast.success("文章已删除");
          } catch {
            // 错误已由全局 Toast 自动提示
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
