/** 全局状态：管理员信息 + 侧边栏折叠。 */

import { createContext, useContext, useState, type ReactNode } from "react";
import { auth, type AdminUser } from "../core/auth";

interface AppState {
  user: AdminUser | null;
  sidebarCollapsed: boolean;
  toggleSidebar: () => void;
}

const AppContext = createContext<AppState | null>(null);

export function AppProvider({ children }: { children: ReactNode }) {
  const [user] = useState<AdminUser | null>(() => auth.getUser());
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  const toggleSidebar = () => setSidebarCollapsed((v) => !v);

  return (
    <AppContext.Provider value={{ user, sidebarCollapsed, toggleSidebar }}>
      {children}
    </AppContext.Provider>
  );
}

export function useApp(): AppState {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error("useApp 必须在 AppProvider 内使用");
  return ctx;
}
