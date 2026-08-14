/** 统一 HTTP 请求封装：处理 { code, message, data } 响应格式。
 *  安全特性：
 *  - 自动携带 Bearer Access Token
 *  - 收到 401 时用 Refresh Token 自动续期并重放请求（一次性，防重放）
 *  - 续期失败（登录过期）→ 清除凭证 → 跳转登录页并提示
 *  - 登录 / 刷新接口自身的 401 不触发续期，避免循环 */

import { auth } from "./auth";
import { reportGlobalError } from "./toast";

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export class ApiError extends Error {
  code: number;
  constructor(code: number, message: string) {
    super(message);
    this.code = code;
  }
}

/** 这些路径的 401 不应触发自动续期（自身就是认证流程） */
function isAuthPath(path: string): boolean {
  return path.includes("/admin/login") || path.includes("/admin/refresh") || path.includes("/admin/logout");
}

/** 全局串行化的刷新锁：并发 401 时只发起一次 refresh */
let refreshing: Promise<boolean> | null = null;

async function refreshAccessToken(): Promise<boolean> {
  if (refreshing) return refreshing;
  refreshing = (async () => {
    const refresh = auth.getRefreshToken();
    if (!refresh) return false;
    try {
      const res = await fetch("/api/v1/admin/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken: refresh }),
      });
      if (!res.ok) return false;
      const body = (await res.json()) as ApiResponse<{ token: string; refreshToken: string }>;
      if (body.code !== 0 || !body.data?.token) return false;
      auth.setTokens(body.data.token, body.data.refreshToken);
      return true;
    } catch {
      return false;
    } finally {
      refreshing = null;
    }
  })();
  return refreshing;
}

/** 登录过期：清除凭证并跳转登录页（带 expired 提示） */
function forceLogout() {
  auth.clear();
  const target = `/login?expired=1&redirect=${encodeURIComponent(
    window.location.pathname + window.location.search
  )}`;
  if (window.location.pathname !== "/login") {
    window.location.assign(target);
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  retries = 1,
  allowAuthRetry = true
): Promise<T> {
  try {
    const token = auth.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    };

    const res = await fetch(path, { ...options, headers });

    // Access Token 过期：尝试用 Refresh Token 续期后重放一次
    if (res.status === 401 && allowAuthRetry && !isAuthPath(path)) {
      const renewed = await refreshAccessToken();
      if (renewed) {
        return request<T>(path, options, retries, false);
      }
      forceLogout();
      throw new ApiError(40101, "登录已过期，请重新登录");
    }

    if (!res.ok) {
      // 尝试解析后端返回的业务错误信息（如「名称必填」），解析失败时回退为通用信息
      let message = `请求失败 (${res.status})`;
      try {
        const errBody = (await res.json()) as ApiResponse<unknown>;
        if (errBody.message) message = errBody.message;
      } catch {
        // 非 JSON 响应，保留默认信息
      }
      throw new ApiError(res.status, message);
    }
    const body = (await res.json()) as ApiResponse<T>;
    if (body.code !== 0) {
      // 业务层 401（Token 无效）同样按登录过期处理
      if (body.code === 40101 && allowAuthRetry && !isAuthPath(path)) {
        const renewed = await refreshAccessToken();
        if (renewed) {
          return request<T>(path, options, retries, false);
        }
        forceLogout();
      }
      throw new ApiError(body.code, body.message || "请求失败");
    }
    return body.data;
  } catch (err) {
    // GET 请求失败时自动重试一次（dev 代理冷启动等瞬时问题自愈）
    const isGet = !options.method || options.method === "GET";
    if (isGet && retries > 0) {
      await new Promise((r) => setTimeout(r, 400));
      return request<T>(path, options, retries - 1, allowAuthRetry);
    }
    // 全局错误提示（认证相关由登录页自行提示，不弹 Toast）
    if (
      err instanceof ApiError &&
      !isAuthPath(path) &&
      err.code !== 401 &&
      err.code !== 40101
    ) {
      reportGlobalError(err.message);
    }
    throw err;
  }
}

export const http = {
  get<T>(path: string) {
    return request<T>(path);
  },
  post<T>(path: string, data?: unknown) {
    return request<T>(path, {
      method: "POST",
      body: data === undefined ? undefined : JSON.stringify(data),
    });
  },
  put<T>(path: string, data?: unknown) {
    return request<T>(path, {
      method: "PUT",
      body: data === undefined ? undefined : JSON.stringify(data),
    });
  },
  delete<T>(path: string) {
    return request<T>(path, { method: "DELETE" });
  },
};
