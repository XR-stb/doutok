import { useState, useRef, useCallback, useEffect } from "react";
import { useFeedStore } from "../../stores/feedStore";
import VideoCard from "../../components/VideoPlayer/VideoCard";
import { feedAPI } from "../../services/api";
import "./Feed.css";

export default function Feed() {
  const { videos, currentIndex, setVideos, setCurrentIndex, appendVideos } = useFeedStore();
  const [activeTab, setActiveTab] = useState<"following" | "recommend">("recommend");
  const containerRef = useRef<HTMLDivElement>(null);
  const startY = useRef(0);
  const deltaY = useRef(0);
  const isDragging = useRef(false);
  const [dragOffset, setDragOffset] = useState(0);
  const [isTransitioning, setIsTransitioning] = useState(false);

  useEffect(() => {
    if (videos.length === 0) {
      loadFeed();
    }
  }, []);

  const loadFeed = async () => {
    try {
      const res: any = await feedAPI.getFeed({ count: 10 });
      if (res.code === 0 && res.data?.videos?.length > 0) {
        setVideos(res.data.videos);
      }
    } catch {
      // API not available - show empty state
    }
  };

  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (isTransitioning) return;
    startY.current = e.touches[0].clientY;
    deltaY.current = 0;
    isDragging.current = true;
  }, [isTransitioning]);

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    if (!isDragging.current || isTransitioning) return;
    deltaY.current = e.touches[0].clientY - startY.current;

    if (currentIndex === 0 && deltaY.current > 0) {
      deltaY.current = deltaY.current * 0.3;
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
    setTimeout(() => setIsTransitioning(false), 400);
  }, [currentIndex, videos.length]);

  const wheelTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);
  const handleWheel = useCallback((e: React.WheelEvent) => {
    if (isTransitioning) return;
    if (wheelTimeout.current) return;

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

  if (videos.length === 0) {
    return (
      <div className="feed-page">
        <div className="feed-header safe-top">
          <button className={`feed-tab ${activeTab === "following" ? "active" : ""}`} onClick={() => setActiveTab("following")}>关注</button>
          <div className="feed-tab-divider" />
          <button className={`feed-tab ${activeTab === "recommend" ? "active" : ""}`} onClick={() => setActiveTab("recommend")}>推荐</button>
        </div>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "center", height: "100%", color: "rgba(255,255,255,0.4)", flexDirection: "column", gap: 12 }}>
          <p style={{ fontSize: 16 }}>No videos yet</p>
          <p style={{ fontSize: 13, opacity: 0.6 }}>Upload your first video!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="feed-page">
      <div className="feed-header safe-top">
        <button className={`feed-tab ${activeTab === "following" ? "active" : ""}`} onClick={() => setActiveTab("following")}>关注</button>
        <div className="feed-tab-divider" />
        <button className={`feed-tab ${activeTab === "recommend" ? "active" : ""}`} onClick={() => setActiveTab("recommend")}>推荐</button>
      </div>

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
