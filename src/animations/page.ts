import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * 页面进入动画：子元素按层级顺序依次出现
 * 顺序：背景 → 子元素按 DOM 顺序 stagger
 */
export function pageEnter(container: Element | string) {
  const dur = getDuration(DURATION.slow);
  const dist = getDistance(12);

  const tl = gsap.timeline({ defaults: { ease: EASE.smooth } });

  // 容器先出现
  tl.fromTo(
    container,
    { opacity: 0 },
    { opacity: 1, duration: dur * 0.3 }
  );

  // 直接子元素按顺序入场
  const children = (container instanceof Element
    ? container
    : document.querySelector(container)
  )?.children;

  if (children && children.length > 0) {
    tl.fromTo(
      children,
      { opacity: 0, y: dist },
      {
        opacity: 1,
        y: 0,
        duration: dur * 0.5,
        stagger: dur * 0.1,
        ease: EASE.smooth,
      },
      `-=${dur * 0.2}`
    );
  }

  return tl;
}

/**
 * 页面退出动画
 */
export function pageLeave(container: Element | string) {
  const dur = getDuration(DURATION.normal);

  return gsap.to(container, {
    opacity: 0,
    y: getDistance(-8),
    duration: dur,
    ease: EASE.in,
  });
}

/**
 * Section 进入动画：带有 stagger 的元素列表
 */
export function sectionEnter(
  elements: Element[] | string,
  options?: { staggerAmount?: number; fromY?: number }
) {
  const dur = getDuration(DURATION.medium);
  const dist = getDistance(options?.fromY ?? 16);
  const stagger = options?.staggerAmount ?? 0.06;

  return gsap.fromTo(
    elements,
    { opacity: 0, y: dist },
    {
      opacity: 1,
      y: 0,
      duration: dur,
      stagger,
      ease: EASE.smooth,
    }
  );
}

/**
 * 面包屑入场
 */
export function breadcrumbEnter(element: Element | string) {
  return gsap.fromTo(
    element,
    { opacity: 0, y: getDistance(-4) },
    {
      opacity: 1,
      y: 0,
      duration: getDuration(DURATION.normal),
      ease: EASE.out,
    }
  );
}
