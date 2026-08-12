/** 管理后台认证 API —— 对接 Go 后端。 */

import { http } from "../core/http";
import type { AdminUser } from "../types";

export interface LoginResult {
  token: string;
  user: AdminUser;
}

export const authApi = {
  login(username: string, password: string): Promise<LoginResult> {
    return http.post<LoginResult>("/api/v1/admin/login", { username, password });
  },
  /** 修改管理员密码 */
  changePassword(oldPassword: string, newPassword: string): Promise<void> {
    return http.put<void>("/api/v1/admin/password", { oldPassword, newPassword });
  },
};
