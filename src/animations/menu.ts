import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * 下拉菜单打开
 * opacity 0→1, y -4→0, scale 0.98→1
 */
export function menuEnter(menu: Element | string) {
  const dur = getDuration(DURATION.fast);

  return gsap.fromTo(
    menu,
    { opacity: 0, y: getDistance(-4), scale: 0.98 },
    {
      opacity: 1,
      y: 0,
      scale: 1,
      duration: dur,
      ease: EASE.smooth,
    }
  );
}

/**
 * 下拉菜单关闭
 */
export function menuLeave(menu: Element | string) {
  return gsap.to(menu, {
    opacity: 0,
    y: getDistance(-2),
    scale: 0.98,
    duration: getDuration(DURATION.fast * 0.7),
    ease: EASE.in,
  });
}

/**
 * 右键菜单 / Context Menu
 */
export function contextMenuEnter(menu: Element | string, x: number, y: number) {
  gsap.set(menu, { x, y, opacity: 0, scale: 0.98 });
  return gsap.to(menu, {
    opacity: 1,
    scale: 1,
    duration: getDuration(DURATION.fast),
    ease: EASE.smooth,
  });
}
