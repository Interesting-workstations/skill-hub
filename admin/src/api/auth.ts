/** 管理后台认证 API —— 对接 Go 后端。 */

import { http } from "../core/http";
import type { LoginResult } from "../core/auth";

export const authApi = {
  /** 账号密码登录：返回 Access + Refresh 凭证 */
  login(username: string, password: string): Promise<LoginResult> {
    return http.post<LoginResult>("/api/v1/admin/login", { username, password });
  },
  /** 主动退出（携带 Refresh Token 使服务端凭证作废） */
  logout(refreshToken?: string): Promise<unknown> {
    return http.post<unknown>("/api/v1/admin/logout", refreshToken ? { refreshToken } : {});
  },
  /** 会话状态检查（Token 有效性由后端裁决） */
  session(): Promise<unknown> {
    return http.get<unknown>("/api/v1/admin/session");
  },
  /** 修改管理员密码（改密后全部旧 Token 失效） */
  changePassword(oldPassword: string, newPassword: string): Promise<void> {
    return http.put<void>("/api/v1/admin/password", { oldPassword, newPassword });
  },
};
