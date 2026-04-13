import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";
import {
  Settings,
  Edit3,
  Grid,
  Heart,
  Bookmark,
  Lock,
  LogOut,
  ChevronRight,
  Share2,
} from "lucide-react";
import "./Profile.css";

export default function Profile() {
  const navigate = useNavigate();
  const { user, isLoggedIn, logout, tapDebug, fetchMe, debugMode, debugTapCount } =
    useAuthStore();
  const [activeTab, setActiveTab] = useState<"works" | "likes" | "favorites">("works");
  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    if (isLoggedIn && !user) {
      fetchMe();
    }
  }, [isLoggedIn]);

  if (!isLoggedIn) {
    return (
      <div className="profile-guest">
        <div className="profile-guest-avatar">
          <div className="guest-avatar-placeholder" />
        </div>
        <p className="profile-guest-text">Log in to see your profile</p>
        <button className="profile-login-btn" onClick={() => navigate("/login")}>
          Log In / Sign Up
        </button>
      </div>
    );
  }

  const handleLogout = () => {
    logout();
    setShowSettings(false);
    navigate("/");
  };

  const formatCount = (n: number): string => {
    if (n >= 10000) return (n / 10000).toFixed(1) + "w";
    if (n >= 1000) return (n / 1000).toFixed(1) + "k";
    return String(n);
  };

  return (
    <div className="profile-page">
      {/* Header */}
      <div className="profile-header safe-top">
        <div className="profile-header-actions">
          <button className="profile-icon-btn" onClick={() => {}}>
            <Share2 size={20} />
          </button>
          <button
            className="profile-icon-btn"
            onClick={() => setShowSettings(!showSettings)}
          >
            <Settings size={20} />
          </button>
        </div>
      </div>

      {/* User Info */}
      <div className="profile-info">
        <div className="profile-avatar-wrap">
          <img
            src={user?.avatar || `https://api.dicebear.com/7.x/initials/svg?seed=${user?.nickname || "U"}`}
            alt="avatar"
            className="profile-avatar"
          />
        </div>
        <h2 className="profile-nickname">{user?.nickname || "User"}</h2>
        <p className="profile-username">@{user?.username || "unknown"}</p>
        {user?.bio && <p className="profile-bio">{user.bio}</p>}

        {/* Stats */}
        <div className="profile-stats">
          <div className="stat-item">
            <span className="stat-value">{formatCount(user?.follow_count || 0)}</span>
            <span className="stat-label">Following</span>
          </div>
          <div className="stat-divider" />
          <div className="stat-item">
            <span className="stat-value">{formatCount(user?.fan_count || 0)}</span>
            <span className="stat-label">Followers</span>
          </div>
          <div className="stat-divider" />
          <div className="stat-item">
            <span className="stat-value">{formatCount(user?.like_count || 0)}</span>
            <span className="stat-label">Likes</span>
          </div>
        </div>

        {/* Edit Profile Button */}
        <div className="profile-actions">
          <button className="profile-edit-btn">
            <Edit3 size={14} />
            Edit profile
          </button>
          <button className="profile-edit-btn secondary">
            <Bookmark size={14} />
            Saved
          </button>
        </div>
      </div>

      {/* Content Tabs */}
      <div className="profile-tabs">
        <button
          className={`profile-tab ${activeTab === "works" ? "active" : ""}`}
          onClick={() => setActiveTab("works")}
        >
          <Grid size={18} />
        </button>
        <button
          className={`profile-tab ${activeTab === "likes" ? "active" : ""}`}
          onClick={() => setActiveTab("likes")}
        >
          <Heart size={18} />
        </button>
        <button
          className={`profile-tab ${activeTab === "favorites" ? "active" : ""}`}
          onClick={() => setActiveTab("favorites")}
        >
          <Lock size={18} />
        </button>
      </div>

      {/* Video Grid */}
      <div className="profile-grid">
        <div className="profile-empty">
          <Grid size={48} strokeWidth={1} />
          <p>No content yet</p>
          <span>Your videos will appear here</span>
        </div>
      </div>

      {/* Version - Debug Entry */}
      <div className="profile-version" onClick={tapDebug}>
        DouTok v0.1.0{debugTapCount > 0 && debugTapCount < 7 ? ` (${debugTapCount}/7)` : ""}
        {debugMode && " [DEBUG]"}
      </div>

      {/* Settings Drawer */}
      {showSettings && (
        <>
          <div className="settings-overlay" onClick={() => setShowSettings(false)} />
          <div className="settings-drawer">
            <div className="settings-header">
              <h3>Settings</h3>
            </div>
            <div className="settings-list">
              <button className="settings-item">
                <span>Account</span>
                <ChevronRight size={18} />
              </button>
              <button className="settings-item">
                <span>Privacy</span>
                <ChevronRight size={18} />
              </button>
              <button className="settings-item">
                <span>Notifications</span>
                <ChevronRight size={18} />
              </button>
              <button className="settings-item danger" onClick={handleLogout}>
                <LogOut size={18} />
                <span>Log Out</span>
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
