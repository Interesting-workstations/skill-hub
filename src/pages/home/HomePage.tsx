import Hero from "../../components/sections/Hero";
import SponsorBanner from "../../components/sections/SponsorBanner";
import OfficialAuthors from "../../components/sections/OfficialAuthors";
import SkillSection from "../../components/sections/SkillSection";
import { authors, featuredSkills, skillCategories } from "../../data/skills";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { site } from "../../config/site";

export default function HomePage() {
  usePageMeta({ title: site.title, description: site.description });
  const pageRef = usePageAnimation();

  return (
    <div ref={pageRef}>
      <Hero />
      <SponsorBanner />
      <OfficialAuthors authors={authors} />
      <SkillSection
        title="精选技能"
        count={featuredSkills.length}
        slug="featured"
        skills={featuredSkills}
      />
      {skillCategories.map((cat) => (
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
