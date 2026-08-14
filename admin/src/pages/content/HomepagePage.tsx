import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import StatCard from "../../components/StatCard";
import { siteApi } from "../../api/site";
import type { Author, Skill, Stats } from "../../types";
import { formatNumber } from "../../utils/format";

export default function HomepagePage() {
  const [stats, setStats] = useState<Stats | null>(null);
  const [authors, setAuthors] = useState<Author[]>([]);
  const [featured, setFeatured] = useState<Skill[]>([]);

  useEffect(() => {
    siteApi.stats().then(setStats).catch(() => null);
    siteApi.authors().then(setAuthors).catch(() => []);
    siteApi.skills({ featured: true }).then(setFeatured).catch(() => []);
  }, []);

  const officialAuthors = authors.filter((a) => (a.officialSkills ?? 0) > 0);

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>首页内容</h1>
          <div className="sub">官网首页的数据展示 —— 全部来自 skill-hub 后端真实统计</div>
        </div>
      </div>

      {/* 概览 */}
      <div className="stat-grid">
        <StatCard label="收录技能" value={stats ? formatNumber(stats.totalSkills) : "--"} extra="官网数据库" />
        <StatCard label="官方作者" value={stats ? formatNumber(stats.totalAuthors) : "--"} extra="官方区块展示" />
        <StatCard label="官方技能" value={stats ? formatNumber(stats.officialSkills) : "--"} extra="官方区块展示" />
        <StatCard label="精选技能" value={stats ? formatNumber(stats.featuredSkills) : "--"} extra="首页精选区块" />
      </div>

      {/* 官方作者 */}
      <div className="card card-pad">
        <div className="card-title">官方作者区块（官网首页展示）</div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(220px, 1fr))", gap: 12 }}>
          {officialAuthors.map((a) => (
            <div className="card" key={a.slug} style={{ padding: 14, display: "flex", alignItems: "center", gap: 12 }}>
              <span style={{ fontSize: 24 }}>{a.avatar}</span>
              <div>
                <div style={{ fontWeight: 500 }}>{a.name}</div>
                <div style={{ fontSize: 12, color: "var(--color-text-tertiary)" }}>
                  {a.officialSkills} 个官方技能
                </div>
              </div>
            </div>
          ))}
          {officialAuthors.length === 0 && <div className="empty">暂无官方作者</div>}
        </div>
      </div>

      {/* 精选技能 */}
      <div className="card card-pad">
        <div className="card-title">精选技能区块（规则生成：官方优先 + 星标 + 分类多样）</div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: 12 }}>
          {featured.map((s) => (
            <div className="card" key={s.id} style={{ padding: 14 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <span style={{ fontWeight: 500 }}>{s.name}</span>
                {s.isOfficial && <span className="badge badge-neutral">官方</span>}
              </div>
              <div style={{ marginTop: 6, fontSize: 12, color: "var(--color-text-tertiary)" }}>
                {s.author} · {s.category}
                {s.githubStars ? ` · ⭐ ${s.githubStars}` : ""}
              </div>
              <div style={{ marginTop: 8, fontSize: 13, lineHeight: 1.6, display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
                {s.description}
              </div>
              <div style={{ marginTop: 10 }}>
                <Link to={`/skill/${s.id}`} className="btn-link" target="_blank">官网查看 →</Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
