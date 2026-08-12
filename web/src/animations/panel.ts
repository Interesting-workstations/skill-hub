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
