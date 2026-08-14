import Hero from "../../components/sections/Hero";
import SponsorBanner from "../../components/sections/SponsorBanner";
import OfficialAuthors from "../../components/sections/OfficialAuthors";
import SkillSection from "../../components/sections/SkillSection";
import { useAuthors, useCategories, useSkills } from "../../hooks/useSkillData";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import PageLoading from "../../components/shared/PageLoading";
import { site } from "../../config/site";

export default function HomePage() {
  usePageMeta({ title: site.title, description: site.description });
  const pageRef = usePageAnimation();
  const { data: featuredSkills } = useSkills({ featured: true });
  const { data: categories } = useCategories();
  const { data: authors } = useAuthors();

  // 首页内容依赖数据，未就绪时显示全局加载
  if (!featuredSkills || !categories || !authors) {
    return <PageLoading />;
  }

  return (
    <div ref={pageRef}>
      <Hero />
      <SponsorBanner />
      <OfficialAuthors authors={authors} />
      <SkillSection
        title="精选技能"
        count={featuredSkills.length}
        slug="featured"
        link="/featured"
        skills={featuredSkills}
      />
      {categories.map((cat) => (
        <SkillSection
          key={cat.slug}
          title={cat.name}
          count={cat.count}
          slug={cat.slug}
          skills={cat.skills}
        />
      ))}
    </div>
  );
}
