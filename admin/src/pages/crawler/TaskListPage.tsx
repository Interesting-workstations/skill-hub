import { useEffect, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import AppTable, { type Column } from "../../components/AppTable";
import AppDialog from "../../components/AppDialog";
import TaskStatus from "../../components/TaskStatus";
import { crawlerApi } from "../../api/crawler";
import type { CrawlTask } from "../../types";

const TASK_TYPES = ["skill", "info", "data", "media", "news", "product"];

export default function TaskListPage() {
  const navigate = useNavigate();
  const [tasks, setTasks] = useState<CrawlTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [keyword, setKeyword] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<CrawlTask | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<CrawlTask | null>(null);

  // 表单
  const [name, setName] = useState("");
  const [type, setType] = useState("skill");
  const [query, setQuery] = useState("");
  const [schedule, setSchedule] = useState("手动");

  const load = async () => {
    setLoading(true);
    const data = await crawlerApi.listTasks();
    setTasks(data);
    setLoading(false);
  };

  useEffect(() => {
    void load();
  }, []);

  const filtered = tasks.filter((t) => {
    if (keyword && !t.name.toLowerCase().includes(keyword.toLowerCase()) && !t.query.toLowerCase().includes(keyword.toLowerCase())) return false;
    if (statusFilter && t.status !== statusFilter) return false;
    if (typeFilter && t.type !== typeFilter) return false;
    return true;
  });

  const openCreate = () => {
    setEditing(null);
    setName("");
    setType("skill");
    setQuery("");
    setSchedule("手动");
    setDialogOpen(true);
  };

  const openEdit = (task: CrawlTask) => {
    setEditing(task);
    setName(task.name);
    setType(task.type);
    setQuery(task.query);
    setSchedule(task.schedule);
    setDialogOpen(true);
  };

  const doSave = async () => {
    if (editing) {
      await crawlerApi.updateTask(editing.id, { name, type, query, schedule });
    } else {
      await crawlerApi.createTask({ name, type, query, schedule });
    }
    setDialogOpen(false);
    await load();
  };

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    void doSave();
  };

  const handleRun = async (task: CrawlTask) => {
    const record = await crawlerApi.runTask(task.id);
    navigate(`/crawler/executions/${record.id}`);
  };

  const columns: Column<CrawlTask>[] = [
    {
      key: "name",
      title: "任务名称",
      render: (t) => (
        <div>
          <div style={{ fontWeight: 500 }}>{t.name}</div>
          <div style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>{t.query}</div>
        </div>
      ),
    },
    {
      key: "type",
      title: "类型",
      render: (t) => <span className="badge badge-neutral">{t.type}</span>,
    },
    {
      key: "status",
      title: "状态",
      render: (t) => <TaskStatus status={t.status} />,
    },
    {
      key: "lastRunAt",
      title: "最后执行",
      render: (t) => (
        <div>
          <div>{t.lastRunAt}</div>
          <div style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>{t.lastDuration}</div>
        </div>
      ),
    },
    {
      key: "runCount",
      title: "累计",
      render: (t) => (
        <div style={{ fontSize: 13 }}>
          共 {t.runCount} 次 · 成功 {t.successCount}
        </div>
      ),
    },
    {
      key: "actions",
      title: "操作",
      width: "240px",
      render: (t) => (
        <div style={{ display: "flex", gap: 4 }}>
          {t.status === "running" ? (
            <button className="btn-link danger" onClick={() => crawlerApi.stopTask(t.id).then(load)}>停止</button>
          ) : (
            <button className="btn-link" onClick={() => handleRun(t)}>执行</button>
          )}
          {t.status === "failed" && (
            <button className="btn-link" onClick={() => handleRun(t)}>重试</button>
          )}
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
          <h1>爬虫任务</h1>
          <div className="sub">管理爬虫任务，随时执行 / 停止 / 重试</div>
        </div>
        <button className="btn btn-primary" onClick={openCreate}>+ 新建任务</button>
      </div>

      <AppTable
        columns={columns}
        data={filtered}
        rowKey={(t) => t.id}
        loading={loading}
        pageSize={10}
        toolbar={
          <>
            <div className="filters">
              <input
                className="input"
                placeholder="搜索任务名称 / 关键词"
                value={keyword}
                onChange={(e) => setKeyword(e.target.value)}
                style={{ width: 220 }}
              />
              <select className="select" value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
                <option value="">全部状态</option>
                <option value="waiting">等待</option>
                <option value="running">运行中</option>
                <option value="success">成功</option>
                <option value="failed">失败</option>
                <option value="stopped">已停止</option>
              </select>
              <select className="select" value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
                <option value="">全部类型</option>
                {TASK_TYPES.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
          </>
        }
      />

      {/* 新建 / 编辑弹窗 */}
      <AppDialog
        open={dialogOpen}
        title={editing ? "编辑任务" : "新建任务"}
        onClose={() => setDialogOpen(false)}
        onConfirm={doSave}
        confirmText="保存"
      >
        <form className="form" onSubmit={handleSubmit}>
          <div className="form-item">
            <label>任务名称</label>
            <input className="input" value={name} onChange={(e) => setName(e.target.value)} placeholder="如：官方技能采集" required />
          </div>
          <div className="form-grid">
            <div className="form-item">
              <label>类型</label>
              <select className="select" value={type} onChange={(e) => setType(e.target.value)} style={{ width: "100%" }}>
                {TASK_TYPES.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div className="form-item">
              <label>执行计划</label>
              <input className="input" value={schedule} onChange={(e) => setSchedule(e.target.value)} placeholder="每天 02:00 / 手动" />
            </div>
          </div>
          <div className="form-item">
            <label>搜索关键词 / 目标仓库</label>
            <input className="input" value={query} onChange={(e) => setQuery(e.target.value)} placeholder="claude skills 或 anthropics/skills" required />
            <span className="hint">支持逗号分隔多个仓库，如 anthropics/skills,openai/codex</span>
          </div>
        </form>
      </AppDialog>

      {/* 删除确认 */}
      <AppDialog
        open={Boolean(confirmDelete)}
        title="删除任务"
        onClose={() => setConfirmDelete(null)}
        onConfirm={async () => {
          if (confirmDelete) {
            await crawlerApi.deleteTask(confirmDelete.id);
            setConfirmDelete(null);
            await load();
          }
        }}
        confirmText="确认删除"
        danger
      >
        <p>确定要删除任务「{confirmDelete?.name}」吗？此操作不可恢复。</p>
      </AppDialog>
    </div>
  );
}
