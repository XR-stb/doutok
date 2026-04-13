import { create } from "zustand";

export interface Video {
  id: number;
  author_id: number;
  author: {
    id: number;
    nickname: string;
    avatar: string;
  };
  title: string;
  description: string;
  cover_url: string;
  play_url: string;
  duration: number;
  like_count: number;
  comment_count: number;
  share_count: number;
  is_liked: boolean;
  is_following: boolean;
}

interface FeedState {
  videos: Video[];
  currentIndex: number;
  loading: boolean;
  setVideos: (videos: Video[]) => void;
  appendVideos: (videos: Video[]) => void;
  setCurrentIndex: (index: number) => void;
  setLoading: (loading: boolean) => void;
  toggleLike: (videoId: number) => void;
}

export const useFeedStore = create<FeedState>((set) => ({
  videos: [],
  currentIndex: 0,
  loading: false,

  setVideos: (videos) => set({ videos }),
  appendVideos: (videos) =>
    set((s) => ({ videos: [...s.videos, ...videos] })),
  setCurrentIndex: (index) => set({ currentIndex: index }),
  setLoading: (loading) => set({ loading }),

  toggleLike: (videoId) =>
    set((s) => ({
      videos: s.videos.map((v) =>
        v.id === videoId
          ? {
              ...v,
              is_liked: !v.is_liked,
              like_count: v.is_liked ? v.like_count - 1 : v.like_count + 1,
            }
          : v
      ),
    })),
}));
