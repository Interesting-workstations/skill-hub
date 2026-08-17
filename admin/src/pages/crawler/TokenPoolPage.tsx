import { useEffect, useState, type FormEvent } from "react";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import { useToast } from "../../components/Toast";
import { crawlerApi } from "../../api/crawler";
import type { GitHubToken, TokenHealth } from "../../types";

export default function TokenPoolPage() {
  const toast = useToast();
  const [tokens, setTokens] = useState<GitHubToken[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [editing, setEditing] = useState<GitHubToken | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<GitHubToken | null>(null);

  const [token, setToken] = useState("");
  const [remark, setRemark] = useState("");
  const [enabled, setEnabled] = useState(true);
  const [page, setPage] = useState(1);

  // 一键检测结果
  const [checking, setChecking] = useState(false);
  const [health, setHealth] = useState<Record<string, TokenHealth>>({});

  const load = async () => {
    setLoading(true);
    try {
      const list = await crawlerApi.listTokens();
      setTokens(list);
      setHealth({});
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void load();
  }, []);

  const resetForm = () => {
    setToken("");
    setRemark("");
    setEnabled(true);
  };

  const openCreate = () => {
    resetForm();
    setCreateOpen(true);
  };

  const openEdit = (t: GitHubToken) => {
    setToken(""); // 编辑时 token 留空表示不修改（列表只回显脱敏值）
    setRemark(t.remark);
    setEnabled(t.enabled);
    setEditing(t);
  };

  const doSave = async () => {
    const wasEditing = Boolean(editing);
    try {
      if (editing) {
        await crawlerApi.updateToken(editing.id, { token, remark, enabled });
        setEditing(null);
      } else {
        await crawlerApi.createToken({ token, remark });
        setCreateOpen(false);
      }
      resetForm();
      await load();
      toast.success(wasEditing ? "Token 已更新" : "Token 已添加");
    } catch {
      // 错误已由全局 Toast 自动提示
    }
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    void doSave();
  };

  // 一键检测：逐个验证池中 token 是否可用
  const handleCheck = async () => {
    setChecking(true);
    setHealth({});
    try {
      const results = await crawlerApi.checkTokens();
      const m: Record<string, TokenHealth> = {};
      for (const r of results) m[r.masked] = r;
      setHealth(m);
      const bad = results.filter((r) => !r.ok);
      if (bad.length > 0) {
        toast.info(`检测完成：${results.length} 个 token 中 ${bad.length} 个不可用`);
      } else {
        toast.success(`检测完成：全部 ${results.length} 个 token 均可用`);
      }
    } catch {
      // 错误已由全局 Toast 自动提示
    } finally {
      setChecking(false);
    }
  };

  const columns: Column<GitHubToken>[] = [
    {
      key: "enabled",
      title: "状态",
      width: "110px",
      render: (t) => (
        <span style={{ display: "flex", flexDirection: "column", gap: 2 }}>
          <span className={`badge ${t.enabled ? "badge-success" : "badge-waiting"}`}>
            {t.enabled ? "启用" : "停用"}
          </span>
          {t.enabled && t.broken && (
            <span className="badge badge-danger" title={`预计 ${t.cooldownAt ?? ""} 恢复`}>
              熔断中
            </span>
          )}
        </span>
      ),
    },
    {
      key: "token",
      title: "Token",
      render: (t) => (
        <code style={{ fontFamily: "monospace", fontSize: 12 }}>{t.token}</code>
      ),
    },
    {
      key: "remark",
      title: "备注",
      render: (t) => <span>{t.remark || <span style={{ color: "var(--color-text-tertiary)" }}>—</span>}</span>,
    },
    {
      key: "createdAt",
      title: "添加时间",
      width: "110px",
      render: (t) => <span className="num">{t.createdAt}</span>,
    },
    {
      key: "health",
      title: "可用性",
      width: "130px",
      render: (t) => {
        const h = health[t.token];
        if (!h) return <span style={{ color: "var(--color-text-tertiary)" }}>未检测</span>;
        return h.ok ? (
          <span className="badge badge-success">✅ {h.detail}</span>
        ) : (
          <span className="badge badge-danger" title={h.detail}>❌ {h.detail}</span>
        );
      },
    },
    {
      key: "actions",
      title: "操作",
      width: "120px",
      render: (t) => (
        <div style={{ display: "flex", gap: 4 }}>
          <button className="btn-link" onClick={() => openEdit(t)}>编辑</button>
          <button className="btn-link danger" onClick={() => setConfirmDelete(t)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>GitHub Token 池</h1>
          <div className="sub">
            配置多个 GitHub Token（每行一个），爬虫请求时轮询使用；某个 token 被限流/拒绝（401/403/429）会自动切换下一个，全部失效才降级匿名。修改即时生效。
          </div>
        </div>
        <div style={{ display: "flex", gap: 8 }}>
          <button className="btn" onClick={handleCheck} disabled={checking || tokens.length === 0}>
            {checking ? "检测中…" : "🔍 一键检测"}
          </button>
          <button className="btn btn-primary" onClick={openCreate}>+ 新增 Token</button>
        </div>
      </div>

      <AppTable
        columns={columns}
        data={tokens}
        rowKey={(t) => t.id}
        loading={loading}
        page={page}
        pageSize={15}
        onPageChange={setPage}
      />

      <AppDialog
        open={createOpen || Boolean(editing)}
        title={editing ? "编辑 Token" : "新增 Token"}
        onClose={() => {
          setCreateOpen(false);
          setEditing(null);
        }}
        onConfirm={doSave}
        confirmText={editing ? "保存" : "添加"}
      >
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-item">
            <label>{editing ? "Token（留空保持不变）" : "Token"}</label>
            <input
              className="input"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              placeholder="github_pat_… 或 ghp_…"
              style={{ fontFamily: "monospace" }}
              required={!editing}
            />
            <div className="hint">多个 token 自动故障切换；列表页只显示脱敏值，不会回显明文</div>
          </div>
          <div className="form-item">
            <label>备注</label>
            <input
              className="input"
              value={remark}
              onChange={(e) => setRemark(e.target.value)}
              placeholder="例如：主账号 / 备用 / 自动发现专用…"
            />
          </div>
          <div className="form-item">
            <label>
              <input
                type="checkbox"
                checked={enabled}
                onChange={(e) => setEnabled(e.target.checked)}
                style={{ marginRight: 6 }}
              />
              启用（停用后不再参与轮询）
            </label>
          </div>
        </form>
      </AppDialog>

      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除 Token"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (!confirmDelete) return;
          try {
            await crawlerApi.deleteToken(confirmDelete.id);
            setConfirmDelete(null);
            await load();
            toast.success("Token 已删除");
          } catch {
            // 错误已由全局 Toast 自动提示
          }
        }}
        confirmText="删除"
        danger
      >
        <p>确定删除该 Token 吗？删除后立即从爬虫轮询池中移除。</p>
        {confirmDelete && (
          <code style={{ fontFamily: "monospace", fontSize: 13 }}>{confirmDelete.token}</code>
        )}
      </AppDialog>
    </div>
  );
}
