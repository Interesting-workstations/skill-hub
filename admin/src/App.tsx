import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AppProvider } from "./store/AppContext";
import { ToastProvider } from "./components/Toast";
import { auth } from "./core/auth";
import AdminLayout from "./layouts/AdminLayout/AdminLayout";
import LoginPage from "./pages/login/LoginPage";
import DashboardPage from "./pages/dashboard/DashboardPage";
import TaskListPage from "./pages/crawler/TaskListPage";
import ExecutionListPage from "./pages/crawler/ExecutionListPage";
import ExecutionDetailPage from "./pages/crawler/ExecutionDetailPage";
import FailurePage from "./pages/crawler/FailurePage";
import ConfigPage from "./pages/crawler/ConfigPage";
import OfficialOrgsPage from "./pages/crawler/OfficialOrgsPage";
import TokenPoolPage from "./pages/crawler/TokenPoolPage";
import CrawlerDataPage from "./pages/data/CrawlerDataPage";
import AuditPage from "./pages/data/AuditPage";
import ExportPage from "./pages/data/ExportPage";
import CategoryPage from "./pages/content/CategoryPage";
import HomepagePage from "./pages/content/HomepagePage";
import ArticlePage from "./pages/content/ArticlePage";
import SeoPage from "./pages/content/SeoPage";
import SponsorsPage from "./pages/content/SponsorsPage";
import AdminSettingsPage from "./pages/system/AdminSettingsPage";
import SiteSettingsPage from "./pages/system/SiteSettingsPage";

/** 认证守卫：未登录跳转登录页 */
function RequireAuth({ children }: { children: React.ReactNode }) {
  if (!auth.isLoggedIn()) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}

export default function App() {
  return (
    <ToastProvider>
      <AppProvider>
        <BrowserRouter basename={import.meta.env.VITE_ADMIN_BASE || "/"}>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route
              path="/"
              element={
                <RequireAuth>
                  <AdminLayout />
                </RequireAuth>
              }
            >
              <Route index element={<DashboardPage />} />
              <Route path="crawler/tasks" element={<TaskListPage />} />
              <Route path="crawler/executions" element={<ExecutionListPage />} />
              <Route path="crawler/executions/:id" element={<ExecutionDetailPage />} />
              <Route path="crawler/failures" element={<FailurePage />} />
              <Route path="crawler/config" element={<ConfigPage />} />
              <Route path="crawler/tokens" element={<TokenPoolPage />} />
              <Route path="crawler/official-orgs" element={<OfficialOrgsPage />} />
              <Route path="data/items" element={<CrawlerDataPage />} />
              <Route path="data/audit" element={<AuditPage />} />
              <Route path="data/export" element={<ExportPage />} />
              <Route path="content/categories" element={<CategoryPage />} />
              <Route path="content/homepage" element={<HomepagePage />} />
              <Route path="content/articles" element={<ArticlePage />} />
              <Route path="content/sponsors" element={<SponsorsPage />} />
              <Route path="content/seo" element={<SeoPage />} />
              <Route path="system/admin" element={<AdminSettingsPage />} />
              <Route path="system/settings" element={<SiteSettingsPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AppProvider>
    </ToastProvider>
  );
}
