import { useEffect, useRef, useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import Navbar from "./Navbar";
import Footer from "./Footer";
import { createAnimationContext } from "../animations";

export default function Layout() {
  const location = useLocation();
  const mainRef = useRef<HTMLElement>(null);
  const ctx = useRef(createAnimationContext());
  const [transitioning, setTransitioning] = useState(false);

  // 路由切换时清理旧动画上下文
  useEffect(() => {
    ctx.current.killAll();
    ctx.current = createAnimationContext();

    // 触发短暂的过渡状态
    setTransitioning(true);
    const timer = setTimeout(() => setTransitioning(false), 50);
    return () => clearTimeout(timer);
  }, [location.pathname]);

  useEffect(() => {
    return () => {
      ctx.current.killAll();
    };
  }, []);

  return (
    <>
      <Navbar />
      <main
        className="main"
        ref={mainRef}
        style={{
          opacity: transitioning ? 0.3 : 1,
          transition: "opacity 0.15s ease",
        }}
      >
        <Outlet />
      </main>
      <Footer />
    </>
  );
}
