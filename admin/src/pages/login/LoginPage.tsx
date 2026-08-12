import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { auth } from "../../core/auth";
import { authApi } from "../../api/auth";
import "./LoginPage.css";

export default function LoginPage() {
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const result = await authApi.login(username.trim(), password);
      auth.login(result.token, result.user);
      navigate("/", { replace: true });
    } catch {
      setError("用户名或密码错误");
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
