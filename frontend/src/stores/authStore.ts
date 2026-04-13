import { create } from "zustand";
import { authAPI, userAPI } from "../services/api";

interface User {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
  bio: string;
  follow_count: number;
  fan_count: number;
  like_count: number;
  video_count?: number;
}

interface AuthState {
  token: string | null;
  user: User | null;
  isLoggedIn: boolean;
  debugMode: boolean;
  debugTapCount: number;
  loading: boolean;
  error: string | null;

  login: (username: string, password: string) => Promise<boolean>;
  register: (
    username: string,
    password: string,
    nickname: string
  ) => Promise<boolean>;
  logout: () => void;
  fetchMe: () => Promise<void>;
  updateProfile: (
    data: Partial<{ nickname: string; bio: string; avatar: string }>
  ) => Promise<void>;
  tapDebug: () => void;
  clearError: () => void;
  restoreSession: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem("doutok_token"),
  user: null,
  isLoggedIn: false,
  debugMode: localStorage.getItem("doutok_debug") === "true",
  debugTapCount: 0,
  loading: false,
  error: null,

  login: async (username, password) => {
    set({ loading: true, error: null });
    try {
      const res: any = await authAPI.login({ username, password });
      if (res.code !== 0) {
        set({ loading: false, error: res.msg || "Login failed" });
        return false;
      }
      const d = res.data;
      const token = d.token;
      const user: User = {
        id: d.user_id,
        username: d.username,
        nickname: d.nickname,
        avatar: d.avatar || "",
        bio: d.bio || "",
        follow_count: d.follow_count || 0,
        fan_count: d.fan_count || 0,
        like_count: d.like_count || 0,
        video_count: d.video_count || 0,
      };
      localStorage.setItem("doutok_token", token);
      set({ token, user, isLoggedIn: true, loading: false });
      return true;
    } catch (err: any) {
      const msg = err?.msg || err?.message || "Login failed";
      set({ loading: false, error: msg });
      return false;
    }
  },

  register: async (username, password, nickname) => {
    set({ loading: true, error: null });
    try {
      const res: any = await authAPI.register({ username, password, nickname });
      if (res.code !== 0) {
        set({ loading: false, error: res.msg || "Register failed" });
        return false;
      }
      const d = res.data;
      const token = d.token;
      const user: User = {
        id: d.user_id,
        username: d.username,
        nickname: d.nickname,
        avatar: d.avatar || "",
        bio: "",
        follow_count: 0,
        fan_count: 0,
        like_count: 0,
        video_count: 0,
      };
      localStorage.setItem("doutok_token", token);
      set({ token, user, isLoggedIn: true, loading: false });
      return true;
    } catch (err: any) {
      const msg = err?.msg || err?.message || "Register failed";
      set({ loading: false, error: msg });
      return false;
    }
  },

  logout: () => {
    localStorage.removeItem("doutok_token");
    set({ token: null, user: null, isLoggedIn: false });
  },

  fetchMe: async () => {
    try {
      const res: any = await userAPI.getMe();
      if (res.code === 0 && res.data) {
        const d = res.data;
        set({
          user: {
            id: d.id,
            username: d.username,
            nickname: d.nickname,
            avatar: d.avatar || "",
            bio: d.bio || "",
            follow_count: d.follow_count || 0,
            fan_count: d.fan_count || 0,
            like_count: d.like_count || 0,
            video_count: d.video_count || 0,
          },
          isLoggedIn: true,
        });
      }
    } catch {
      // Token invalid
      localStorage.removeItem("doutok_token");
      set({ token: null, user: null, isLoggedIn: false });
    }
  },

  updateProfile: async (data) => {
    try {
      const res: any = await userAPI.updateMe(data);
      if (res.code === 0) {
        const { user } = get();
        if (user) {
          set({ user: { ...user, ...data } });
        }
      }
    } catch {
      // ignore
    }
  },

  tapDebug: () => {
    const { debugTapCount } = get();
    const next = debugTapCount + 1;
    if (next >= 7) {
      const newMode = !get().debugMode;
      localStorage.setItem("doutok_debug", String(newMode));
      set({ debugMode: newMode, debugTapCount: 0 });
    } else {
      set({ debugTapCount: next });
      // Reset tap counter after 3 seconds of inactivity
      setTimeout(() => {
        if (get().debugTapCount === next) {
          set({ debugTapCount: 0 });
        }
      }, 3000);
    }
  },

  clearError: () => set({ error: null }),

  restoreSession: async () => {
    const token = localStorage.getItem("doutok_token");
    if (token) {
      set({ token });
      await get().fetchMe();
    }
  },
}));
