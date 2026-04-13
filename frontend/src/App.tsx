import { useEffect } from "react";
import { Routes, Route } from "react-router-dom";
import { useAuthStore } from "./stores/authStore";
import MainLayout from "./components/Layout/MainLayout";
import DebugPanel from "./components/Debug/DebugPanel";
import Feed from "./pages/Feed";
import Discover from "./pages/Discover";
import Upload from "./pages/Upload";
import Inbox from "./pages/Inbox";
import Profile from "./pages/Profile";
import VideoDetail from "./pages/VideoDetail";
import Live from "./pages/Live";
import Chat from "./pages/Chat";
import Login from "./pages/Login";
import api from "./services/api";

function App() {
  const { debugMode, restoreSession, token } = useAuthStore();

  // Restore session & custom API URL on mount
  useEffect(() => {
    // Restore custom API base URL from localStorage
    const savedUrl = localStorage.getItem("doutok_api_url");
    if (savedUrl) {
      api.defaults.baseURL = savedUrl.endsWith("/api/v1")
        ? savedUrl
        : `${savedUrl}/api/v1`;
    }

    // Set auth header if token exists
    if (token) {
      api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    }

    restoreSession();
  }, []);

  // Keep auth header in sync
  useEffect(() => {
    if (token) {
      api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    } else {
      delete api.defaults.headers.common["Authorization"];
    }
  }, [token]);

  return (
    <div className="app-container">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<MainLayout />}>
          <Route path="/" element={<Feed />} />
          <Route path="/discover" element={<Discover />} />
          <Route path="/upload" element={<Upload />} />
          <Route path="/inbox" element={<Inbox />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/profile/:id" element={<Profile />} />
        </Route>
        <Route path="/video/:id" element={<VideoDetail />} />
        <Route path="/live/:id" element={<Live />} />
        <Route path="/chat/:id" element={<Chat />} />
      </Routes>
      {debugMode && <DebugPanel />}
    </div>
  );
}

export default App;
