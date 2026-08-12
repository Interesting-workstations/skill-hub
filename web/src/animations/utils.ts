/** 检测用户是否开启了减少动画偏好 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/** 统一的动画配置常量 */
export const DURATION = {
  fast: 0.15,
  normal: 0.25,
  medium: 0.35,
  slow: 0.5,
} as const;

export const EASE = {
  out: "power2.out",
  in: "power2.in",
  inOut: "power2.inOut",
  smooth: "power3.out",
} as const;

/** 获取经过 reduced-motion 调整的 duration */
export function getDuration(base: number): number {
  return prefersReducedMotion() ? 0 : base;
}

/** 获取经过 reduced-motion 调整的位移 */
export function getDistance(base: number): number {
  return prefersReducedMotion() ? 0 : base;
}

/** 创建一个干净的动画上下文，便于统一清理 */
export function createAnimationContext() {
  const tweens: gsap.core.Tween[] = [];
  const timelines: gsap.core.Timeline[] = [];

  return {
    add(t: gsap.core.Tween | gsap.core.Timeline) {
      if ("add" in t) {
        timelines.push(t as gsap.core.Timeline);
      } else {
        tweens.push(t as gsap.core.Tween);
      }
      return t;
    },
    killAll() {
      tweens.forEach((t) => t.kill());
      timelines.forEach((tl) => tl.kill());
      tweens.length = 0;
      timelines.length = 0;
    },
  };
}
