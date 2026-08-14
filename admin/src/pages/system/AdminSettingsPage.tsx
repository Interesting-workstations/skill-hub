import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { auth } from "../../core/auth";
import { authApi } from "../../api/auth";

export default function AdminSettingsPage() {
  const navigate = useNavigate();
  const [displayName, setDisplayName] = useState("管理员");
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [confirmPwd, setConfirmPwd] = useState("");
  const [saved, setSaved] = useState(false);
  const [pwdMsg, setPwdMsg] = useState("");
  const [pwdSaving, setPwdSaving] = useState(false);

  const saveProfile = (e: FormEvent) => {
    e.preventDefault();
    setSaved(true);
    setTimeout(() => setSaved(false), 2000);
  };

  const savePassword = async (e: FormEvent) => {
    e.preventDefault();
    if (newPwd !== confirmPwd) {
      setPwdMsg("两次输入的新密码不一致");
      return;
    }
    setPwdMsg("");
    setPwdSaving(true);
    try {
      await authApi.changePassword(oldPwd, newPwd);
      // 改密后全部旧 Token 失效：清除本地凭证并跳转登录页重新登录
      auth.clear();
      navigate("/login?expired=1");
    } catch (err) {
      setPwdMsg(err instanceof Error ? err.message : "修改失败");
      setPwdSaving(false);
    }
  };

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1>管理员设置</h1>
          <div className="sub">单管理员账号信息与密码安全</div>
        </div>
      </div>

      <form className="card card-pad form" onSubmit={saveProfile} style={{ maxWidth: 480 }}>
        <div className="card-title">账号信息</div>
        <div className="form-item">
          <label>登录用户名</label>
          <input className="input" value="admin" disabled />
        </div>
        <div className="form-item">
          <label>显示名称</label>
          <input className="input" value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <button className="btn btn-primary" type="submit">保存</button>
          {saved && <span style={{ color: "var(--color-success)", fontSize: 13 }}>✓ 已保存</span>}
        </div>
      </form>

      <form className="card card-pad form" onSubmit={savePassword} style={{ maxWidth: 480 }}>
        <div className="card-title">修改密码</div>
        <div className="form-item">
          <label>当前密码</label>
          <input className="input" type="password" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} required />
        </div>
        <div className="form-grid">
          <div className="form-item">
            <label>新密码</label>
            <input className="input" type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)} required />
          </div>
          <div className="form-item">
            <label>确认新密码</label>
            <input className="input" type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)} required />
          </div>
        </div>
        {pwdMsg && <div style={{ fontSize: 13, color: "var(--color-danger)" }}>{pwdMsg}</div>}
        <div>
          <button className="btn btn-primary" type="submit" disabled={pwdSaving}>
            {pwdSaving ? "提交中…" : "更新密码"}
          </button>
        </div>
      </form>

      <div className="card card-pad">
        <div className="card-title">登录安全</div>
        <div className="desc-grid" style={{ maxWidth: 480 }}>
          <div className="desc-item"><span className="k">登录失败限制</span><span className="v">5 次后锁定 15 分钟</span></div>
          <div className="desc-item"><span className="k">会话有效期</span><span className="v">7 天</span></div>
          <div className="desc-item"><span className="k">操作日志</span><span className="v">已开启（记录登录、审核、发布等关键操作）</span></div>
        </div>
      </div>
    </div>
  );
}
