import { useEffect, useRef, useState } from "react";
import { Link, useParams } from "react-router-dom";
import TaskStatus from "../../components/TaskStatus";
import TaskProgress from "../../components/TaskProgress";
import ExecutionLog from "../../components/ExecutionLog";
import { crawlerApi } from "../../api/crawler";
import type { ExecutionRecord, LogLine } from "../../types";

/** WebSocket 推送事件 */
interface ExecEvent {
  type: "snapshot" | "log" | "progress" | "status";
  execId: string;
  progress?: number;
  step?: string;
  status?: ExecutionRecord["status"];
  log?: LogLine;
  logs?: LogLine[];
  duration?: string;
}

type WsState = "connecting" | "open" | "closed";

function isTerminal(status?: ExecutionRecord["status"]): boolean {
  return status === "success" || status === "failed";
}

export default function ExecutionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [record, setRecord] = useState<ExecutionRecord | null>(null);
  const [step, setStep] = useState("");
  const [wsState, setWsState] = useState<WsState>("connecting");
  const retryRef = useRef(0);
  const closedRef = useRef(false); // 任务终态后不再重连

  useEffect(() => {
    if (!id) return;
    // 重置连接状态（React StrictMode 开发环境会双调用 effect，首次 cleanup 后需重置）
    closedRef.current = false;
    retryRef.current = 0;
    let disposed = false;
    let ws: WebSocket | null = null;
    let retryTimer: ReturnType<typeof setTimeout> | null = null;

    const connect = async () => {
      if (disposed || closedRef.current) return;
      setWsState("connecting");
      try {
        const { ticket } = await crawlerApi.wsTicket(id!);
        const proto = window.location.protocol === "https:" ? "wss" : "ws";
        ws = new WebSocket(
          `${proto}://${window.location.host}/api/v1/admin/executions/${id}/ws?ticket=${ticket}`
        );

        ws.onopen = () => {
          if (disposed) return;
          setWsState("open");
          retryRef.current = 0;
        };

        ws.onmessage = (e) => {
          if (disposed) return;
          let ev: ExecEvent;
          try {
            ev = JSON.parse(e.data) as ExecEvent;
          } catch {
            return;
          }
          switch (ev.type) {
            case "snapshot":
              setRecord((prev) =>
                prev
                  ? {
                      ...prev,
                      progress: ev.progress ?? prev.progress,
                      status: ev.status ?? prev.status,
                      logs: ev.logs ?? prev.logs,
                      duration: ev.duration ?? prev.duration,
                    }
                  : prev
              );
              if (ev.status && isTerminal(ev.status)) closedRef.current = true;
              break;
            case "log":
              if (ev.log) {
                setRecord((prev) => (prev ? { ...prev, logs: [...prev.logs, ev.log!] } : prev));
              }
              break;
            case "progress":
              setStep(ev.step ?? "");
              setRecord((prev) => (prev ? { ...prev, progress: ev.progress ?? prev.progress } : prev));
              break;
            case "status":
              setRecord((prev) => (prev ? { ...prev, status: ev.status ?? prev.status } : prev));
              if (ev.status && isTerminal(ev.status)) closedRef.current = true;
              break;
          }
        };

        ws.onclose = () => {
          if (disposed) return;
          setWsState("closed");
          if (!closedRef.current) {
            // 指数退避重连（1s → 2s → 4s …，上限 15s）
            const delay = Math.min(1000 * 2 ** retryRef.current, 15000);
            retryRef.current++;
            retryTimer = setTimeout(connect, delay);
          }
        };
      } catch {
        if (!disposed && !closedRef.current) {
          retryTimer = setTimeout(connect, 3000);
        }
      }
    };

    // 先拉取一次快照（含任务名 / 统计），再建立 WebSocket 推送增量
    crawlerApi
      .executionDetail(id)
      .then((r) => {
        if (disposed) return;
        setRecord(r);
        // 终态任务无需连接实时推送，直接显示「已断开」
        if (isTerminal(r.status)) {
          closedRef.current = true;
          setWsState("closed");
          return;
        }
        connect();
      })
      .catch(() => {
        if (!disposed) connect();
      });

    return () => {
      disposed = true;
      closedRef.current = true;
      if (retryTimer) clearTimeout(retryTimer);
      if (ws) ws.close();
    };
  }, [id]);

  if (!record) {
    return (
      <div className="page">
        <div className="loading">
          <span className="spin" />
          加载中…
        </div>
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
        <span
          className={`ws-badge ${wsState === "open" ? "on" : wsState === "connecting" ? "ing" : "off"}`}
          title={wsState === "open" ? "实时推送已连接" : wsState === "connecting" ? "实时推送连接中…" : "连接已断开"}
        >
          <span className="ws-dot" />
          {wsState === "open" ? "实时推送" : wsState === "connecting" ? "连接中" : "已断开"}
        </span>
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
          <div className="desc-item"><span className="k">执行进度</span><span className="v" style={{ width: 240 }}>
            <TaskProgress
              value={record.progress}
              status={record.status === "success" ? "success" : record.status === "failed" ? "danger" : "normal"}
            />
          </span></div>
        </div>
        {step && <div className="exec-step">▸ {step}</div>}
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
