import { useEffect, useState, type FormEvent } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { contentApi } from "../../api/content";
import type { Sponsor } from "../../types";

const POSITION_OPTIONS = [
  { value: "home", label: "首页横幅" },
  { value: "sidebar", label: "详情页侧边栏" },
  { value: "both", label: "首页 + 侧边栏" },
];

export default function SponsorsPage() {
  const toast = useToast();
  const [sponsors, setSponsors] = useState<Sponsor[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<Sponsor | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<Sponsor | null>(null);

  const [name, setName] = useState("");
  const [logo, setLogo] = useState("");
  const [descriptionZh, setDescriptionZh] = useState("");
  const [descriptionEn, setDescriptionEn] = useState("");
  const [url, setUrl] = useState("");
  const [position, setPosition] = useState("home");
  const [enabled, setEnabled] = useState(true);
  const [sortOrder, setSortOrder] = useState(0);

  const load = async () => {
    setLoading(true);
    setSponsors(await contentApi.listSponsors());
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const resetForm = () => {
    setName("");
    setLogo("");
    setDescriptionZh("");
    setDescriptionEn("");
    setUrl("");
    setPosition("home");
    setEnabled(true);
    setSortOrder(0);
  };

  const openCreate = () => {
    resetForm();
    setCreateOpen(true);
  };

  const openEdit = (s: Sponsor) => {
    setName(s.name);
    setLogo(s.logo);
    setDescriptionZh(s.descriptionZh);
    setDescriptionEn(s.descriptionEn);
    setUrl(s.url);
    setPosition(s.position);
    setEnabled(s.enabled);
    setSortOrder(s.sortOrder);
    setEditing(s);
  };

  const doSave = async () => {
    const wasEditing = Boolean(editing);
    try {
      const input = { name, logo, descriptionZh, descriptionEn, url, position, enabled, sortOrder };
      if (editing) {
        await contentApi.updateSponsor(editing.id, input);
        setEditing(null);
      } else {
        await contentApi.createSponsor(input);
        setCreateOpen(false);
      }
      resetForm();
      await load();
      toast.success(wasEditing ? "赞助已更新" : "赞助已创建");
    } catch {
      // 错误已由全局 Toast 自动提示
    }
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    void doSave();
  };

  const positionLabel = (p: string) =>
    POSITION_OPTIONS.find((o) => o.value === p)?.label ?? p;

  const columns: Column<Sponsor>[] = [
    {
      key: "logo",
      title: "标识",
      width: "70px",
      render: (s) => (
        <span style={{ fontSize: 26 }}>{s.logo || "🪧"}</span>
      ),
    },
    {
      key: "name",
      title: "名称",
      render: (s) => <span style={{ fontWeight: 500 }}>{s.name}</span>,
    },
    {
      key: "position",
      title: "位置",
      render: (s) => <span className="badge badge-neutral">{positionLabel(s.position)}</span>,
    },
    {
      key: "enabled",
      title: "状态",
      render: (s) => (
        <span className={`badge ${s.enabled ? "badge-success" : "badge-waiting"}`}>
          {s.enabled ? "启用" : "停用"}
        </span>
      ),
    },
    {
      key: "sortOrder",
      title: "排序",
      render: (s) => <span className="num">{s.sortOrder}</span>,
    },
    {
      key: "clicks",
      title: "点击",
      render: (s) => <span className="num">{s.clicks}</span>,
    },
    {
      key: "actions",
      title: "操作",
      width: "120px",
      render: (s) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link" onClick={() => openEdit(s)}>编辑</button>
          <button className="btn-link danger" onClick={() => setConfirmDelete(s)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>赞助管理</h1>
          <div className="sub">官网首页横幅与详情页侧边栏的赞助位内容，支持中英文描述</div>
        </div>
        <button className="btn btn-primary" onClick={openCreate}>+ 新增赞助</button>
      </div>

      <AppTable
        columns={columns}
        data={sponsors}
        rowKey={(s) => s.id}
        loading={loading}
        pageSize={10}
      />

      <AppDialog
        open={createOpen || Boolean(editing)}
        title={editing ? "编辑赞助" : "新增赞助"}
        onClose={() => {
          setCreateOpen(false);
          setEditing(null);
        }}
        onConfirm={doSave}
        confirmText={editing ? "保存" : "创建"}
      >
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-item">
            <label>名称</label>
            <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="例如：ego lite browser" required />
          </div>
          <div className="form-item">
            <label>标识（emoji 或图片 URL）</label>
            <input className="input" value={logo} onChange={(e) => setLogo(e.target.value)} placeholder="例如：🌐 或 https://…/logo.png" />
          </div>
          <div className="form-item">
            <label>跳转链接</label>
            <input className="input" value={url} onChange={(e) => setUrl(e.target.value)} placeholder="https://…" type="url" />
          </div>
          <div className="form-item">
            <label>展示位置</label>
            <select className="select" value={position} onChange={(e) => setPosition(e.target.value)}>
              {POSITION_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
          <div className="form-item">
            <label>中文描述</label>
            <textarea className="textarea" value={descriptionZh} onChange={(e) => setDescriptionZh(e.target.value)} rows={2} placeholder="中文介绍…" />
          </div>
          <div className="form-item">
            <label>英文描述</label>
            <textarea className="textarea" value={descriptionEn} onChange={(e) => setDescriptionEn(e.target.value)} rows={2} placeholder="English description…" />
          </div>
          <div className="form-item">
            <label>排序（越小越靠前）</label>
            <input
              className="input"
              type="number"
              value={sortOrder}
              onChange={(e) => setSortOrder(Number(e.target.value))}
              style={{ width: 120 }}
            />
          </div>
          <div className="form-item">
            <label className="checkbox">
              <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />
              立即启用
            </label>
          </div>
        </form>
      </AppDialog>

      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除赞助"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (!confirmDelete) return;
          try {
            await contentApi.deleteSponsor(confirmDelete.id);
            setConfirmDelete(null);
            await load();
            toast.success("赞助已删除");
          } catch {
            // 错误已由全局 Toast 自动提示
          }
        }}
        confirmText="确认删除"
        danger
      >
        <p>确定要删除赞助「{confirmDelete?.name}」吗？</p>
      </AppDialog>
    </div>
  );
}
