import gsap from "gsap";
import { getDuration, EASE, DURATION } from "./utils";

/**
 * 进度条填充动画
 * from 0 → targetPercent (0-100)
 */
export function progressFill(
  bar: Element | string,
  targetPercent: number,
  duration?: number
) {
  const dur = getDuration(duration ?? DURATION.medium);

  return gsap.fromTo(
    bar,
    { width: "0%" },
    {
      width: `${targetPercent}%`,
      duration: dur,
      ease: EASE.smooth,
    }
  );
}

/**
 * 进度条更新动画（从当前值到新值）
 */
export function progressUpdate(
  bar: Element | string,
  targetPercent: number
) {
  return gsap.to(bar, {
    width: `${targetPercent}%`,
    duration: getDuration(DURATION.normal),
    ease: EASE.out,
  });
}

/**
 * 数字滚动动画
 * elem: 显示数字的 DOM 元素
 * from: 起始数字
 * to: 目标数字
 * options.decimals: 小数位数
 * options.prefix/suffix: 前后缀
 */
export function numberRoll(
  elem: Element | string,
  from: number,
  to: number,
  options?: {
    decimals?: number;
    prefix?: string;
    suffix?: string;
    duration?: number;
  }
) {
  const decimals = options?.decimals ?? 0;
  const prefix = options?.prefix ?? "";
  const suffix = options?.suffix ?? "";
  const dur = getDuration(options?.duration ?? DURATION.medium);

  const obj = { value: from };

  return gsap.to(obj, {
    value: to,
    duration: dur,
    ease: EASE.out,
    onUpdate: () => {
      const el =
        typeof elem === "string" ? document.querySelector(elem) : elem;
      if (el) {
        el.textContent = `${prefix}${obj.value.toFixed(decimals)}${suffix}`;
      }
    },
  });
}

/**
 * 状态指示器颜色过渡
 */
export function statusTransition(
  indicator: Element | string,
  fromColor: string,
  toColor: string
) {
  return gsap.fromTo(
    indicator,
    { backgroundColor: fromColor },
    {
      backgroundColor: toColor,
      duration: getDuration(DURATION.normal),
      ease: EASE.out,
    }
  );
}
