/** 统计卡片：工作台核心指标。 */

import type { ReactNode } from "react";

interface Props {
  label: string;
  value: ReactNode;
  extra?: ReactNode;
  extraTone?: "default" | "ok" | "bad";
}

export default function StatCard({ label, value, extra, extraTone = "default" }: Props) {
  const extraCls =
    extraTone === "ok" ? "extra ok" : extraTone === "bad" ? "extra bad" : "extra";
  return (
    <div className="card stat-card">
      <div className="label">{label}</div>
      <div className="value">{value}</div>
      {extra && <div className={extraCls}>{extra}</div>}
    </div>
  );
}
