import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { Home, Compass, PlusSquare, MessageCircle, User } from "lucide-react";
import "./MainLayout.css";

const tabs = [
  { path: "/", icon: Home, label: "首页" },
  { path: "/discover", icon: Compass, label: "发现" },
  { path: "/upload", icon: PlusSquare, label: "", isCenter: true },
  { path: "/inbox", icon: MessageCircle, label: "消息" },
  { path: "/profile", icon: User, label: "我" },
];

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();

  return (
    <div className="main-layout">
      <div className="main-content">
        <Outlet />
      </div>
      <nav className="tab-bar safe-bottom">
        {tabs.map((tab) => {
          const active = location.pathname === tab.path;
          const Icon = tab.icon;

          if (tab.isCenter) {
            return (
              <button
                key={tab.path}
                className="tab-item tab-center"
                onClick={() => navigate(tab.path)}
              >
                <div className="tab-center-btn">
                  <PlusSquare size={24} />
                </div>
              </button>
            );
          }

          return (
            <button
              key={tab.path}
              className={`tab-item ${active ? "active" : ""}`}
              onClick={() => navigate(tab.path)}
            >
              <Icon size={22} />
              <span>{tab.label}</span>
            </button>
          );
        })}
      </nav>
    </div>
  );
}
