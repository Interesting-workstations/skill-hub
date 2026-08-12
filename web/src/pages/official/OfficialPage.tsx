import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useSkills } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { site } from "../../config/site";

export default function OfficialPage() {
  const { data: skills, loading } = useSkills({ official: true });
  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [skills?.length]);

  usePageMeta({ title: `官方技能 — ${site.name}` });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "官方技能" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>官方技能</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        由官方维护和认证的高质量技能，共 <strong style={{ color: "var(--color-text)" }}>{skills?.length ?? 0}</strong> 个
      </p>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
        {(skills ?? []).map((s) => (
          <SkillCard key={s.id} skill={s} />
        ))}
      </div>
    </PageContainer>
  );
}
