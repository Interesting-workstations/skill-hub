// ── 页面动画 ──
export { pageEnter, pageLeave, sectionEnter, breadcrumbEnter } from "./page";

// ── 卡片动画 ──
export { cardHoverEnter, cardHoverLeave, cardListEnter } from "./card";

// ── 按钮动画 ──
export { buttonHoverEnter, buttonHoverLeave, buttonClick } from "./button";

// ── 弹窗动画 ──
export {
  dialogEnter,
  dialogLeave,
  lightDialogEnter,
  lightDialogLeave,
} from "./dialog";

// ── 菜单动画 ──
export { menuEnter, menuLeave, contextMenuEnter } from "./menu";

// ── 面板动画 ──
export {
  panelEnterRight,
  panelLeaveRight,
  panelEnterLeft,
  panelLeaveLeft,
  panelEnterFloat,
  panelToggle,
} from "./panel";

// ── 加载动画 ──
export {
  loadingSpinner,
  loadingSkeleton,
  loadingProgress,
  loadingComplete,
  loadingDots,
  pageLoadingBar,
} from "./loading";

// ── 进度与数字动画 ──
export {
  progressFill,
  progressUpdate,
  numberRoll,
  statusTransition,
} from "./progress";

// ── 工具函数 ──
export {
  prefersReducedMotion,
  createAnimationContext,
  getDuration,
  getDistance,
  DURATION,
  EASE,
} from "./utils";
