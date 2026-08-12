import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { auth } from "../../core/auth";
import { site } from "../../config/site";
import "./LoginPage.css";

export default function LoginPage() {
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    // 模拟认证：单管理员
    setTimeout(() => {
      if (
        username === site.defaultAdmin.username &&
        password === site.defaultAdmin.password
      ) {
        auth.login(
          `mock-token-${Date.now()}`,
          { username, displayName: "管理员", role: "admin" },
        );
        navigate("/", { replace: true });
      } else {
        setError("用户名或密码错误");
        setLoading(false);
      }
    }, 400);
  };

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-brand">
          <div className="login-logo">S</div>
          <h1>{site.name}</h1>
          <p>单管理员运营后台 · 官网 / 爬虫 / 数据</p>
        </div>
        <form className="login-form" onSubmit={handleSubmit}>
          <div className="form-item">
            <label>用户名</label>
            <input
              className="input"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="admin"
              autoFocus
            />
          </div>
          <div className="form-item">
            <label>密码</label>
            <input
              className="input"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
            />
          </div>
          {error && <div className="login-error">{error}</div>}
          <button className="btn btn-primary login-submit" type="submit" disabled={loading}>
            {loading ? "登录中…" : "登 录"}
          </button>
        </form>
        <div className="login-hint">
          演示账号：admin / admin123
        </div>
      </div>
    </div>
  );
}
