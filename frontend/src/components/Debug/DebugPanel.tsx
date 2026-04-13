import { useState, useEffect } from "react";
import { useAuthStore } from "../../stores/authStore";
import { X, Server, Wifi, WifiOff, RefreshCw, Trash2, Bug, ChevronRight } from "lucide-react";
import api from "../../services/api";
import "./DebugPanel.css";

const VERSION = "0.1.0-debug";

export default function DebugPanel() {
  const { debugMode, user, token, isLoggedIn } = useAuthStore();
  const [expanded, setExpanded] = useState(false);
  const [activeSection, setActiveSection] = useState<string | null>(null);
  const [apiUrl, setApiUrl] = useState(() => localStorage.getItem("doutok_api_url") || "");
  const [apiStatus, setApiStatus] = useState<"unknown" | "ok" | "error">("unknown");
  const [apiLatency, setApiLatency] = useState(0);
  const [envInfo, setEnvInfo] = useState<Record<string, string>>({});
  const [logs, setLogs] = useState<string[]>([]);

  useEffect(() => {
    if (!debugMode) return;
    setEnvInfo({
      "User Agent": navigator.userAgent.slice(0, 60) + "...",
      "Screen": `${screen.width}x${screen.height} @${devicePixelRatio}x`,
      "Viewport": `${window.innerWidth}x${window.innerHeight}`,
      "Platform": navigator.platform,
      "Language": navigator.language,
      "Online": navigator.onLine ? "Yes" : "No",
      "Memory": (navigator as any).deviceMemory ? `${(navigator as any).deviceMemory}GB` : "N/A",
      "Cores": navigator.hardwareConcurrency ? `${navigator.hardwareConcurrency}` : "N/A",
    });
  }, [debugMode]);

  if (!debugMode) return null;

  const addLog = (msg: string) => {
    setLogs((prev) => [`[${new Date().toLocaleTimeString()}] ${msg}`, ...prev].slice(0, 50));
  };

  const testApi = async () => {
    setApiStatus("unknown");
    const start = performance.now();
    try {
      await api.get("/health");
      const latency = Math.round(performance.now() - start);
      setApiLatency(latency);
      setApiStatus("ok");
      addLog(`API health check OK (${latency}ms)`);
    } catch (e: any) {
      setApiStatus("error");
      setApiLatency(0);
      addLog(`API health check FAILED: ${e.message}`);
    }
  };

  const applyApiUrl = () => {
    const url = apiUrl.trim();
    if (url) {
      localStorage.setItem("doutok_api_url", url);
      api.defaults.baseURL = url.endsWith("/api/v1") ? url : `${url}/api/v1`;
      addLog(`API URL changed to: ${api.defaults.baseURL}`);
    } else {
      localStorage.removeItem("doutok_api_url");
      api.defaults.baseURL = "/api/v1";
      addLog("API URL reset to default: /api/v1");
    }
    testApi();
  };

  const clearStorage = () => {
    const keys = Object.keys(localStorage).filter((k) => k.startsWith("doutok_"));
    keys.forEach((k) => localStorage.removeItem(k));
    addLog(`Cleared ${keys.length} storage keys`);
  };

  // Minimized pill
  if (!expanded) {
    return (
      <button className="debug-pill" onClick={() => setExpanded(true)}>
        <Bug size={14} />
        <span>DEBUG</span>
      </button>
    );
  }

  const sections = [
    { id: "server", label: "API Server", icon: Server },
    { id: "env", label: "Environment", icon: Wifi },
    { id: "auth", label: "Auth State", icon: RefreshCw },
    { id: "logs", label: "Logs", icon: Bug },
  ];

  return (
    <div className="debug-overlay">
      <div className="debug-panel">
        {/* Header */}
        <div className="debug-header">
          <div className="debug-title">
            <Bug size={18} />
            <span>Debug Panel</span>
            <span className="debug-version">{VERSION}</span>
          </div>
          <button className="debug-close" onClick={() => setExpanded(false)}>
            <X size={20} />
          </button>
        </div>

        {/* Status bar */}
        <div className="debug-status-bar">
          <div className={`debug-status-dot ${apiStatus}`} />
          <span>API: {apiStatus === "ok" ? `Connected (${apiLatency}ms)` : apiStatus === "error" ? "Disconnected" : "Unknown"}</span>
          <span className="debug-status-sep">|</span>
          <span>Auth: {isLoggedIn ? `✓ ${user?.username}` : "Not logged in"}</span>
        </div>

        {/* Sections */}
        <div className="debug-sections">
          {sections.map((s) => {
            const Icon = s.icon;
            const isOpen = activeSection === s.id;
            return (
              <div key={s.id} className="debug-section">
                <button
                  className={`debug-section-header ${isOpen ? "active" : ""}`}
                  onClick={() => setActiveSection(isOpen ? null : s.id)}
                >
                  <Icon size={16} />
                  <span>{s.label}</span>
                  <ChevronRight size={16} className={`debug-chevron ${isOpen ? "open" : ""}`} />
                </button>

                {isOpen && s.id === "server" && (
                  <div className="debug-section-body">
                    <div className="debug-field">
                      <label>API Base URL</label>
                      <div className="debug-input-row">
                        <input
                          type="text"
                          value={apiUrl}
                          onChange={(e) => setApiUrl(e.target.value)}
                          placeholder="e.g. http://192.168.1.100:8080"
                          className="debug-input"
                        />
                        <button className="debug-btn" onClick={applyApiUrl}>Apply</button>
                      </div>
                      <p className="debug-hint">
                        Leave empty for default (/api/v1). For mobile, enter your server's LAN IP.
                      </p>
                    </div>
                    <div className="debug-field">
                      <label>Connection Test</label>
                      <div className="debug-input-row">
                        <div className={`debug-api-status ${apiStatus}`}>
                          {apiStatus === "ok" ? <Wifi size={14} /> : <WifiOff size={14} />}
                          <span>{apiStatus === "ok" ? `OK ${apiLatency}ms` : apiStatus === "error" ? "Failed" : "Not tested"}</span>
                        </div>
                        <button className="debug-btn" onClick={testApi}>Test</button>
                      </div>
                    </div>
                    <div className="debug-field">
                      <label>Current baseURL</label>
                      <code className="debug-code">{api.defaults.baseURL}</code>
                    </div>
                  </div>
                )}

                {isOpen && s.id === "env" && (
                  <div className="debug-section-body">
                    {Object.entries(envInfo).map(([k, v]) => (
                      <div key={k} className="debug-kv">
                        <span className="debug-key">{k}</span>
                        <span className="debug-value">{v}</span>
                      </div>
                    ))}
                  </div>
                )}

                {isOpen && s.id === "auth" && (
                  <div className="debug-section-body">
                    <div className="debug-kv">
                      <span className="debug-key">Logged In</span>
                      <span className="debug-value">{isLoggedIn ? "Yes" : "No"}</span>
                    </div>
                    {user && (
                      <>
                        <div className="debug-kv">
                          <span className="debug-key">User ID</span>
                          <span className="debug-value">{user.id}</span>
                        </div>
                        <div className="debug-kv">
                          <span className="debug-key">Username</span>
                          <span className="debug-value">{user.username}</span>
                        </div>
                      </>
                    )}
                    <div className="debug-kv">
                      <span className="debug-key">Token</span>
                      <span className="debug-value debug-token">{token ? token.slice(0, 20) + "..." : "N/A"}</span>
                    </div>
                    <button className="debug-btn debug-btn-danger" onClick={clearStorage}>
                      <Trash2 size={14} /> Clear All Storage
                    </button>
                  </div>
                )}

                {isOpen && s.id === "logs" && (
                  <div className="debug-section-body">
                    <div className="debug-logs">
                      {logs.length === 0 && <span className="debug-hint">No logs yet. Try testing the API.</span>}
                      {logs.map((log, i) => (
                        <div key={i} className="debug-log-line">{log}</div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
