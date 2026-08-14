import { NavLink, Outlet, useNavigate } from "react-router-dom";
import { useApp } from "../../store/AppContext";
import { auth } from "../../core/auth";
import { authApi } from "../../api/auth";
import { site } from "../../config/site";
import "./AdminLayout.css";

interface MenuItem {
  path?: string;
  label?: string;
  icon?: string;
  section?: string;
}

const MENU: MenuItem[] = [
  { path: "/", label: "工作台", icon: "📊" },
  { section: "爬虫管理" },
  { path: "/crawler/tasks", label: "爬虫任务", icon: "🕷️" },
  { path: "/crawler/executions", label: "执行记录", icon: "🕐" },
  { path: "/crawler/failures", label: "失败任务", icon: "⚠️" },
  { path: "/crawler/config", label: "爬虫配置", icon: "🧩" },
  { section: "数据管理" },
  { path: "/data/items", label: "抓取数据", icon: "🗂️" },
  { path: "/data/audit", label: "数据审核", icon: "✅" },
  { path: "/data/export", label: "数据导出", icon: "📤" },
  { section: "官网内容" },
  { path: "/content/categories", label: "分类管理", icon: "📑" },
  { path: "/content/homepage", label: "首页内容", icon: "🏠" },
  { path: "/content/articles", label: "文章管理", icon: "📝" },
  { path: "/content/sponsors", label: "赞助管理", icon: "🪧" },
  { path: "/content/seo", label: "SEO 配置", icon: "🔍" },
  { section: "系统设置" },
  { path: "/system/admin", label: "管理员设置", icon: "👤" },
  { path: "/system/settings", label: "网站配置", icon: "🛠️" },
];

export default function AdminLayout() {
  const { user, sidebarCollapsed, toggleSidebar } = useApp();
  const navigate = useNavigate();

  const handleLogout = async () => {
    // 先通知后端作废凭证（黑名单），失败也继续本地退出
    try {
      await authApi.logout(auth.getRefreshToken() ?? undefined);
    } catch {
      // 忽略：本地凭证仍会清除
    }
    auth.clear();
    navigate("/login");
  };

  return (
    <div className="admin-layout">
      <aside className={`sidebar ${sidebarCollapsed ? "collapsed" : ""}`}>
        <div className="sidebar-brand">
          <div className="logo">S</div>
          {!sidebarCollapsed && <div className="name">{site.shortName}</div>}
        </div>
        <nav className="sidebar-nav">
          {MENU.map((item, i) =>
            item.section ? (
              <div className="nav-section" key={i}>{item.section}</div>
            ) : (
              <NavLink
                key={item.path}
                to={item.path!}
                end={item.path === "/"}
                className={({ isActive }) => `nav-item ${isActive ? "active" : ""}`}
              >
                <span className="nav-icon">{item.icon}</span>
                <span className="nav-label">{item.label}</span>
              </NavLink>
            ),
          )}
        </nav>
      </aside>

      <div className="main">
        <header className="header">
          <div className="header-left">
            <button className="icon-btn" onClick={toggleSidebar} title="折叠/展开菜单">
              ☰
            </button>
            <span className="header-title">运营后台</span>
          </div>
          <div className="header-right">
            <a className="portal-link" href={site.portalUrl} target="_blank" rel="noreferrer">
              ↗ 查看官网
            </a>
            <div className="user-box">
              <span className="user-avatar">{user?.displayName?.charAt(0) ?? "A"}</span>
              <span className="user-name">{user?.displayName ?? "admin"}</span>
              <button className="btn-link danger" onClick={handleLogout}>退出</button>
            </div>
          </div>
        </header>
        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
