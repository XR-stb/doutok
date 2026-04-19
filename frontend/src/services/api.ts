import axios from "axios";
import type { AxiosRequestConfig } from "axios";

// Restore custom API URL if previously set in Debug panel
const savedBaseUrl = localStorage.getItem("doutok_api_url");
const baseURL = savedBaseUrl
  ? savedBaseUrl.endsWith("/api/v1")
    ? savedBaseUrl
    : `${savedBaseUrl}/api/v1`
  : `${window.location.protocol}//${window.location.hostname}:8080/api/v1`;

const api = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor: attach JWT from localStorage
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("doutok_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: unwrap axios response, handle errors
api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    // Return the backend error body if available
    if (err.response?.data) {
      return Promise.reject(err.response.data);
    }
    return Promise.reject({ code: -1, msg: err.message || "Network error" });
  }
);

// === Auth ===
export const authAPI = {
  register: (data: { username: string; password: string; nickname: string }) =>
    api.post("/auth/register", data),
  login: (data: { username: string; password: string }) =>
    api.post("/auth/login", data),
  refresh: () => api.post("/auth/refresh"),
};

// === Feed ===
export const feedAPI = {
  getFeed: (params?: { cursor?: number; count?: number }) =>
    api.get("/feed", { params }),
};

// === Video ===
export const videoAPI = {
  getVideo: (id: number) => api.get(`/videos/${id}`),
  upload: (formData: FormData, config?: AxiosRequestConfig) =>
    api.post("/videos", formData, {
      headers: { "Content-Type": "multipart/form-data" },
      timeout: 300000, // 5 min for large uploads
      ...config,
    }),
  delete: (id: number) => api.delete(`/videos/${id}`),
  like: (id: number) => api.post(`/videos/${id}/like`),
  unlike: (id: number) => api.delete(`/videos/${id}/like`),
};

// === Comment ===
export const commentAPI = {
  getComments: (
    videoId: number,
    params?: { cursor?: number; count?: number }
  ) => api.get(`/videos/${videoId}/comments`, { params }),
  create: (
    videoId: number,
    data: { content: string; parent_id?: number }
  ) => api.post(`/videos/${videoId}/comments`, data),
  delete: (id: number) => api.delete(`/comments/${id}`),
  like: (id: number) => api.post(`/comments/${id}/like`),
};

// === User ===
export const userAPI = {
  getMe: () => api.get("/me"),
  updateMe: (
    data: Partial<{ nickname: string; avatar: string; bio: string }>
  ) => api.put("/me", data),
  getProfile: (id: number) => api.get(`/users/${id}`),
};

// === Social ===
export const socialAPI = {
  follow: (id: number) => api.post(`/users/${id}/follow`),
  unfollow: (id: number) => api.delete(`/users/${id}/follow`),
  getFollowing: () => api.get("/me/following"),
  getFollowers: () => api.get("/me/followers"),
};

// === Live ===
export const liveAPI = {
  listRooms: () => api.get("/lives"),
  getRoom: (id: number) => api.get(`/lives/${id}`),
  createRoom: (data: { title: string }) => api.post("/lives", data),
  getRank: (id: number) => api.get(`/lives/${id}/rank`),
  sendGift: (
    id: number,
    data: { gift_id: number; count: number }
  ) => api.post(`/lives/${id}/gift`, data),
  like: (id: number) => api.post(`/lives/${id}/like`),
};

// === Chat ===
export const chatAPI = {
  listConversations: () => api.get("/conversations"),
  createConversation: (userId: number) =>
    api.post("/conversations", { user_id: userId }),
  getMessages: (convId: number, params?: { cursor?: number }) =>
    api.get(`/conversations/${convId}/messages`, { params }),
  sendMessage: (
    convId: number,
    data: { content: string; type?: string }
  ) => api.post(`/conversations/${convId}/messages`, data),
};

// === Search ===
export const searchAPI = {
  search: (params: { q: string; type?: string; cursor?: number }) =>
    api.get("/search", { params }),
};

export default api;
