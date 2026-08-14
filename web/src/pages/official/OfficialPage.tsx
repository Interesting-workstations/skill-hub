import { useSearchParams, Link } from "react-router-dom";
import SkillCard from "../../components/skill/SkillCard";
import OrgAvatar from "../../components/shared/OrgAvatar";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useOfficialOrgs, useSkills } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n } from "../../i18n";

export default function OfficialPage() {
  const { t } = useI18n();
  const [searchParams] = useSearchParams();
  const orgOwner = searchParams.get("org");

  // 支持按官方组织筛选：/official?org=anthropics
  const { data: orgs } = useOfficialOrgs();
  const currentOrg = orgs?.find((o) => o.owner === orgOwner);
  const { data: skills, loading } = useSkills(
    currentOrg ? { official: true, author: currentOrg.displayName } : { official: true }
  );

  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [skills?.length]);

  const title = currentOrg ? currentOrg.displayName : t("officialPage.title");
  usePageMeta({ title: `${title} — ${t("brand.name")}` });

  if (loading) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb
        items={[
          { label: t("breadcrumb.home"), to: "/" },
          // 组织筛选时父级为「官方组织」页，保证导航链路一致：官方组织 → 该组织官方技能
          currentOrg
            ? { label: t("orgs.title"), to: "/orgs" }
            : { label: t("officialPage.title"), to: "/official" },
          ...(currentOrg ? [{ label: currentOrg.displayName }] : []),
        ]}
      />

      {currentOrg ? (
        <div style={{ display: "flex", alignItems: "center", gap: 14, margin: "0 0 8px" }}>
          <OrgAvatar src={currentOrg.logoUrl} fallback={currentOrg.avatar} size={44} />
          <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: 0 }}>
            {currentOrg.displayName}
          </h1>
        </div>
      ) : (
        <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>
          {t("officialPage.title")}
        </h1>
      )}
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 20px" }}>
        {currentOrg
          ? t("officialPage.orgDesc", { name: currentOrg.displayName, n: skills?.length ?? 0 })
          : t("officialPage.desc", { n: skills?.length ?? 0 })}
      </p>
      {currentOrg && (
        <p style={{ margin: "0 0 20px" }}>
          <Link to="/orgs" style={{ color: "var(--color-primary)", textDecoration: "none", fontSize: 14 }}>
            ← {t("officialPage.backOrgs")}
          </Link>
        </p>
      )}

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
        {(skills ?? []).map((s) => (
          <SkillCard key={s.id} skill={s} />
        ))}
      </div>
    </PageContainer>
  );
}
