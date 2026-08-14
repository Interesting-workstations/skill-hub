import { useEffect, useState, type FormEvent } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { crawlerApi } from "../../api/crawler";
import type { OfficialOrg } from "../../types";

const AVATAR_SUGGESTIONS = ["🅰️", "🤖", "🇬", "☁️", "🐙", "🔵", "🔷", "🟠", "🟢", "🔴", "🐳", "🦊", "🌊", "🐋", "🔥", "⚡"];

export default function OfficialOrgsPage() {
  const toast = useToast();
  const [orgs, setOrgs] = useState<OfficialOrg[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<OfficialOrg | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<OfficialOrg | null>(null);

  const [owner, setOwner] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [avatar, setAvatar] = useState("🅰️");
  const [sortOrder, setSortOrder] = useState(0);
  const [enabled, setEnabled] = useState(true);

  const load = async () => {
    setLoading(true);
    setOrgs(await crawlerApi.listOfficialOrgs());
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const resetForm = () => {
    setOwner("");
    setDisplayName("");
    setAvatar("🅰️");
    setSortOrder(0);
    setEnabled(true);
  };

  const openCreate = () => {
    resetForm();
    setCreateOpen(true);
  };

  const openEdit = (o: OfficialOrg) => {
    setOwner(o.owner);
    setDisplayName(o.displayName);
    setAvatar(o.avatar);
    setSortOrder(o.sortOrder);
    setEnabled(o.enabled);
    setEditing(o);
  };

  const doSave = async () => {
    const wasEditing = Boolean(editing);
    try {
      if (editing) {
        await crawlerApi.updateOfficialOrg(editing.owner, { displayName, avatar, sortOrder, enabled });
        setEditing(null);
      } else {
        await crawlerApi.createOfficialOrg({ owner, displayName, avatar, sortOrder, enabled });
        setCreateOpen(false);
      }
      resetForm();
      await load();
      toast.success(wasEditing ? "官方组织已更新" : "官方组织已添加");
    } catch {
      // 错误已由全局 Toast 自动提示
    }
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    void doSave();
  };

  const columns: Column<OfficialOrg>[] = [
    {
      key: "avatar",
      title: "标识",
      width: "70px",
      render: (o) => <span style={{ fontSize: 24 }}>{o.avatar || "🏛️"}</span>,
    },
    {
      key: "owner",
      title: "GitHub 组织",
      render: (o) => (
        <span style={{ fontWeight: 500 }}>
          {o.owner}
          <span style={{ color: "var(--color-text-tertiary)", fontWeight: 400, marginLeft: 6 }}>@{o.owner}</span>
        </span>
      ),
    },
    {
      key: "displayName",
      title: "展示名",
      render: (o) => <span>{o.displayName}</span>,
    },
    {
      key: "enabled",
      title: "状态",
      render: (o) => (
        <span className={`badge ${o.enabled ? "badge-success" : "badge-waiting"}`}>
          {o.enabled ? "启用" : "停用"}
        </span>
      ),
    },
    {
      key: "sortOrder",
      title: "排序",
      render: (o) => <span className="num">{o.sortOrder}</span>,
    },
    {
      key: "actions",
      title: "操作",
      width: "120px",
      render: (o) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link" onClick={() => openEdit(o)}>编辑</button>
          <button className="btn-link danger" onClick={() => setConfirmDelete(o)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>官方组织</h1>
          <div className="sub">动态识别官方来源：爬取到这些 GitHub 组织的仓库会自动标记为「官方」。修改即时生效，无需改代码。</div>
        </div>
        <button className="btn btn-primary" onClick={openCreate}>+ 新增组织</button>
      </div>

      <AppTable
        columns={columns}
        data={orgs}
        rowKey={(o) => o.owner}
        loading={loading}
        pageSize={15}
      />

      <AppDialog
        open={createOpen || Boolean(editing)}
        title={editing ? "编辑官方组织" : "新增官方组织"}
        onClose={() => {
          setCreateOpen(false);
          setEditing(null);
        }}
        onConfirm={doSave}
        confirmText={editing ? "保存" : "添加"}
      >
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-item">
            <label>GitHub 组织名</label>
            <input
              className="input"
              value={owner}
              onChange={(e) => setOwner(e.target.value)}
              placeholder="例如：nvidia、langchain-ai"
              required
              disabled={Boolean(editing)}
            />
            <div className="hint">仓库 owner 完全匹配（不区分大小写）即视为官方</div>
          </div>
          <div className="form-item">
            <label>展示名</label>
            <input
              className="input"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder="例如：NVIDIA、LangChain"
              required
            />
          </div>
          <div className="form-item">
            <label>标识 emoji</label>
            <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
              {AVATAR_SUGGESTIONS.map((a) => (
                <button
                  key={a}
                  type="button"
                  className={`avatar-pick${avatar === a ? " active" : ""}`}
                  onClick={() => setAvatar(a)}
                  aria-label={`选择 ${a}`}
                >
                  {a}
                </button>
              ))}
            </div>
            <input
              className="input"
              value={avatar}
              onChange={(e) => setAvatar(e.target.value)}
              placeholder="或输入自定义 emoji"
              style={{ marginTop: 8, width: 160 }}
            />
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
              启用（参与官方来源识别）
            </label>
          </div>
        </form>
      </AppDialog>

      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除官方组织"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (!confirmDelete) return;
          try {
            await crawlerApi.deleteOfficialOrg(confirmDelete.owner);
            setConfirmDelete(null);
            await load();
            toast.success("官方组织已删除");
          } catch {
            // 错误已由全局 Toast 自动提示
          }
        }}
        confirmText="确认删除"
        danger
      >
        <p>确定要删除官方组织「{confirmDelete?.displayName}（{confirmDelete?.owner}）」吗？</p>
        <p style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>
          删除后该组织下的技能将不再标记为官方（已有数据不受影响）。
        </p>
      </AppDialog>
    </div>
  );
}
