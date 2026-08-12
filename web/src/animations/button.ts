import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * 按钮 hover：轻微上浮
 */
export function buttonHoverEnter(element: Element | string) {
  return gsap.to(element, {
    y: getDistance(-1),
    duration: getDuration(DURATION.fast),
    ease: EASE.out,
  });
}

/**
 * 按钮 hover 离开
 */
export function buttonHoverLeave(element: Element | string) {
  return gsap.to(element, {
    y: 0,
    duration: getDuration(DURATION.fast),
    ease: EASE.out,
  });
}

/**
 * 按钮点击反馈：按下 → 弹回
 */
export function buttonClick(element: Element | string) {
  const dur = getDuration(0.08);
  const tl = gsap.timeline();

  tl.to(element, {
    scale: 0.97,
    duration: dur,
    ease: EASE.in,
  }).to(element, {
    scale: 1,
    duration: dur * 1.5,
    ease: EASE.out,
  });

  return tl;
}
