/**
 * 全局动效参数配置
 * 统一 duration / ease / distance，确保整体动效语言一致
 */

/** 动画时长 (秒) */
export const DURATION = {
  /** 按钮点击、hover 反馈 */
  micro: 0.15,
  /** 卡片 hover、下拉菜单 */
  quick: 0.2,
  /** 普通 UI 过渡 */
  normal: 0.3,
  /** 页面元素进入 */
  enter: 0.45,
  /** 页面整体进入 stagger 间隔 */
  stagger: 0.06,
} as const;

/** 缓动函数 — 统一使用 power 系列，禁止 bounce/elastic/linear */
export const EASE = {
  /** 进入动画默认 */
  enter: "power2.out",
  /** 退出动画 */
  exit: "power2.in",
  /** 强调动画 */
  emphasis: "power3.out",
} as const;

/** 位移距离 (px) — 克制的小幅位移 */
export const DISTANCE = {
  /** 按钮 hover 微上浮 */
  hover: 1,
  /** 卡片 hover 微上浮 */
  cardHover: 2,
  /** 页面元素进入 (Y) */
  enterY: 10,
  /** 页面元素进入 (X) */
  enterX: 12,
} as const;

/** 缩放值 — 极小幅度 */
export const SCALE = {
  /** 元素进入起始 */
  enterFrom: 0.98,
  /** 按钮按下 */
  press: 0.98,
  /** hover 轻微放大 */
  hover: 1.01,
} as const;
