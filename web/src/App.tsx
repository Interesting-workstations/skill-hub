import { Suspense } from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import ErrorBoundary from "./app/ErrorBoundary";
import MainLayout from "./layouts/MainLayout";
import PageLoading from "./components/shared/PageLoading";
import { routes } from "./app/router/routes";
import "./App.css";

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <div className="app">
          <Suspense fallback={<PageLoading />}>
            <Routes>
              <Route element={<MainLayout />}>
                {routes.map((route) => (
                  <Route key={route.path} path={route.path} element={route.element} />
                ))}
              </Route>
            </Routes>
          </Suspense>
        </div>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
