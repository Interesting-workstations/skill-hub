import { useState } from "react";
import "./Navbar.css";

export default function Navbar() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <nav className="navbar">
      <div className="navbar-inner">
        <a href="/" className="navbar-logo">
          <svg
            width="32"
            height="32"
            viewBox="0 0 32 32"
            fill="none"
            className="logo-icon"
          >
            <rect width="32" height="32" rx="8" fill="#7c3aed" />
            <path
              d="M8 12h16M8 16h12M8 20h14"
              stroke="white"
              strokeWidth="2.5"
              strokeLinecap="round"
            />
          </svg>
          <span className="logo-text">Agent Skills</span>
        </a>

        <div className="navbar-actions">
          <button className="btn-lang">
            简体中文
            <svg width="12" height="12" viewBox="0 0 12 12">
              <path d="M3 5l3 3 3-3" stroke="currentColor" strokeWidth="1.5" fill="none" />
            </svg>
          </button>
          <a href="/submit" className="btn-submit">
            提交
          </a>
          <button
            className="btn-menu"
            onClick={() => setMenuOpen(!menuOpen)}
            aria-label="Open menu"
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
    </nav>
  );
}
