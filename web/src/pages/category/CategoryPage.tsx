import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import SkillCard from "../../components/skill/SkillCard";
import { sectionEnter } from "../../animations";
import { usePageAnimation } from "../../hooks/usePageAnimation";
import { usePageMeta } from "../../hooks/usePageMeta";
import { useCategory, useSkills } from "../../hooks/useSkillData";
import PageContainer from "../../components/shared/PageContainer";
import Breadcrumb from "../../components/shared/Breadcrumb";
import PageLoading from "../../components/shared/PageLoading";
import { useI18n, categoryName } from "../../i18n";
import type { Skill } from "../../data/types";

/** 分类页每页加载数量（分页拉取 + 触底自动加载，避免一次性渲染全部技能导致卡顿） */
const PAGE_SIZE = 24;

export default function CategoryPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data: category, loading } = useCategory(slug);
  // 分页加载分类技能（累积）
  const [items, setItems] = useState<Skill[]>([]);
  const [offset, setOffset] = useState(0);
  const [hasMore, setHasMore] = useState(false);
  const { data: page, loading: pageLoading } = useSkills(
    slug ? { category: slug, limit: PAGE_SIZE, offset } : null
  );
  const { t } = useI18n();
  // 触底哨兵节点：进入视口时自动加载下一页
  const sentinelRef = useRef<HTMLDivElement>(null);
  // 首屏卡片入场动画只播一次（加载更多时不动画，避免整页闪烁/空白）
  const animatedRef = useRef(false);

  // 切换分类时重置分页状态
  useEffect(() => {
    setItems([]);
    setOffset(0);
    setHasMore(false);
    animatedRef.current = false;
  }, [slug]);

  // 累积分页结果；返回不足一页说明已到底
  useEffect(() => {
    if (page) {
      setItems((prev) => (offset === 0 ? page : [...prev, ...page]));
      setHasMore(page.length >= PAGE_SIZE);
    }
  }, [page]);

  // 首次数据就绪后对新卡片播放入场动画（仅一次；后续加载更多直接显示）
  useEffect(() => {
    if (items.length > 0 && !animatedRef.current) {
      animatedRef.current = true;
      const cards = pageRef.current?.querySelectorAll(".skill-card") ?? [];
      if (cards.length > 0) {
        const raf = requestAnimationFrame(() => {
          sectionEnter(cards, { fromY: 12 });
        });
        return () => cancelAnimationFrame(raf);
      }
    }
  }, [items.length]);

  // 触底自动加载：哨兵进入视口（提前 240px）且仍有更多、未在加载时，翻下一页
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el) return;
    const io = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !pageLoading) {
          setOffset((o) => o + PAGE_SIZE);
        }
      },
      { rootMargin: "240px 0px" }
    );
    io.observe(el);
    return () => io.disconnect();
  }, [hasMore, pageLoading]);

  // 兜底显示名（数据未就绪时用 slug 转换）
  const fallbackName = slug
    ? slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())
    : "";
  const displayName = category
    ? categoryName(t, category.slug, category.name)
    : fallbackName;
  const totalCount = category?.count ?? items.length;

  // 入场动画只在切换分类时触发（不再依赖 items.length，避免加载更多时整页重播动画闪烁）
  const pageRef = usePageAnimation((container, ctx) => {
    const cards = container.querySelectorAll(".skill-card");
    if (cards.length > 0) {
      ctx.add(sectionEnter(cards, { fromY: 12 }));
    }
  }, [slug]);

  usePageMeta({
    title: displayName ? `${displayName} — ${t("brand.name")}` : t("brand.title"),
  });

  if (loading && !items.length) {
    return <PageLoading />;
  }

  return (
    <PageContainer ref={pageRef}>
      <Breadcrumb
        items={[
          { label: t("breadcrumb.home"), to: "/" },
          { label: t("categories.title"), to: "/categories" },
          { label: displayName },
        ]}
      />

      <h1 style={{ fontSize: 28, fontWeight: 800, color: "var(--color-text)", margin: "0 0 8px" }}>{displayName}</h1>
      <p style={{ fontSize: 15, color: "var(--color-text-secondary)", margin: "0 0 28px" }}>
        {t("category.skillCount", { n: totalCount })}
      </p>

      {items.length > 0 ? (
        <>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
            {items.map((s) => (
              <SkillCard key={s.id} skill={s} />
            ))}
          </div>
          {/* 触底自动加载哨兵 + 底部状态提示 */}
          <div ref={sentinelRef} style={{ textAlign: "center", marginTop: 28, minHeight: 48 }}>
            {pageLoading ? (
              <span style={{ color: "var(--color-text-secondary)", fontSize: 14 }}>{t("category.loading")}</span>
            ) : hasMore ? (
              <span style={{ color: "var(--color-text-muted)", fontSize: 14 }}>…</span>
            ) : (
              <span style={{ color: "var(--color-text-muted)", fontSize: 13 }}>{t("category.end")}</span>
            )}
          </div>
        </>
      ) : (
        <p style={{ color: "var(--color-text-muted)", fontSize: 14 }}>{t("category.empty")}</p>
      )}
    </PageContainer>
  );
}
