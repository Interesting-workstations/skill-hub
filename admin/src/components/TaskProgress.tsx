/** 执行进度条：百分比 + 颜色状态。 */

interface Props {
  value: number;
  status?: "normal" | "success" | "danger";
}

export default function TaskProgress({ value, status = "normal" }: Props) {
  const clamped = Math.min(100, Math.max(0, value));
  const barCls = status === "success" ? "progress-bar success" : status === "danger" ? "progress-bar danger" : "progress-bar";
  return (
    <div className="progress">
      <div className="progress-track">
        <div className={barCls} style={{ width: `${clamped}%` }} />
      </div>
      <span className="progress-text">{Math.round(clamped)}%</span>
    </div>
  );
}
