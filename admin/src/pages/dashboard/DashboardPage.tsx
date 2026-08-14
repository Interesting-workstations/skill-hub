import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import StatCard from "../../components/StatCard";
import TaskStatus from "../../components/TaskStatus";
import { crawlerApi } from "../../api/crawler";
import { contentApi } from "../../api/content";
import type { AdminStats, ExecutionRecord, FailureRecord, CrawledDataItem } from "../../types";
import { formatNumber } from "../../utils/format";
import "./DashboardPage.css";

export default function DashboardPage() {
  const [executions, setExecutions] = useState<ExecutionRecord[]>([]);
  const [failures, setFailures] = useState<FailureRecord[]>([]);
  const [data, setData] = useState<CrawledDataItem[]>([]);
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    Promise.all([
      crawlerApi.listExecutions(),
      crawlerApi.listFailures(),
      contentApi.listData(),
      crawlerApi.stats().catch(() => null),
    ]).then(([e, f, d, s]) => {
      if (!alive) return;
      setExecutions(e);
      setFailures(f);
      setData(d);
      setStats(s);
      setLoading(false);
    });
    return () => {
      alive = false;
    };
  }, []);

  if (loading) {
    return (
      <div className="page">
        <div className="loading"><span className="spin" />加载中…</div>
      </div>
    );
  }

  const trend = stats?.trend ?? [];
  const maxTrend = Math.max(1, ...trend.map((t) => t.count));

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>工作台</h1>
          <div className="sub">网站与爬虫运行概览 · {new Date().toLocaleDateString("zh-CN")}</div>
        </div>
      </div>

      {/* 爬虫核心指标（Go 后端实时统计） */}
      <div className="stat-grid">
        <StatCard label="爬虫任务" value={stats?.todayTasks ?? 0} extra="全部任务" />
        <StatCard label="执行成功" value={stats?.runSuccess ?? 0} extra="累计成功记录" extraTone="ok" />
        <StatCard label="失败记录" value={stats?.runFailed ?? 0} extra="需处理失败任务" extraTone={stats && stats.runFailed > 0 ? "bad" : "ok"} />
        <StatCard label="执行中" value={stats?.runRunning ?? 0} extra="实时监控中" />
      </div>
      <div className="stat-grid">
        <StatCard label="待审核数据" value={stats?.pendingData ?? 0} extra="等待人工审核" />
        <StatCard label="收录技能" value={stats ? formatNumber(stats.totalSkills) : "--"} extra="官网数据库" />
        <StatCard label="官方技能" value={stats ? formatNumber(stats.officialSkills) : "--"} extra={`${stats?.totalAuthors ?? 0} 位官方作者`} />
        <StatCard label="官网分类" value={stats ? formatNumber(stats.totalCategories ?? 0) : "--"} extra="官网展示" />
      </div>

      {/* 趋势 */}
      <div className="card card-pad">
        <div className="card-title">近 7 天爬虫任务执行趋势</div>
        {trend.length === 0 ? (
          <div className="empty">暂无执行数据</div>
        ) : (
          <div className="trend-chart">
            {trend.map((t) => (
              <div className="trend-col" key={t.day}>
                <div className="val">{t.count}</div>
                <div
                  className="trend-bar"
                  style={{ height: `${(t.count / maxTrend) * 100}%` }}
                  title={`${t.day}：${t.count} 次，成功 ${t.success}`}
                />
                <div className="label">{t.day}</div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 数据看板：状态分布 + 来源分布 */}
      <div className="card card-pad">
        <div className="card-title">数据看板</div>
        <div className="dash-grid-2">
          <div>
            <div style={{ fontSize: 13, color: "var(--color-text-secondary)", marginBottom: 10 }}>数据状态分布</div>
            {(
              [
                { key: "pending", label: "待审核", color: "#f59e0b" },
                { key: "approved", label: "已批准", color: "#3b82f6" },
                { key: "published", label: "已发布", color: "#22c55e" },
                { key: "ignored", label: "已忽略", color: "#9ca3af" },
              ] as const
            ).map((m) => {
              const n = stats?.statusDist?.[m.key] ?? 0;
              const pct = stats && stats.totalSkills > 0 ? Math.round((n / stats.totalSkills) * 100) : 0;
              return (
                <div key={m.key} style={{ marginBottom: 10 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12, marginBottom: 4 }}>
                    <span style={{ color: "var(--color-text-secondary)" }}>{m.label}</span>
                    <span style={{ fontWeight: 600 }}>{n}（{pct}%）</span>
                  </div>
                  <div style={{ height: 8, borderRadius: 4, background: "var(--color-surface-hover)", overflow: "hidden" }}>
                    <div style={{ height: "100%", width: `${pct}%`, background: m.color, borderRadius: 4 }} />
                  </div>
                </div>
              );
            })}
          </div>
          <div>
            <div style={{ fontSize: 13, color: "var(--color-text-secondary)", marginBottom: 10 }}>数据来源分布</div>
            {(
              [
                { key: "official", label: "官方来源", color: "#8b5cf6" },
                { key: "community", label: "社区个人", color: "#64748b" },
              ] as const
            ).map((m) => {
              const n = stats?.typeDist?.[m.key] ?? 0;
              const pct = stats && stats.totalSkills > 0 ? Math.round((n / stats.totalSkills) * 100) : 0;
              return (
                <div key={m.key} style={{ marginBottom: 10 }}>
                  <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12, marginBottom: 4 }}>
                    <span style={{ color: "var(--color-text-secondary)" }}>{m.label}</span>
                    <span style={{ fontWeight: 600 }}>{n}（{pct}%）</span>
                  </div>
                  <div style={{ height: 8, borderRadius: 4, background: "var(--color-surface-hover)", overflow: "hidden" }}>
                    <div style={{ height: "100%", width: `${pct}%`, background: m.color, borderRadius: 4 }} />
                  </div>
                </div>
              );
            })}
            <div style={{ marginTop: 14, fontSize: 12, color: "var(--color-text-tertiary)" }}>
              官方来源技能自动发布，社区来源需人工审核
            </div>
          </div>
        </div>
      </div>

      {/* 最近执行 + 最近失败 */}
      <div className="dash-grid-2">
        <div className="card card-pad">
          <div className="card-title">最近执行任务</div>
          <div className="dash-list">
            {executions.slice(0, 5).map((e) => (
              <Link to={`/crawler/executions/${e.id}`} className="dash-list-row" key={e.id} style={{ textDecoration: "none" }}>
                <span className="name">{e.taskName} <TaskStatus status={e.status} /></span>
                <span className="time">{e.startTime}</span>
              </Link>
            ))}
          </div>
          <div className="dash-more"><Link to="/crawler/executions">全部执行记录 →</Link></div>
        </div>
        <div className="card card-pad">
          <div className="card-title">最近失败任务</div>
          <div className="dash-list">
            {failures.slice(0, 5).map((f) => (
              <Link to="/crawler/failures" className="dash-list-row" key={f.id} style={{ textDecoration: "none" }}>
                <span className="name">
                  <span className="badge badge-danger"><span className="dot" />失败</span>
                  {f.taskName}
                </span>
                <span className="err">{f.reason}</span>
              </Link>
            ))}
          </div>
          <div className="dash-more"><Link to="/crawler/failures">全部失败任务 →</Link></div>
        </div>
      </div>

      {/* 最近抓取数据 */}
      <div className="card card-pad">
        <div className="card-title">最近抓取数据</div>
        <div className="dash-list">
          {data.slice(0, 6).map((d) => (
            <div className="dash-list-row" key={d.id}>
              <span className="name">
                {d.name}
                {d.isOfficial && <span className="badge badge-neutral">官方</span>}
                <span className="badge badge-neutral">{d.category}</span>
              </span>
              <span className="time">{d.source}</span>
            </div>
          ))}
        </div>
        <div className="dash-more"><Link to="/data/items">查看全部数据 →</Link></div>
      </div>
    </div>
  );
}
