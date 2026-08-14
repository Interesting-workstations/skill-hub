import { useState, useRef, useEffect } from "react";
import { Link } from "react-router-dom";
import { buttonHoverEnter, buttonHoverLeave, buttonClick } from "../../animations";
import { useI18n, type Lang } from "../../i18n";
import "./Navbar.css";

export default function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [langOpen, setLangOpen] = useState(false);
  const submitRef = useRef<HTMLAnchorElement>(null);
  const langBtnRef = useRef<HTMLDivElement>(null);
  const { lang, setLang, t } = useI18n();

  // 点击语言下拉外部时关闭
  useEffect(() => {
    if (!langOpen) return;
    const onDocClick = (e: MouseEvent) => {
      if (langBtnRef.current && !langBtnRef.current.contains(e.target as Node)) {
        setLangOpen(false);
      }
    };
    document.addEventListener("mousedown", onDocClick);
    return () => document.removeEventListener("mousedown", onDocClick);
  }, [langOpen]);

  const handleSubmitClick = () => {
    if (submitRef.current) buttonClick(submitRef.current);
  };

  const selectLang = (l: Lang) => {
    setLang(l);
    setLangOpen(false);
  };

  return (
    <nav className="navbar">
      <div className="navbar-inner">
        <Link to="/" className="navbar-logo">
          <svg
            width="32"
            height="32"
            viewBox="0 0 32 32"
            fill="none"
            className="logo-icon"
          >
            <rect width="32" height="32" rx="8" fill="var(--color-primary)" />
            <path
              d="M8 12h16M8 16h12M8 20h14"
              stroke="white"
              strokeWidth="2.5"
              strokeLinecap="round"
            />
          </svg>
          <span className="logo-text">Agent Skills</span>
        </Link>

        <div className="navbar-actions">
          <div className="lang-switch" ref={langBtnRef}>
            <button
              className="btn-lang"
              onClick={() => setLangOpen((o) => !o)}
              aria-haspopup="listbox"
              aria-expanded={langOpen}
              aria-label={t("nav.language")}
            >
              {lang === "zh" ? t("nav.zh") : t("nav.en")}
              <svg width="12" height="12" viewBox="0 0 12 12">
                <path d="M3 5l3 3 3-3" stroke="currentColor" strokeWidth="1.5" fill="none" />
              </svg>
            </button>
            <div className={`lang-menu${langOpen ? " open" : ""}`} role="listbox">
              <button
                className={`lang-option${lang === "zh" ? " active" : ""}`}
                onClick={() => selectLang("zh")}
                role="option"
                aria-selected={lang === "zh"}
              >
                简体中文
              </button>
              <button
                className={`lang-option${lang === "en" ? " active" : ""}`}
                onClick={() => selectLang("en")}
                role="option"
                aria-selected={lang === "en"}
              >
                English
              </button>
            </div>
          </div>
          <Link
            to="/submit"
            className="btn-submit"
            ref={submitRef}
            onMouseEnter={() => submitRef.current && buttonHoverEnter(submitRef.current)}
            onMouseLeave={() => submitRef.current && buttonHoverLeave(submitRef.current)}
            onClick={handleSubmitClick}
          >
            {t("nav.submit")}
          </Link>
          <button
            className="btn-menu"
            onClick={() => setMenuOpen(!menuOpen)}
            aria-label={t("nav.openMenu")}
            aria-expanded={menuOpen}
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
              <path
                d="M3 5h14M3 10h14M3 15h14"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
              />
            </svg>
          </button>
        </div>
      </div>

      <div className={`navbar-menu${menuOpen ? " open" : ""}`}>
        <Link to="/official" className="navbar-menu-link" onClick={() => setMenuOpen(false)}>
          {t("nav.official")}
        </Link>
        <Link to="/featured" className="navbar-menu-link" onClick={() => setMenuOpen(false)}>
          {t("nav.featured")}
        </Link>
        <Link to="/categories" className="navbar-menu-link" onClick={() => setMenuOpen(false)}>
          {t("nav.categories")}
        </Link>
        <Link to="/submit" className="navbar-menu-link" onClick={() => setMenuOpen(false)}>
          {t("nav.submitSkill")}
        </Link>
      </div>
    </nav>
  );
}
