import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import StatCard from "../../components/StatCard";
import TaskStatus from "../../components/TaskStatus";
import { siteApi } from "../../api/site";
import { crawlerApi } from "../../api/crawler";
import { contentApi } from "../../api/content";
import type { CrawlTask, ExecutionRecord, FailureRecord, CrawledDataItem, Stats } from "../../types";
import { formatNumber } from "../../utils/format";
import "./DashboardPage.css";

/** 最近 7 天执行趋势（Mock 数据） */
const TREND = [
  { day: "08-06", count: 18, success: 16 },
  { day: "08-07", count: 21, success: 19 },
  { day: "08-08", count: 15, success: 13 },
  { day: "08-09", count: 24, success: 22 },
  { day: "08-10", count: 20, success: 17 },
  { day: "08-11", count: 26, success: 24 },
  { day: "08-12", count: 24, success: 21 },
];

export default function DashboardPage() {
  const [tasks, setTasks] = useState<CrawlTask[]>([]);
  const [executions, setExecutions] = useState<ExecutionRecord[]>([]);
  const [failures, setFailures] = useState<FailureRecord[]>([]);
  const [data, setData] = useState<CrawledDataItem[]>([]);
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let alive = true;
    Promise.all([
      crawlerApi.listTasks(),
      crawlerApi.listExecutions(),
      crawlerApi.listFailures(),
      contentApi.listData(),
      siteApi.stats().catch(() => null),
    ]).then(([t, e, f, d, s]) => {
      if (!alive) return;
      setTasks(t);
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

  const statusCount = (s: string) => tasks.filter((t) => t.status === s).length;
  const pending = data.filter((d) => d.status === "pending").length;
  const maxTrend = Math.max(...TREND.map((t) => t.count));

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>工作台</h1>
          <div className="sub">网站与爬虫运行概览 · {new Date().toLocaleDateString("zh-CN")}</div>
        </div>
      </div>

      {/* 爬虫核心指标 */}
      <div className="stat-grid">
        <StatCard label="今日爬虫任务" value={24} extra="较昨日 +2" extraTone="ok" />
        <StatCard label="执行成功" value={21} extra="成功率 87.5%" extraTone="ok" />
        <StatCard label="执行失败" value={2} extra="需处理失败任务" extraTone="bad" />
        <StatCard label="执行中" value={statusCount("running")} extra="实时监控中" />
      </div>
      <div className="stat-grid">
        <StatCard label="新增数据" value={formatNumber(1284)} extra="近 24 小时" />
        <StatCard label="待审核" value={pending} extra="等待人工审核" />
        <StatCard label="收录技能" value={stats ? formatNumber(stats.totalSkills) : "--"} extra="官网数据库" />
        <StatCard label="官方技能" value={stats ? formatNumber(stats.officialSkills) : "--"} extra={`${stats?.totalAuthors ?? 0} 位官方作者`} />
      </div>

      {/* 趋势 */}
      <div className="card card-pad">
        <div className="card-title">近 7 天爬虫任务执行趋势</div>
        <div className="trend-chart">
          {TREND.map((t) => (
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
                {d.title}
                {d.isOfficial && <span className="badge badge-neutral">官方</span>}
                <span className="badge badge-neutral">{d.category}</span>
              </span>
              <span className="time">{d.fetchedAt}</span>
            </div>
          ))}
        </div>
        <div className="dash-more"><Link to="/data/items">查看全部数据 →</Link></div>
      </div>
    </div>
  );
}
