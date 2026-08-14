import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import AppTable, { type Column } from "../../components/AppTable";
import TaskStatus from "../../components/TaskStatus";
import { crawlerApi } from "../../api/crawler";
import type { ExecutionRecord } from "../../types";

export default function ExecutionListPage() {
  const [records, setRecords] = useState<ExecutionRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  const load = () => {
    crawlerApi.listExecutions().then((r) => {
      setRecords(r);
      setLoading(false);
    });
  };

  useEffect(load, []);

  // 停止运行中的执行（后端取消爬虫 goroutine 并标记 stopped）
  const stopExecution = (e: ExecutionRecord) => {
    if (!window.confirm(`确定停止执行「${e.taskName}」吗？`)) return;
    crawlerApi.stopExecution(e.id).then(() => load());
  };

  // 删除执行记录
  const deleteExecution = (e: ExecutionRecord) => {
    if (!window.confirm(`确定删除执行记录「${e.taskName}」吗？删除后不可恢复。`)) return;
    crawlerApi.deleteExecution(e.id).then(() => load());
  };

  const columns: Column<ExecutionRecord>[] = [
    {
      key: "taskName",
      title: "任务",
      render: (e) => (
        <Link to={`/crawler/executions/${e.id}`} style={{ fontWeight: 500 }}>
          {e.taskName}
        </Link>
      ),
    },
    {
      key: "status",
      title: "状态",
      render: (e) => <TaskStatus status={e.status} />,
    },
    {
      key: "startTime",
      title: "开始时间",
      render: (e) => <span className="num">{e.startTime}</span>,
    },
    {
      key: "duration",
      title: "耗时",
      render: (e) => <span className="num">{e.duration}</span>,
    },
    {
      key: "progress",
      title: "进度",
      render: (e) => {
        const cls = e.status === "success" ? "success" : e.status === "failed" ? "danger" : "normal";
        return (
          <div style={{ width: 180 }}>
            <div className="progress">
              <div className="progress-track">
                <div
                  className={`progress-bar ${cls === "success" ? "success" : cls === "danger" ? "danger" : ""}`}
                  style={{ width: `${e.progress}%` }}
                />
              </div>
              <span className="progress-text">{e.progress}%</span>
            </div>
          </div>
        );
      },
    },
    {
      key: "stats",
      title: "结果",
      render: (e) => (
        <div style={{ fontSize: 13 }}>
          抓取 {e.stats.pages} · 新增 {e.stats.newData}
        </div>
      ),
    },
    {
      key: "action",
      title: "操作",
      render: (e) => (
        <div className="ops">
          <Link to={`/crawler/executions/${e.id}`} className="btn-link">查看详情</Link>
          {e.status === "running" && (
            <button className="btn-link danger" onClick={() => stopExecution(e)}>停止</button>
          )}
          <button className="btn-link danger" onClick={() => deleteExecution(e)}>删除</button>
        </div>
      ),
    },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>执行记录</h1>
          <div className="sub">查看每次爬虫执行的进度、结果与日志</div>
        </div>
      </div>
      <AppTable
        columns={columns}
        data={records}
        rowKey={(e) => e.id}
        loading={loading}
        page={page}
        pageSize={10}
        onPageChange={setPage}
      />
    </div>
  );
}
