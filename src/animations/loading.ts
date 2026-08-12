import gsap from "gsap";
import { getDuration, EASE, DURATION } from "./utils";

/**
 * 简单旋转 Loading（适用于小图标）
 */
export function loadingSpinner(element: Element | string) {
  return gsap.to(element, {
    rotation: 360,
    duration: 0.8,
    repeat: -1,
    ease: "none",
  });
}

/**
 * 骨架屏闪烁 Loading
 */
export function loadingSkeleton(element: Element | string) {
  return gsap.fromTo(
    element,
    { opacity: 0.3 },
    {
      opacity: 0.7,
      duration: 0.8,
      repeat: -1,
      yoyo: true,
      ease: EASE.inOut,
    }
  );
}

/**
 * 进度条动画：0 → 目标值
 */
export function loadingProgress(
  bar: Element | string,
  targetPercent: number = 100
) {
  const dur = getDuration(DURATION.slow * 2); // 进度条稍慢一点

  return gsap.fromTo(
    bar,
    { width: "0%", opacity: 0.3 },
    {
      width: `${targetPercent}%`,
      opacity: 1,
      duration: dur,
      ease: EASE.inOut,
    }
  );
}

/**
 * 停止 Loading 并淡出
 */
export function loadingComplete(
  element: Element | string,
  onComplete?: () => void
) {
  return gsap.to(element, {
    opacity: 0,
    duration: getDuration(DURATION.fast),
    ease: EASE.in,
    onComplete: onComplete || undefined,
  });
}

/**
 * Dot 脉冲动画（三点加载）
 */
export function loadingDots(container: Element | string) {
  const dots = (
    typeof container === "string"
      ? document.querySelector(container)
      : container
  )?.querySelectorAll(".loading-dot");

  if (!dots || dots.length === 0) return null;

  return gsap.fromTo(
    dots,
    { scale: 0.6, opacity: 0.3 },
    {
      scale: 1,
      opacity: 1,
      duration: 0.4,
      stagger: 0.15,
      repeat: -1,
      yoyo: true,
      ease: EASE.inOut,
    }
  );
}

/**
 * 页面加载进度条（顶部细线）
 */
export function pageLoadingBar(bar: Element | string) {
  const dur = getDuration(0.6);

  const tl = gsap.timeline();

  tl.fromTo(
    bar,
    { scaleX: 0, transformOrigin: "left center" },
    { scaleX: 0.3, duration: dur * 0.3, ease: EASE.out }
  )
    .to(bar, { scaleX: 0.7, duration: dur * 0.4, ease: "none" })
    .to(bar, { scaleX: 0.9, duration: dur * 0.2, ease: "none" })
    .to(bar, { scaleX: 1, duration: dur * 0.1, ease: EASE.out })
    .to(bar, { opacity: 0, duration: getDuration(DURATION.fast) });

  return tl;
}
