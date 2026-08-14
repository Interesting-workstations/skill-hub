import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { setGlobalToastHandler } from "../core/toast";
import "./Toast.css";

export type ToastType = "success" | "error" | "info" | "warning";

interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
  leaving?: boolean;
}

export interface ToastApi {
  success: (message: string) => void;
  error: (message: string) => void;
  info: (message: string) => void;
  warning: (message: string) => void;
}

const ToastContext = createContext<ToastApi | null>(null);

const DURATION: Record<ToastType, number> = {
  success: 3000,
  info: 3000,
  warning: 4500,
  error: 5000,
};

const ICONS: Record<ToastType, ReactNode> = {
  success: (
    <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true">
      <circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15" />
      <path
        d="M5 8.2 7.2 10.4 11 6"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  ),
  error: (
    <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true">
      <circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15" />
      <path
        d="m5.5 5.5 5 5M10.5 5.5l-5 5"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
      />
    </svg>
  ),
  warning: (
    <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true">
      <circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15" />
      <path d="M8 4.8v4.2M8 11.4v.4" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  ),
  info: (
    <svg viewBox="0 0 16 16" width="16" height="16" fill="none" aria-hidden="true">
      <circle cx="8" cy="8" r="7" fill="currentColor" opacity="0.15" />
      <path d="M8 7.4v4.2M8 4.6v.4" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
    </svg>
  ),
};

let nextId = 1;
const MAX_VISIBLE = 5;

/** 全局信息提示框：右上角浮层，成功/错误/信息/警告四种类型。 */
export function ToastProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const timersRef = useRef<Map<number, number>>(new Map());

  const dismiss = useCallback((id: number) => {
    // 先标记 leaving 播放退场动画，再真正移除
    setItems((prev) => prev.map((it) => (it.id === id ? { ...it, leaving: true } : it)));
    window.setTimeout(() => {
      setItems((prev) => prev.filter((it) => it.id !== id));
      timersRef.current.delete(id);
    }, 200);
  }, []);

  const push = useCallback(
    (type: ToastType, message: string) => {
      const id = nextId++;
      setItems((prev) => [...prev.slice(-(MAX_VISIBLE - 1)), { id, type, message }]);
      const t = window.setTimeout(() => dismiss(id), DURATION[type]);
      timersRef.current.set(id, t);
    },
    [dismiss]
  );

  const apiRef = useRef<ToastApi>({
    success: (m) => push("success", m),
    error: (m) => push("error", m),
    info: (m) => push("info", m),
    warning: (m) => push("warning", m),
  });
  apiRef.current.success = (m) => push("success", m);
  apiRef.current.error = (m) => push("error", m);
  apiRef.current.info = (m) => push("info", m);
  apiRef.current.warning = (m) => push("warning", m);

  // 注册全局错误处理器（http.ts 自动上报）
  useEffect(() => {
    setGlobalToastHandler({ error: (m) => push("error", m) });
    return () => setGlobalToastHandler(null);
  }, [push]);

  // 卸载时清理所有定时器
  useEffect(() => {
    const map = timersRef.current;
    return () => map.forEach((t) => window.clearTimeout(t));
  }, []);

  return (
    <ToastContext.Provider value={apiRef.current}>
      {children}
      <div className="toast-container" aria-live="polite" aria-label="全局提示">
        {items.map((it) => (
          <div
            key={it.id}
            className={`toast toast-${it.type}${it.leaving ? " leaving" : ""}`}
            role="status"
          >
            <span className="toast-icon">{ICONS[it.type]}</span>
            <span className="toast-message">{it.message}</span>
            <button
              type="button"
              className="toast-close"
              onClick={() => dismiss(it.id)}
              aria-label="关闭"
            >
              <svg viewBox="0 0 12 12" width="12" height="12" fill="none" aria-hidden="true">
                <path d="m3 3 6 6M9 3l-6 6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastApi {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast 必须在 ToastProvider 内使用");
  return ctx;
}
