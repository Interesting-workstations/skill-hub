import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * 卡片 hover 进入：轻微上浮 + 阴影增强
 */
export function cardHoverEnter(element: Element | string) {
  return gsap.to(element, {
    y: getDistance(-3),
    boxShadow: "0 8px 30px rgba(124, 58, 237, 0.12)",
    borderColor: "rgba(124, 58, 237, 0.5)",
    duration: getDuration(DURATION.fast),
    ease: EASE.out,
  });
}

/**
 * 卡片 hover 离开：恢复原位
 */
export function cardHoverLeave(element: Element | string) {
  return gsap.to(element, {
    y: 0,
    boxShadow: "0 0 0 rgba(0,0,0,0)",
    borderColor: "#e5e7eb",
    duration: getDuration(DURATION.fast),
    ease: EASE.out,
  });
}
