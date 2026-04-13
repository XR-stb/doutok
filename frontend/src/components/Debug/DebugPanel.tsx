import { useState } from "react";
import { useAuthStore } from "../../stores/authStore";
import { Bug, X, Server, Database, Wifi, Zap } from "lucide-react";
import "./DebugPanel.css";

interface LogEntry {
  time: string;
  level: "info" | "warn" | "error";
  msg: string;
}

export default function DebugPanel() {
  const [expanded, setExpanded] = useState(false);
  const [activeTab, setActiveTab] = useState<"info" | "network" | "logs">("info");
  const token = useAuthStore((s) => s.token);

  const sysInfo = {
    env: import.meta.env.MODE,
    api: "/api/v1",
    version: __APP_VERSION__ ?? "0.1.0-dev",
    build: __BUILD_TIME__ ?? new Date().toISOString(),
    platform: navigator.userAgent.includes("Android") ? "Android (Capacitor)" : "Web",
    screen: `${window.innerWidth}x${window.innerHeight}`,
    dpr: window.devicePixelRatio,
    token: token ? `${token.slice(0, 20)}...` : "none",
  };

  const [logs] = useState<LogEntry[]>([
    { time: "15:30:01", level: "info", msg: "App initialized" },
    { time: "15:30:02", level: "info", msg: "Feed loaded: 10 videos" },
  ]);

  if (!expanded) {
    return (
      <button className="debug-fab" onClick={() => setExpanded(true)}>
        <Bug size={18} />
      </button>
    );
  }

  return (
    <div className="debug-panel">
      <div className="debug-header">
        <span>🐛 Debug Panel</span>
        <button onClick={() => setExpanded(false)}>
          <X size={18} />
        </button>
      </div>

      <div className="debug-tabs">
        {(["info", "network", "logs"] as const).map((tab) => (
          <button
            key={tab}
            className={activeTab === tab ? "active" : ""}
            onClick={() => setActiveTab(tab)}
          >
            {tab === "info" && <Server size={14} />}
            {tab === "network" && <Wifi size={14} />}
            {tab === "logs" && <Zap size={14} />}
            {tab}
          </button>
        ))}
      </div>

      <div className="debug-body">
        {activeTab === "info" && (
          <div className="debug-info-list">
            {Object.entries(sysInfo).map(([k, v]) => (
              <div key={k} className="debug-info-row">
                <span className="debug-key">{k}</span>
                <span className="debug-val">{String(v)}</span>
              </div>
            ))}
          </div>
        )}

        {activeTab === "network" && (
          <div className="debug-placeholder">
            <Database size={24} />
            <p>Network inspector coming soon</p>
          </div>
        )}

        {activeTab === "logs" && (
          <div className="debug-log-list">
            {logs.map((log, i) => (
              <div key={i} className={`debug-log-item ${log.level}`}>
                <span className="log-time">{log.time}</span>
                <span className={`log-level ${log.level}`}>{log.level}</span>
                <span className="log-msg">{log.msg}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

// Vite env type declarations
declare const __APP_VERSION__: string;
declare const __BUILD_TIME__: string;
