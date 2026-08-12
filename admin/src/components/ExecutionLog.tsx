/** 执行日志视图：深色终端风格，按级别着色。 */

import type { LogLine } from "../types";

export default function ExecutionLog({ logs }: { logs: LogLine[] }) {
  return (
    <div className="log-view">
      {logs.map((log, i) => (
        <div className="log-line" key={i}>
          <span className="log-time">{log.time}</span>
          <span className={`log-${log.level}`}>{log.text}</span>
        </div>
      ))}
    </div>
  );
}
