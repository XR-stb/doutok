import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import "./Login.css";

export default function Login() {
  const navigate = useNavigate();
  const { login, register, loading, error, clearError } = useAuthStore();
  const [isRegister, setIsRegister] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [nickname, setNickname] = useState("");

  const handleSubmit = async () => {
    if (!username || !password) return;
    if (isRegister && !nickname) return;

    let ok: boolean;
    if (isRegister) {
      ok = await register(username, password, nickname);
    } else {
      ok = await login(username, password);
    }
    if (ok) {
      navigate("/");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") handleSubmit();
  };

  const toggleMode = () => {
    setIsRegister(!isRegister);
    clearError();
  };

  return (
    <div className="login-page">
      <div className="login-logo">
        <div className="login-logo-icon">
          <svg viewBox="0 0 48 48" width="64" height="64" fill="none">
            <path
              d="M34.5 8.5a8 8 0 0 0 8 8v6a14 14 0 0 1-8-2.5V30a12 12 0 1 1-10-11.8v6.3a6 6 0 1 0 4 5.5V4h6v4.5z"
              fill="currentColor"
            />
          </svg>
        </div>
        <h1>DouTok</h1>
        <p>{isRegister ? "Create your account" : "Welcome back"}</p>
      </div>

      {error && (
        <div className="login-error" onClick={clearError}>
          {error}
        </div>
      )}

      <div className="login-form" onKeyDown={handleKeyDown}>
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="login-input"
          autoComplete="username"
        />
        {isRegister && (
          <input
            type="text"
            placeholder="Nickname"
            value={nickname}
            onChange={(e) => setNickname(e.target.value)}
            className="login-input"
          />
        )}
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="login-input"
          autoComplete={isRegister ? "new-password" : "current-password"}
        />
        <button
          className="login-btn"
          onClick={handleSubmit}
          disabled={loading || !username || !password || (isRegister && !nickname)}
        >
          {loading ? (
            <span className="login-spinner" />
          ) : isRegister ? (
            "Sign Up"
          ) : (
            "Log In"
          )}
        </button>
      </div>

      <div className="login-footer">
        <button className="login-toggle" onClick={toggleMode}>
          {isRegister
            ? "Already have an account? Log In"
            : "No account? Sign Up"}
        </button>
        <button className="login-skip" onClick={() => navigate("/")}>
          Browse as guest
        </button>
      </div>
    </div>
  );
}
