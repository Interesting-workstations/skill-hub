import gsap from "gsap";
import { DURATION, EASE, DISTANCE, SCALE } from "./config";

/**
 * 按钮 hover 进入 — 微上浮
 */
export function buttonHoverEnter(el: Element) {
  gsap.to(el, {
    y: -DISTANCE.hover,
    scale: SCALE.hover,
    duration: DURATION.micro,
    ease: EASE.enter,
    overwrite: "auto",
  });
}

/**
 * 按钮 hover 离开 — 复位
 */
export function buttonHoverLeave(el: Element) {
  gsap.to(el, {
    y: 0,
    scale: 1,
    duration: DURATION.micro,
    ease: EASE.exit,
    overwrite: "auto",
  });
}

/**
 * 按钮按下
 */
export function buttonPress(el: Element) {
  gsap.to(el, {
    scale: SCALE.press,
    duration: DURATION.micro / 2,
    ease: EASE.exit,
    overwrite: "auto",
    onComplete: () => {
      gsap.to(el, {
        scale: 1,
        duration: DURATION.micro,
        ease: EASE.enter,
        overwrite: "auto",
      });
    },
  });
}

/**
 * 卡片 hover 进入 — 微上浮 + 阴影
 */
export function cardHoverEnter(el: Element) {
  gsap.to(el, {
    y: -DISTANCE.cardHover,
    duration: DURATION.quick,
    ease: EASE.enter,
    overwrite: "auto",
  });
}

/**
 * 卡片 hover 离开 — 复位
 */
export function cardHoverLeave(el: Element) {
  gsap.to(el, {
    y: 0,
    duration: DURATION.quick,
    ease: EASE.exit,
    overwrite: "auto",
  });
}
