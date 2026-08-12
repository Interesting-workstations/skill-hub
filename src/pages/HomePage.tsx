import { usePageEnter } from "../animations";
import Hero from "../components/Hero";
import SponsorBanner from "../components/SponsorBanner";
import OfficialAuthors from "../components/OfficialAuthors";
import SkillSection from "../components/SkillSection";
import { authors, featuredSkills, skillCategories } from "../data/skills";

export default function HomePage() {
  usePageEnter();

  return (
    <>
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
    </>
  );
}
