import { featuredSkills } from "../../data/skills";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { site } from "../../config/site";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";

export default function FeaturedPage() {
  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  });

  usePageMeta({ title: `精选技能 — ${site.name}` });

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: "Agent Skills 资源库", to: "/" }, { label: "精选技能" }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>精选技能</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        编辑精选的高质量技能，共 <strong style={{ color: "var(--color-text)" }}>{featuredSkills.length}</strong> 个
      </p>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
        {featuredSkills.map((s) => (
          <SkillCard key={s.id} skill={s} />
        ))}
      </div>
    </PageContainer>
  );
}
