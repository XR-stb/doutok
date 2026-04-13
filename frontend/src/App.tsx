import { useEffect } from "react";
import { Routes, Route } from "react-router-dom";
import { useAuthStore } from "./stores/authStore";
import api from "./services/api";
import MainLayout from "./components/Layout/MainLayout";
import Feed from "./pages/Feed";
import Discover from "./pages/Discover";
import Upload from "./pages/Upload";
import Inbox from "./pages/Inbox";
import Profile from "./pages/Profile";
import VideoDetail from "./pages/VideoDetail";
import Live from "./pages/Live";
import Chat from "./pages/Chat";
import Login from "./pages/Login";
import DebugPanel from "./components/Debug/DebugPanel";

function App() {
  const isDebug = useAuthStore((s) => s.debugMode);
  const restoreSession = useAuthStore((s) => s.restoreSession);

  useEffect(() => {
    // Setup axios interceptor: attach token to every request
    const interceptor = api.interceptors.request.use((config) => {
      const token = localStorage.getItem("doutok_token");
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // Restore session from localStorage
    restoreSession();

    return () => {
      api.interceptors.request.eject(interceptor);
    };
  }, []);

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
      {isDebug && <DebugPanel />}
    </div>
  );
}

export default App;
