import { useEffect, useRef } from "react";
import gsap from "gsap";
import { useReducedMotion } from "./reducedMotion";
import { DURATION, EASE, DISTANCE, SCALE } from "./config";

/**
 * 模块级 Set：记录当前会话中已完成入场动画的页面。
 * 浏览器刷新后重置，SPA 内路由切换不重复播放。
 */
const animatedPages = new Set<string>();

/**
 * 页面进入动画 Hook
 * 首次进入时播放 GSAP Timeline 分层入场；
 * SPA 路由回退时直接显示（不重复播放），避免"空白等待"。
 *
 * @param pageKey  页面唯一标识，同 key 同会话只播一次
 */
export function usePageEnter(pageKey = "default") {
  const reduced = useReducedMotion();
  const ctxRef = useRef<gsap.Context | null>(null);
  const shouldAnimate = !animatedPages.has(pageKey);

  useEffect(() => {
    if (!shouldAnimate) return;

    // 标记已播放，后续同页面路由切换直接跳过
    animatedPages.add(pageKey);

    const ctx = gsap.context(() => {
      if (reduced) {
        gsap.set("[data-animate]", { opacity: 1, clearProps: "transform" });
        return;
      }

      const tl = gsap.timeline({ defaults: { ease: EASE.enter } });

      function safeFromTo(
        selector: string,
        fromVars: gsap.TweenVars,
        toVars: gsap.TweenVars,
        position?: gsap.Position
      ) {
        if (document.querySelectorAll(selector).length > 0) {
          tl.fromTo(selector, fromVars, toVars, position);
        }
      }

      // 层级 1: 面包屑 / 区域标题
      safeFromTo(
        "[data-animate='breadcrumb'], [data-animate='section-title']",
        { opacity: 0, y: DISTANCE.enterY },
        { opacity: 1, y: 0, duration: DURATION.enter, stagger: DURATION.stagger },
        0
      );

      // 层级 2: Hero 内容
      safeFromTo(
        "[data-animate='hero-content'] > *",
        { opacity: 0, y: DISTANCE.enterY },
        { opacity: 1, y: 0, duration: DURATION.enter, stagger: DURATION.stagger * 1.5 },
        0.05
      );

      // 层级 3: 卡片网格
      safeFromTo(
        "[data-animate='card']",
        { opacity: 0, y: DISTANCE.enterY, scale: SCALE.enterFrom },
        { opacity: 1, y: 0, scale: 1, duration: DURATION.enter, stagger: { each: DURATION.stagger * 1.2, from: "start" } },
        0.1
      );

      // 层级 4: 详情页内容区
      safeFromTo(
        "[data-animate='detail-content']",
        { opacity: 0, y: DISTANCE.enterY },
        { opacity: 1, y: 0, duration: DURATION.enter },
        0.08
      );

      // 层级 5: 侧边栏
      safeFromTo(
        "[data-animate='sidebar']",
        { opacity: 0, x: DISTANCE.enterX },
        { opacity: 1, x: 0, duration: DURATION.enter },
        0.15
      );

      // 层级 6: 详情 section
      safeFromTo(
        "[data-animate='detail-section']",
        { opacity: 0, y: DISTANCE.enterY },
        { opacity: 1, y: 0, duration: DURATION.enter, stagger: DURATION.stagger * 1.5 },
        0.2
      );
    });

    ctxRef.current = ctx;

    return () => {
      ctx.revert();
    };
  }, [shouldAnimate, pageKey, reduced]);

  return ctxRef;
}
