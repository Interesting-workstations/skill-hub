import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchSkills } from "../../services/api/skills";
import type { Skill } from "../../data/types";
import { useI18n } from "../../i18n";
import "./GlobalSearch.css";

const PAGE_SIZE = 10;
const DEBOUNCE_MS = 250;

/** 全局搜索：导航栏搜索框 + 实时下拉结果（分页 + 滚动加载更多），支持键盘导航。 */
export default function GlobalSearch() {
  const { t, lang } = useI18n();
  const [value, setValue] = useState("");
  const [results, setResults] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);
  const boxRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  // 分页滚动加载：当前已加载偏移量 + 请求序号（防止旧响应覆盖新结果）
  const offsetRef = useRef(0);
  const seqRef = useRef(0);

  // 点击外部 / Esc 关闭
  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (boxRef.current && !boxRef.current.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  // 全局 "/" 快捷键聚焦搜索框（GitHub 风格），输入框内除外
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== "/") return;
      const el = e.target as HTMLElement | null;
      if (el && (el.tagName === "INPUT" || el.tagName === "TEXTAREA" || el.isContentEditable)) return;
      e.preventDefault();
      inputRef.current?.focus();
      setOpen(true);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  // 防抖搜索（仅首屏，重置分页）；滚动到下拉底部时通过 loadMore 加载后续页
  useEffect(() => {
    const q = value.trim();
    if (!q) {
      setResults([]);
      setLoading(false);
      setHasMore(false);
      offsetRef.current = 0;
      setActive(-1);
      return;
    }
    setLoading(true);
    offsetRef.current = 0;
    const seq = ++seqRef.current;
    const id = window.setTimeout(() => {
      fetchSkills({ q, limit: PAGE_SIZE, offset: 0 })
        .then((list) => {
          if (seqRef.current !== seq) return;
          setResults(list);
          setHasMore(list.length >= PAGE_SIZE);
          setLoading(false);
        })
        .catch(() => {
          if (seqRef.current !== seq) return;
          setResults([]);
          setLoading(false);
        });
    }, DEBOUNCE_MS);
    return () => window.clearTimeout(id);
  }, [value]);

  // 加载下一页并追加到结果（滚动到底触发）
  const loadMore = () => {
    const q = value.trim();
    if (!q || loading || !hasMore) return;
    setLoading(true);
    const seq = seqRef.current;
    const nextOffset = offsetRef.current + PAGE_SIZE;
    offsetRef.current = nextOffset;
    fetchSkills({ q, limit: PAGE_SIZE, offset: nextOffset })
      .then((list) => {
        if (seqRef.current !== seq) return;
        setResults((prev) => [...prev, ...list]);
        setHasMore(list.length >= PAGE_SIZE);
        setLoading(false);
      })
      .catch(() => {
        if (seqRef.current === seq) setLoading(false);
      });
  };

  // 下拉列表滚动到底部（提前 60px）时自动加载更多
  const onPanelScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const el = e.currentTarget;
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 60) {
      loadMore();
    }
  };

  const close = () => {
    setOpen(false);
    setValue("");
    setResults([]);
    setActive(-1);
  };

  const go = (skill: Skill) => {
    close();
    navigate(`/skill/${skill.id}`);
  };

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setOpen(true);
      setActive((i) => Math.min(i + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      const target = active >= 0 ? results[active] : results[0];
      if (target) {
        e.preventDefault();
        go(target);
      }
    }
  };

  const showPanel = open && value.trim() !== "";

  return (
    <div className="global-search" ref={boxRef}>
      <div className={`global-search-box${open ? " focused" : ""}`}>
        <svg
          className="global-search-icon"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          aria-hidden="true"
        >
          <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2" />
          <path d="m20 20-3.5-3.5" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
        </svg>
        <input
          ref={inputRef}
          type="text"
          value={value}
          onChange={(e) => {
            setValue(e.target.value);
            setOpen(true);
            setActive(-1);
          }}
          onFocus={() => setOpen(true)}
          onKeyDown={onKeyDown}
          placeholder={t("nav.searchPlaceholder")}
          aria-label={t("nav.search")}
          aria-expanded={showPanel}
          role="combobox"
        />
        {loading && <span className="global-search-spinner" aria-hidden="true" />}
        {!open && value === "" && (
          <kbd className="global-search-kbd" aria-hidden="true">
            /
          </kbd>
        )}
      </div>
      <div
        className={`global-search-panel${showPanel ? " open" : ""}`}
        role="listbox"
        aria-label={t("nav.search")}
        onScroll={onPanelScroll}
      >
        {loading && results.length === 0 ? (
          <div className="global-search-empty">{t("nav.searchLoading")}</div>
        ) : results.length === 0 ? (
          <div className="global-search-empty">{t("nav.searchNoResult")}</div>
        ) : (
          <>
            {results.map((s, i) => (
              <button
                key={s.id}
                type="button"
                role="option"
                aria-selected={i === active}
                className={`global-search-item${i === active ? " active" : ""}`}
                onMouseEnter={() => setActive(i)}
                onClick={() => go(s)}
              >
                <span className="global-search-item-icon">{s.isOfficial ? "⭐" : "📦"}</span>
                <span className="global-search-item-body">
                  <span className="global-search-item-name">
                    {lang === "zh" ? s.nameZh || s.name : s.name}
                  </span>
                  <span className="global-search-item-meta">
                    {s.author}
                    {s.category ? ` · ${s.category}` : ""}
                  </span>
                </span>
                {s.githubStars && (
                  <span className="global-search-item-stars" title="GitHub Stars">
                    ★ {s.githubStars}
                  </span>
                )}
              </button>
            ))}
            {/* 滚动加载状态：加载中 / 已到底 */}
            <div className="global-search-more" aria-live="polite">
              {loading ? (
                <span className="global-search-spinner" aria-hidden="true" />
              ) : hasMore ? (
                <span style={{ color: "var(--color-text-muted)", fontSize: 12 }}>…</span>
              ) : (
                <span style={{ color: "var(--color-text-muted)", fontSize: 12 }}>{t("category.end")}</span>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
