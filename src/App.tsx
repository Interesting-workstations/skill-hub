import Navbar from "./components/Navbar";
import Hero from "./components/Hero";
import SponsorBanner from "./components/SponsorBanner";
import OfficialAuthors from "./components/OfficialAuthors";
import SkillSection from "./components/SkillSection";
import Footer from "./components/Footer";
import { authors, featuredSkills, skillCategories } from "./data/skills";
import "./App.css";

function App() {
  return (
    <div className="app">
      <Navbar />
      <main className="main">
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
      </main>
      <Footer />
    </div>
  );
}

export default App;
