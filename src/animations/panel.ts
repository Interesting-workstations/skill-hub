import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * 侧边 Panel 打开（从右侧滑入）
 * x: 20→0, opacity 0→1
 */
export function panelEnterRight(panel: Element | string) {
  return gsap.fromTo(
    panel,
    { x: getDistance(20), opacity: 0 },
    {
      x: 0,
      opacity: 1,
      duration: getDuration(DURATION.medium),
      ease: EASE.smooth,
    }
  );
}

/**
 * 侧边 Panel 关闭（向右滑出）
 */
export function panelLeaveRight(panel: Element | string) {
  return gsap.to(panel, {
    x: getDistance(20),
    opacity: 0,
    duration: getDuration(DURATION.normal),
    ease: EASE.in,
  });
}

/**
 * 侧边 Panel 打开（从左侧滑入）
 */
export function panelEnterLeft(panel: Element | string) {
  return gsap.fromTo(
    panel,
    { x: getDistance(-20), opacity: 0 },
    {
      x: 0,
      opacity: 1,
      duration: getDuration(DURATION.medium),
      ease: EASE.smooth,
    }
  );
}

/**
 * 侧边 Panel 关闭（向左滑出）
 */
export function panelLeaveLeft(panel: Element | string) {
  return gsap.to(panel, {
    x: getDistance(-20),
    opacity: 0,
    duration: getDuration(DURATION.normal),
    ease: EASE.in,
  });
}

/**
 * 浮动 Panel 打开（下方弹出）
 */
export function panelEnterFloat(panel: Element | string) {
  return gsap.fromTo(
    panel,
    { y: getDistance(12), opacity: 0, scale: 0.98 },
    {
      y: 0,
      opacity: 1,
      scale: 1,
      duration: getDuration(DURATION.normal),
      ease: EASE.smooth,
    }
  );
}

/**
 * 内容区 Panel 展开/收起（高度变化 + 内容淡入）
 */
export function panelToggle(
  wrapper: Element | string,
  content: Element | string,
  open: boolean
) {
  const dur = getDuration(DURATION.medium);

  if (open) {
    return gsap
      .timeline({ defaults: { ease: EASE.smooth } })
      .to(wrapper, { height: "auto", duration: dur })
      .fromTo(
        content,
        { opacity: 0, y: getDistance(8) },
        { opacity: 1, y: 0, duration: dur * 0.6 },
        `-=${dur * 0.5}`
      );
  } else {
    return gsap
      .timeline({ defaults: { ease: EASE.in } })
      .to(content, { opacity: 0, y: getDistance(4), duration: dur * 0.4 })
      .to(wrapper, { height: 0, duration: dur * 0.7 }, `-=${dur * 0.2}`);
  }
}
