// 通用异步数据 Hook：管理 loading / error / data 三态。
// deps 由调用方显式传入并受控，effect 不会因 setState 产生无限更新链，
// 此处禁用 exhaustive-deps 误报。
/* oxlint-disable react-hooks/exhaustive-deps */
import { useEffect, useState } from "react";

export interface AsyncState<T> {
  data: T | null;
  loading: boolean;
  error: Error | null;
}

/**
 * 执行异步请求并管理状态。
 * @param fetcher 返回 Promise 的数据获取函数
 * @param deps 依赖数组（变化时重新请求）
 */
export function useAsyncData<T>(
  fetcher: () => Promise<T>,
  deps: readonly unknown[]
): AsyncState<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    fetcher()
      .then((d) => {
        if (!cancelled) {
          setData(d);
          setLoading(false);
        }
      })
      .catch((e: unknown) => {
        if (!cancelled) {
          setError(e instanceof Error ? e : new Error(String(e)));
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return { data, loading, error };
}
