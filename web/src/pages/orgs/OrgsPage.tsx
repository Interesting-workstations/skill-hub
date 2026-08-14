import { Link } from "react-router-dom";
import { useOfficialOrgs } from "../../hooks/useSkillData";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import OrgAvatar from "../../components/shared/OrgAvatar";
import { useI18n } from "../../i18n";
import "./OrgsPage.css";

export default function OrgsPage() {
  const { data: orgs, loading } = useOfficialOrgs();
  const { t } = useI18n();
  const pageRef = usePageAnimation();

  usePageMeta({ title: `${t("orgs.title")} — ${t("brand.name")}` });

  if (loading) {
    return <PageLoading />;
  }

  const list = orgs ?? [];
  // 有官方技能的组织优先展示（按技能数降序），其余按接口排序
  const sorted = [...list].sort((a, b) => {
    if ((a.officialCount > 0) !== (b.officialCount > 0)) {
      return a.officialCount > 0 ? -1 : 1;
    }
    return b.officialCount - a.officialCount;
  });
  const withSkills = sorted.filter((o) => o.officialCount > 0).length;

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb items={[{ label: t("breadcrumb.home"), to: "/" }, { label: t("orgs.title") }]} />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>
        {t("orgs.title")}
      </h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        {t("orgs.desc", { total: sorted.length, n: withSkills })}
      </p>

      <div className="orgs-grid">
        {sorted.map((org) => (
          <Link
            key={org.owner}
            to={`/official?org=${encodeURIComponent(org.owner)}`}
            className={`org-card${org.officialCount > 0 ? " has-skills" : ""}`}
          >
            <OrgAvatar
              src={org.logoUrl}
              fallback={org.avatar}
              size={40}
              className="org-card-avatar"
            />
            <div className="org-card-info">
              <h3 className="org-card-name">{org.displayName}</h3>
              <p className="org-card-count">
                {org.officialCount > 0
                  ? t("orgs.skillCount", { n: org.officialCount })
                  : t("orgs.pending")}
              </p>
            </div>
          </Link>
        ))}
      </div>
    </PageContainer>
  );
}
