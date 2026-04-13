import { useState, useRef, useCallback, useEffect } from "react";
import { useFeedStore } from "../../stores/feedStore";
import type { Video } from "../../stores/feedStore";
import VideoCard from "../../components/VideoPlayer/VideoCard";
import "./Feed.css";

// Mock data for development
const mockVideos: Video[] = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1,
  author_id: i + 100,
  author: {
    id: i + 100,
    nickname: `creator_${i + 1}`,
    avatar: `https://picsum.photos/seed/avatar${i}/100/100`,
  },
  title: `Video ${i + 1}`,
  description: `This is video #${i + 1} description #douyin #test ${i % 2 === 0 ? "#trending" : "#daily"}`,
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
  const deltaY = useRef(0);
  const isDragging = useRef(false);
  const [dragOffset, setDragOffset] = useState(0);
  const [isTransitioning, setIsTransitioning] = useState(false);

  useEffect(() => {
    if (videos.length === 0) {
      setVideos(mockVideos);
    }
  }, []);

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (isTransitioning) return;
    startY.current = e.touches[0].clientY;
    deltaY.current = 0;
    isDragging.current = true;
  }, [isTransitioning]);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isDragging.current || isTransitioning) return;
    deltaY.current = e.touches[0].clientY - startY.current;

    // Clamp: don't drag beyond first/last
    if (currentIndex === 0 && deltaY.current > 0) {
      deltaY.current = deltaY.current * 0.3; // rubber band
    }
    if (currentIndex === videos.length - 1 && deltaY.current < 0) {
      deltaY.current = deltaY.current * 0.3;
    }

    setDragOffset(deltaY.current);
  }, [currentIndex, videos.length, isTransitioning]);

  const handleTouchEnd = useCallback(() => {
    if (!isDragging.current) return;
    isDragging.current = false;
    const threshold = 80;

    setIsTransitioning(true);
    setDragOffset(0);

    if (deltaY.current < -threshold && currentIndex < videos.length - 1) {
      setCurrentIndex(currentIndex + 1);
    } else if (deltaY.current > threshold && currentIndex > 0) {
      setCurrentIndex(currentIndex - 1);
    }

    // Reset transition lock
    setTimeout(() => setIsTransitioning(false), 400);
  }, [currentIndex, videos.length]);

  // Mouse wheel support for desktop
  const wheelTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);
  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (isTransitioning) return;
    if (wheelTimeout.current) return; // debounce

    wheelTimeout.current = setTimeout(() => {
      wheelTimeout.current = null;
    }, 600);

    if (e.deltaY > 30 && currentIndex < videos.length - 1) {
      setIsTransitioning(true);
      setCurrentIndex(currentIndex + 1);
      setTimeout(() => setIsTransitioning(false), 400);
    } else if (e.deltaY < -30 && currentIndex > 0) {
      setIsTransitioning(true);
      setCurrentIndex(currentIndex - 1);
      setTimeout(() => setIsTransitioning(false), 400);
    }
  }, [currentIndex, videos.length, isTransitioning]);

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

      {/* Video swiper - only render 3 videos: prev, current, next */}
      <div
        ref={containerRef}
        className="feed-swiper"
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onWheel={handleWheel}
      >
        <div
          className="feed-track"
          style={{
            transform: `translateY(calc(${-currentIndex * 100}% + ${dragOffset}px))`,
            transition: isDragging.current ? "none" : "transform 0.4s cubic-bezier(0.16, 1, 0.3, 1)",
          }}
        >
          {videos.map((video, index) => {
            // Only render ±2 from current for performance
            if (Math.abs(index - currentIndex) > 2) {
              return <div key={video.id} className="feed-slide" />;
            }
            return (
              <div key={video.id} className="feed-slide">
                <VideoCard video={video} isActive={index === currentIndex} />
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
