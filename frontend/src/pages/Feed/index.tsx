import { useState, useRef, useCallback, useEffect } from "react";
import { useFeedStore } from "../../stores/feedStore";
import type { Video } from "../../stores/feedStore";
import VideoCard from "../../components/VideoPlayer/VideoCard";
import "./Feed.css";

// Mock data for development
const mockVideos: Video[] = Array.from({ length: 10 }, (_, i) => ({
  id: i + 1,
  author_id: i + 100,
  author: {
    id: i + 100,
    nickname: `creator_${i + 1}`,
    avatar: `https://picsum.photos/seed/avatar${i}/100/100`,
  },
  title: `视频标题 ${i + 1}`,
  description: `这是第 ${i + 1} 个视频的描述 #抖音 #测试 ${i % 2 === 0 ? "#热门" : "#日常"}`,
  cover_url: `https://picsum.photos/seed/cover${i}/720/1280`,
  play_url: "",
  duration: 15 + i * 3,
  like_count: Math.floor(Math.random() * 100000),
  comment_count: Math.floor(Math.random() * 5000),
  share_count: Math.floor(Math.random() * 2000),
  is_liked: false,
  is_following: false,
}));

export default function Feed() {
  const { videos, currentIndex, setVideos, setCurrentIndex } = useFeedStore();
  const [activeTab, setActiveTab] = useState<"following" | "recommend">("recommend");
  const containerRef = useRef<HTMLDivElement>(null);
  const startY = useRef(0);
  const currentY = useRef(0);
  const isDragging = useRef(false);

  useEffect(() => {
    if (videos.length === 0) {
      setVideos(mockVideos);
    }
  }, []);

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    startY.current = e.touches[0].clientY;
    isDragging.current = true;
  }, []);

  const handleTouchEnd = useCallback(
    (e: React.TouchEvent) => {
      if (!isDragging.current) return;
      isDragging.current = false;
      const diff = startY.current - e.changedTouches[0].clientY;
      const threshold = 80;

      if (diff > threshold && currentIndex < videos.length - 1) {
        setCurrentIndex(currentIndex + 1);
      } else if (diff < -threshold && currentIndex > 0) {
        setCurrentIndex(currentIndex - 1);
      }
    },
    [currentIndex, videos.length]
  );

  return (
    <div className="feed-page">
      {/* Top tabs */}
      <div className="feed-header safe-top">
        <button
          className={`feed-tab ${activeTab === "following" ? "active" : ""}`}
          onClick={() => setActiveTab("following")}
        >
          关注
        </button>
        <div className="feed-tab-divider" />
        <button
          className={`feed-tab ${activeTab === "recommend" ? "active" : ""}`}
          onClick={() => setActiveTab("recommend")}
        >
          推荐
        </button>
      </div>

      {/* Video swiper */}
      <div
        ref={containerRef}
        className="feed-swiper"
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
      >
        <div
          className="feed-track"
          style={{
            transform: `translateY(-${currentIndex * 100}%)`,
            transition: "transform 0.35s cubic-bezier(0.25, 0.46, 0.45, 0.94)",
          }}
        >
          {videos.map((video, index) => (
            <VideoCard
              key={video.id}
              video={video}
              isActive={index === currentIndex}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
