/** 单管理员认证：Token 存储与登录状态。 */

const TOKEN_KEY = "admin_token";
const USER_KEY = "admin_user";

export interface AdminUser {
  username: string;
  displayName: string;
  role: "admin";
}

export const auth = {
  getToken(): string | null {
    return localStorage.getItem(TOKEN_KEY);
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
    return Boolean(this.getToken() && this.getUser());
  },
  login(token: string, user: AdminUser) {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },
  logout() {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
};
