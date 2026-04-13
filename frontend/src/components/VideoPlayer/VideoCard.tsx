import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import {
  Heart,
  MessageCircle,
  Share2,
  Music,
  Plus,
} from "lucide-react";
import type { Video } from "../../stores/feedStore";
import { useFeedStore } from "../../stores/feedStore";
import "./VideoCard.css";

interface Props {
  video: Video;
  isActive: boolean;
}

export default function VideoCard({ video, isActive }: Props) {
  const navigate = useNavigate();
  const toggleLike = useFeedStore((s) => s.toggleLike);
  const [showHeart, setShowHeart] = useState(false);
  const lastTap = useRef(0);

  const handleDoubleTap = () => {
    const now = Date.now();
    if (now - lastTap.current < 300) {
      if (!video.is_liked) {
        toggleLike(video.id);
      }
      setShowHeart(true);
      setTimeout(() => setShowHeart(false), 800);
    }
    lastTap.current = now;
  };

  const formatCount = (n: number): string => {
    if (n >= 10000) return (n / 10000).toFixed(1) + "w";
    if (n >= 1000) return (n / 1000).toFixed(1) + "k";
    return String(n);
  };

  return (
    <div className="video-card" onClick={handleDoubleTap}>
      {/* Video / Cover placeholder */}
      <div
        className="video-bg"
        style={{ backgroundImage: `url(${video.cover_url})` }}
      />

      {/* Double tap heart animation */}
      {showHeart && (
        <div className="double-tap-heart">
          <Heart size={80} fill="#FE2C55" color="#FE2C55" />
        </div>
      )}

      {/* Right sidebar actions */}
      <div className="video-sidebar">
        {/* Avatar */}
        <div className="sidebar-avatar" onClick={(e) => { e.stopPropagation(); navigate(`/profile/${video.author_id}`); }}>
          <img src={video.author.avatar} alt="" />
          {!video.is_following && (
            <div className="avatar-follow-btn">
              <Plus size={12} />
            </div>
          )}
        </div>

        {/* Like */}
        <button
          className={`sidebar-btn ${video.is_liked ? "liked" : ""}`}
          onClick={(e) => { e.stopPropagation(); toggleLike(video.id); }}
        >
          <Heart
            size={28}
            fill={video.is_liked ? "#FE2C55" : "none"}
            color={video.is_liked ? "#FE2C55" : "white"}
          />
          <span>{formatCount(video.like_count)}</span>
        </button>

        {/* Comment */}
        <button className="sidebar-btn" onClick={(e) => { e.stopPropagation(); }}>
          <MessageCircle size={28} />
          <span>{formatCount(video.comment_count)}</span>
        </button>

        {/* Share */}
        <button className="sidebar-btn">
          <Share2 size={28} />
          <span>{formatCount(video.share_count)}</span>
        </button>

        {/* Music disc */}
        <div className={`sidebar-disc ${isActive ? "spinning" : ""}`}>
          <img src={video.author.avatar} alt="" />
        </div>
      </div>

      {/* Bottom info */}
      <div className="video-info">
        <div className="video-author">@{video.author.nickname}</div>
        <div className="video-desc ellipsis-2">{video.description}</div>
        <div className="video-music">
          <Music size={14} />
          <span className="music-marquee">原声 - {video.author.nickname}</span>
        </div>
      </div>
    </div>
  );
}
