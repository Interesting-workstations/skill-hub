import { useEffect, useRef } from "react";
import { useParams, Link } from "react-router-dom";
import { authors } from "../data/skills";
import { pageEnter, createAnimationContext } from "../animations";

export default function AuthorPage() {
  const { authorSlug } = useParams<{ authorSlug: string }>();
  const author = authors.find((a) => a.slug === authorSlug);
  const pageRef = useRef<HTMLDivElement>(null);
  const ctx = useRef(createAnimationContext());

  useEffect(() => {
    if (!pageRef.current) return;
    ctx.current.killAll();
    const tl = pageEnter(pageRef.current);
    ctx.current.add(tl);
    return () => {
      ctx.current.killAll();
    };
  }, [authorSlug]);

  if (!author) {
    return (
      <div className="detail-not-found">
        <h2>作者未找到</h2>
        <Link to="/">返回首页</Link>
      </div>
    );
  }

  return (
    <div ref={pageRef} style={{ maxWidth: 1280, margin: "0 auto", padding: "40px 24px" }}>
      <nav style={{ fontSize: 13, color: "#9ca3af", marginBottom: 32 }}>
        <Link to="/" style={{ color: "#6b7280" }}>Agent Skills 资源库</Link>
        <span style={{ margin: "0 8px" }}>/</span>
        <span style={{ color: "#111827", fontWeight: 500 }}>{author.name}</span>
      </nav>

      <div style={{ display: "flex", alignItems: "center", gap: 20, marginBottom: 24 }}>
        <span style={{ fontSize: 52, background: "#f5f3ff", borderRadius: 16, width: 80, height: 80, display: "flex", alignItems: "center", justifyContent: "center" }}>
          {author.avatar}
        </span>
        <div>
          <h1 style={{ fontSize: 32, fontWeight: 800, margin: "0 0 4px" }}>{author.name}</h1>
          <p style={{ fontSize: 16, color: "#6b7280", margin: 0 }}>{author.skillCount} 个技能</p>
        </div>
      </div>

      <p style={{ color: "#9ca3af", fontSize: 14 }}>
        该作者的技能将在后续版本中展示。
      </p>
    </div>
  );
}
