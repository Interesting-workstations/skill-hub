/** 单管理员认证：Access + Refresh Token 存储与登录状态。
 *  安全原则：Token 仅存 localStorage（同源隔离）；不保存密码；
 *  所有权限以后端校验结果为准，前端仅作路由引导。 */

const TOKEN_KEY = "admin_token";
const REFRESH_KEY = "admin_refresh";
const USER_KEY = "admin_user";

export interface AdminUser {
  username: string;
  displayName: string;
  role: "admin";
}

export interface LoginResult {
  token: string;
  refreshToken: string;
  user: AdminUser;
}

export const auth = {
  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
  },
  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_KEY);
  },
  getUser(): AdminUser | null {
    const raw = localStorage.getItem(USER_KEY);
    if (!raw) return null;
    try {
      return JSON.parse(raw) as AdminUser;
    } catch {
      return null;
    }
  },
  isLoggedIn(): boolean {
    return Boolean(this.getToken() && this.getRefreshToken() && this.getUser());
  },
  /** 登录成功：保存凭证与用户信息 */
  setTokens(token: string, refreshToken: string) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(REFRESH_KEY, refreshToken);
  },
  login(result: LoginResult) {
    this.setTokens(result.token, result.refreshToken);
    localStorage.setItem(USER_KEY, JSON.stringify(result.user));
  },
  /** 仅清除本地凭证（后端主动注销由调用方发起） */
  clear() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(REFRESH_KEY);
    localStorage.removeItem(USER_KEY);
  },
};
