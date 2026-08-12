import { useRef, useEffect, useCallback } from "react";
import gsap from "gsap";
import { useReducedMotion } from "./reducedMotion";
import { DURATION, EASE, DISTANCE, SCALE } from "./config";

/**
 * 按钮微交互 Hook — hover 微上浮 + press 微缩
 * 返回 callback ref，兼容任意 HTML 元素（button / a / div）
 */
export function useButtonMicro() {
  const reduced = useReducedMotion();
  const elRef = useRef<HTMLElement | null>(null);

  const ref = useCallback((node: HTMLElement | null) => {
    elRef.current = node;
  }, []);

  useEffect(() => {
    const el = elRef.current;
    if (!el || reduced) return;

    const onEnter = () => {
      gsap.to(el, {
        y: -DISTANCE.hover,
        scale: SCALE.hover,
        duration: DURATION.micro,
        ease: EASE.enter,
        overwrite: "auto",
      });
    };
    const onLeave = () => {
      gsap.to(el, {
        y: 0,
        scale: 1,
        duration: DURATION.micro,
        ease: EASE.exit,
        overwrite: "auto",
      });
    };
    const onDown = () => {
      gsap.to(el, {
        scale: SCALE.press,
        duration: DURATION.micro / 2,
        ease: EASE.exit,
        overwrite: "auto",
      });
    };
    const onUp = () => {
      gsap.to(el, {
        scale: SCALE.hover,
        duration: DURATION.micro,
        ease: EASE.enter,
        overwrite: "auto",
      });
    };

    el.addEventListener("mouseenter", onEnter);
    el.addEventListener("mouseleave", onLeave);
    el.addEventListener("mousedown", onDown);
    el.addEventListener("mouseup", onUp);

    return () => {
      el.removeEventListener("mouseenter", onEnter);
      el.removeEventListener("mouseleave", onLeave);
      el.removeEventListener("mousedown", onDown);
      el.removeEventListener("mouseup", onUp);
    };
  }, [reduced]);

  return ref;
}

/**
 * 卡片 hover Hook — 微上浮
 */
export function useCardHover() {
  const reduced = useReducedMotion();
  const elRef = useRef<HTMLElement | null>(null);

  const ref = useCallback((node: HTMLElement | null) => {
    elRef.current = node;
  }, []);

  useEffect(() => {
    const el = elRef.current;
    if (!el || reduced) return;

    const onEnter = () => {
      gsap.to(el, {
        y: -DISTANCE.cardHover,
        duration: DURATION.quick,
        ease: EASE.enter,
        overwrite: "auto",
      });
    };
    const onLeave = () => {
      gsap.to(el, {
        y: 0,
        duration: DURATION.quick,
        ease: EASE.exit,
        overwrite: "auto",
      });
    };

    el.addEventListener("mouseenter", onEnter);
    el.addEventListener("mouseleave", onLeave);

    return () => {
      el.removeEventListener("mouseenter", onEnter);
      el.removeEventListener("mouseleave", onLeave);
    };
  }, [reduced]);

  return ref;
}

/**
 * 复制按钮反馈 — 短暂闪烁
 */
export function useCopyFeedback() {
  const reduced = useReducedMotion();
  const elRef = useRef<HTMLElement | null>(null);

  const ref = useCallback((node: HTMLElement | null) => {
    elRef.current = node;
  }, []);

  const flash = useCallback(() => {
    if (reduced) return;
    const el = elRef.current;
    if (!el) return;
    gsap.fromTo(
      el,
      { backgroundColor: "#d9f99d" },
      {
        backgroundColor: "transparent",
        duration: 0.6,
        ease: EASE.exit,
        overwrite: "auto",
      }
    );
  }, [reduced]);

  return { ref, flash };
}
