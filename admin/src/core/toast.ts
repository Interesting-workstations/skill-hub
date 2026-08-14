/** 全局 Toast 桥：供非 React 模块（如 http.ts）在出错时弹出全局提示。
 *  ToastProvider 挂载时注册处理器，卸载时清除。 */

interface GlobalToastHandler {
  error: (message: string) => void;
}

let handler: GlobalToastHandler | null = null;

export function setGlobalToastHandler(h: GlobalToastHandler | null): void {
  handler = h;
}

/** 上报全局错误（自动弹出 error Toast） */
export function reportGlobalError(message: string): void {
  handler?.error(message);
}
