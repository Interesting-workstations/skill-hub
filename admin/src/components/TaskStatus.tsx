/** 任务状态徽章：成功/失败/运行中/等待/已停止 一眼可识别。 */

import type { TaskStatus } from "../types";

const STATUS_MAP: Record<TaskStatus, { label: string; cls: string }> = {
  waiting: { label: "等待", cls: "badge-waiting" },
  running: { label: "运行中", cls: "badge-running" },
  success: { label: "成功", cls: "badge-success" },
  failed: { label: "失败", cls: "badge-danger" },
  stopped: { label: "已停止", cls: "badge-stopped" },
};

export default function TaskStatus({ status }: { status: TaskStatus }) {
  const s = STATUS_MAP[status] ?? STATUS_MAP.waiting;
  return (
    <span className={`badge ${s.cls}`}>
      <span className="dot" />
      {s.label}
    </span>
  );
}
