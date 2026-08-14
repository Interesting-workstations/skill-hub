import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { auth } from "../../core/auth";
import { authApi } from "../../api/auth";
import "./LoginPage.css";

export default function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  // 登录过期（Token 失效）跳转回来的提示
  const expired = searchParams.get("expired") === "1";
  const redirect = searchParams.get("redirect") || "/";

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const result = await authApi.login(username.trim(), password);
      // 登录成功：立即清空内存中的密码变量（前端不保存密码）
      setPassword("");
      auth.login(result);
      navigate(redirect.startsWith("/") ? redirect : "/", { replace: true });
    } catch (err) {
      // 统一安全提示，不暴露具体失败原因
      const msg = err instanceof Error ? err.message : "";
      if (msg.includes("锁定") || msg.includes("频繁")) {
        setError(msg);
      } else {
        setError("用户名或密码错误");
      }
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-brand">
          <div className="login-logo">S</div>
          <h1>Agent Skills 运营后台</h1>
          <p>单管理员运营后台 · 官网 / 爬虫 / 数据</p>
        </div>
        {expired && (
          <div className="login-expired">登录已过期，请重新登录</div>
        )}
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
          账号 admin / 密码 admin123（由 Go 后端校验）
        </div>
      </div>
    </div>
  );
}
