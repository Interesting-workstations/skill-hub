import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useSkills } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n } from "../../i18n";

export default function FeaturedPage() {
  const { data: skills, loading } = useSkills({ featured: true });
  const { t } = useI18n();
  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [skills?.length]);

  usePageMeta({ title: `${t("featured.title")} — ${t("brand.name")}` });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: t("featured.title") }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>{t("featured.title")}</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        {t("featured.desc", { n: skills?.length ?? 0 })}
      </p>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
        {(skills ?? []).map((s) => (
          <SkillCard key={s.id} skill={s} />
        ))}
      </div>
    </PageContainer>
  );
}
