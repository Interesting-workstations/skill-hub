import { useParams, Link } from "react-router-dom";
import { authors } from "../../data/skills";
import { getSkillsByAuthor } from "../../data/queries";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { site } from "../../config/site";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";

export default function AuthorPage() {
  const { authorSlug } = useParams<{ authorSlug: string }>();
  const author = authors.find((a) => a.slug === authorSlug);
  const skills = author ? getSkillsByAuthor(author.name) : [];

  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [authorSlug]);

  usePageMeta({
    title: author ? `${author.name} — ${site.name}` : site.title,
    description: author ? `浏览 ${author.name} 发布的全部技能` : undefined,
  });

  if (!author) {
    return (
      <div className="detail-not-found">
        <h2>作者未找到</h2>
        <Link to="/">返回首页</Link>
      </div>
    );
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: author.name }]} />

      <div style={{ display: "flex", alignItems: "center", gap: 20, marginBottom: 36 }}>
        <span style={{ fontSize: 48, background: "var(--color-primary-50)", borderRadius: 16, width: 80, height: 80, display: "flex", alignItems: "center", justifyContent: "center", flexShrink: 0 }}>
          {author.avatar}
        </span>
        <div>
          <h1 style={{ fontSize: 32, fontWeight: 800, margin: "0 0 4px", color: "var(--color-text)" }}>{author.name}</h1>
          <p style={{ fontSize: 16, color: "var(--color-text-secondary)", margin: 0 }}>
            共发布 <strong style={{ color: "var(--color-text)" }}>{skills.length}</strong> 个技能
          </p>
        </div>
      </div>

      {skills.length > 0 ? (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
          {skills.map((s) => (
            <SkillCard key={s.id} skill={s} />
          ))}
        </div>
      ) : (
        <p style={{ color: "var(--color-text-muted)", fontSize: 14 }}>该作者暂无已收录的技能。</p>
      )}
    </PageContainer>
  );
}
