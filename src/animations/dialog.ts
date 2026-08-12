import gsap from "gsap";
import { getDuration, getDistance, EASE, DURATION } from "./utils";

/**
 * Dialog / Modal 打开动画
 * Overlay: opacity 0→1
 * Dialog: opacity 0→1, scale 0.97→1, y 8→0
 */
export function dialogEnter(
  overlay: Element | string,
  dialog: Element | string
) {
  const dur = getDuration(DURATION.normal);
  const dist = getDistance(8);

  const tl = gsap.timeline({ defaults: { ease: EASE.smooth } });

  tl.fromTo(
    overlay,
    { opacity: 0 },
    { opacity: 1, duration: dur * 0.6 }
  ).fromTo(
    dialog,
    { opacity: 0, scale: 0.97, y: dist },
    { opacity: 1, scale: 1, y: 0, duration: dur },
    `-=${dur * 0.3}`
  );

  return tl;
}

/**
 * Dialog / Modal 关闭动画
 * Dialog: opacity 1→0, scale 1→0.98
 * Overlay: opacity 1→0
 */
export function dialogLeave(
  overlay: Element | string,
  dialog: Element | string
) {
  const dur = getDuration(DURATION.fast);

  const tl = gsap.timeline({ defaults: { ease: EASE.in } });

  tl.to(dialog, {
    opacity: 0,
    scale: 0.98,
    y: getDistance(4),
    duration: dur,
  }).to(
    overlay,
    { opacity: 0, duration: dur * 0.8 },
    `-=${dur * 0.5}`
  );

  return tl;
}

/**
 * 轻量 Dialog（Alert / Confirm）
 */
export function lightDialogEnter(dialog: Element | string) {
  return gsap.fromTo(
    dialog,
    { opacity: 0, scale: 0.98, y: getDistance(6) },
    {
      opacity: 1,
      scale: 1,
      y: 0,
      duration: getDuration(DURATION.fast),
      ease: EASE.smooth,
    }
  );
}

export function lightDialogLeave(dialog: Element | string) {
  return gsap.to(dialog, {
    opacity: 0,
    scale: 0.98,
    duration: getDuration(DURATION.fast),
    ease: EASE.in,
  });
}
