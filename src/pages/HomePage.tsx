import { useEffect, useRef } from "react";
import Hero from "../components/Hero";
import SponsorBanner from "../components/SponsorBanner";
import OfficialAuthors from "../components/OfficialAuthors";
import SkillSection from "../components/SkillSection";
import { authors, featuredSkills, skillCategories } from "../data/skills";
import { pageEnter, createAnimationContext } from "../animations";

export default function HomePage() {
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
  }, []);

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
