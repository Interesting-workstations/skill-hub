import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import TaskStatus from "../../components/TaskStatus";
import TaskProgress from "../../components/TaskProgress";
import ExecutionLog from "../../components/ExecutionLog";
import { crawlerApi } from "../../api/crawler";
import type { ExecutionRecord } from "../../types";

export default function ExecutionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [record, setRecord] = useState<ExecutionRecord | null>(null);

  useEffect(() => {
    if (!id) return;
    crawlerApi.executionDetail(id).then(setRecord);
  }, [id]);

  if (!record) {
    return (
      <div className="page">
        <div className="loading"><span className="spin" />加载中…</div>
      </div>
    );
  }

  const s = record.stats;
  const resultCards = [
    { label: "抓取页面", value: s.pages },
    { label: "成功页面", value: s.fetched },
    { label: "失败页面", value: s.failed, tone: s.failed > 0 ? "bad" : "ok" },
    { label: "新增数据", value: s.newData, tone: "ok" },
    { label: "更新数据", value: s.updated },
    { label: "重复数据", value: s.duplicate },
  ];

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>
            执行详情：{record.taskName} <TaskStatus status={record.status} />
          </h1>
          <div className="sub">
            <Link to="/crawler/executions">← 返回执行记录</Link>
          </div>
        </div>
      </div>

      {/* 基本信息 */}
      <div className="card card-pad">
        <div className="card-title">基本信息</div>
        <div className="desc-grid">
          <div className="desc-item"><span className="k">任务名称</span><span className="v">{record.taskName}</span></div>
          <div className="desc-item"><span className="k">执行状态</span><span className="v"><TaskStatus status={record.status} /></span></div>
          <div className="desc-item"><span className="k">开始时间</span><span className="v num">{record.startTime}</span></div>
          <div className="desc-item"><span className="k">结束时间</span><span className="v num">{record.endTime}</span></div>
          <div className="desc-item"><span className="k">执行耗时</span><span className="v num">{record.duration}</span></div>
          <div className="desc-item"><span className="k">执行进度</span><span className="v" style={{ width: 240 }}><TaskProgress value={record.progress} status={record.status === "success" ? "success" : record.status === "failed" ? "danger" : "normal"} /></span></div>
        </div>
      </div>

      {/* 执行结果 */}
      <div className="card card-pad">
        <div className="card-title">执行结果</div>
        <div className="stat-grid">
          {resultCards.map((c) => (
            <div className="card stat-card" key={c.label}>
              <div className="label">{c.label}</div>
              <div className="value" style={{ fontSize: 22 }}>{c.value}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 执行日志 */}
      <div className="card card-pad">
        <div className="card-title">执行日志</div>
        <ExecutionLog logs={record.logs} />
      </div>
    </div>
  );
}
