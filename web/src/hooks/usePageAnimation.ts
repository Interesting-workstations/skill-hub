import { useEffect, useRef } from "react";
import { pageEnter, createAnimationContext } from "../animations";

export type AnimationContext = ReturnType<typeof createAnimationContext>;

/**
 * 页面入场动画 Hook：
 * - 页面挂载 / deps 变化时播放 pageEnter 入场动画
 * - 统一管理动画上下文，卸载时清理全部动画
 * - onEnter 回调用于追加额外动画（如 sectionEnter、panelEnterRight）
 */
export function usePageAnimation(
  onEnter?: (container: HTMLElement, ctx: AnimationContext) => void,
  deps: readonly unknown[] = []
) {
  const pageRef = useRef<HTMLDivElement>(null);
  const ctxRef = useRef<AnimationContext>(createAnimationContext());

  useEffect(() => {
    const container = pageRef.current;
    if (!container) return;
    const ctx = ctxRef.current;
    ctx.killAll();
    ctx.add(pageEnter(container));
    onEnter?.(container, ctx);
    return () => {
      ctx.killAll();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return pageRef;
}
