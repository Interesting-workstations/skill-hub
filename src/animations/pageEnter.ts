import { useEffect, useRef } from "react";
import gsap from "gsap";
import { useReducedMotion } from "./reducedMotion";
import { DURATION, EASE, DISTANCE, SCALE } from "./config";

/**
 * 页面进入动画 Hook
 * 使用 GSAP Timeline 建立有序的进入层级
 *
 * @param enabled 是否启用动画（路由切换时用于控制）
 */
export function usePageEnter(enabled = true) {
  const reduced = useReducedMotion();
  const ctxRef = useRef<gsap.Context | null>(null);

  useEffect(() => {
    if (!enabled) return;

    const ctx = gsap.context(() => {
      if (reduced) {
        // Reduced motion: 只做极简 opacity
        gsap.set("[data-animate]", { opacity: 1, clearProps: "transform" });
        return;
      }

      const tl = gsap.timeline({ defaults: { ease: EASE.enter } });

      /** 安全添加动画：选择器无匹配元素时跳过 */
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
        {
          opacity: 1,
          y: 0,
          duration: DURATION.enter,
          stagger: DURATION.stagger,
        },
        0
      );

      // 层级 2: Hero 内容
      safeFromTo(
        "[data-animate='hero-content'] > *",
        { opacity: 0, y: DISTANCE.enterY },
        {
          opacity: 1,
          y: 0,
          duration: DURATION.enter,
          stagger: DURATION.stagger * 1.5,
        },
        reduced ? 0 : 0.05
      );

      // 层级 3: 卡片网格
      safeFromTo(
        "[data-animate='card']",
        { opacity: 0, y: DISTANCE.enterY, scale: SCALE.enterFrom },
        {
          opacity: 1,
          y: 0,
          scale: 1,
          duration: DURATION.enter,
          stagger: {
            each: DURATION.stagger * 1.2,
            from: "start",
          },
        },
        reduced ? 0 : 0.1
      );

      // 层级 4: 详情页内容区
      safeFromTo(
        "[data-animate='detail-content']",
        { opacity: 0, y: DISTANCE.enterY },
        {
          opacity: 1,
          y: 0,
          duration: DURATION.enter,
        },
        reduced ? 0 : 0.08
      );

      // 层级 5: 侧边栏
      safeFromTo(
        "[data-animate='sidebar']",
        { opacity: 0, x: DISTANCE.enterX },
        {
          opacity: 1,
          x: 0,
          duration: DURATION.enter,
        },
        reduced ? 0 : 0.15
      );

      // 层级 6: 详情 section
      safeFromTo(
        "[data-animate='detail-section']",
        { opacity: 0, y: DISTANCE.enterY },
        {
          opacity: 1,
          y: 0,
          duration: DURATION.enter,
          stagger: DURATION.stagger * 1.5,
        },
        reduced ? 0 : 0.2
      );
    });

    ctxRef.current = ctx;

    return () => {
      ctx.revert();
    };
  }, [enabled, reduced]);

  return ctxRef;
}
