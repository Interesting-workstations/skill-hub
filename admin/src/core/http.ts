/** 统一 HTTP 请求封装：处理 { code, message, data } 响应格式。 */

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

async function request<T>(path: string, options: RequestInit = {}, retries = 1): Promise<T> {
  try {
    const token = localStorage.getItem("admin_token");
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    };

    const res = await fetch(path, { ...options, headers });
    if (!res.ok) {
      throw new ApiError(res.status, `请求失败 (${res.status})`);
    }
    const body = (await res.json()) as ApiResponse<T>;
    if (body.code !== 0) {
      throw new ApiError(body.code, body.message || "请求失败");
    }
    return body.data;
  } catch (err) {
    // GET 请求失败时自动重试一次（dev 代理冷启动等瞬时问题自愈）
    const isGet = !options.method || options.method === "GET";
    if (isGet && retries > 0) {
      await new Promise((r) => setTimeout(r, 400));
      return request<T>(path, options, retries - 1);
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
