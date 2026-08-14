// API 客户端：统一处理请求与后端响应格式 { code, message, data }。
// 组件不直接调用 fetch，统一走此模块。

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

/** 发起请求并解包统一响应；业务错误（code !== 0）抛异常 */
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, init);
  if (!res.ok) {
    throw new Error(`请求失败（HTTP ${res.status}）`);
  }
  const body = (await res.json()) as ApiResponse<T>;
  if (body.code !== 0) {
    throw new Error(body.message || "请求失败");
  }
  return body.data;
}
