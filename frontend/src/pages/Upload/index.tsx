import { useState, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { Upload as UploadIcon, X, Film, Music, Hash, Type, ChevronRight, Loader } from "lucide-react";
import { useAuthStore } from "../../stores/authStore";
import { videoAPI } from "../../services/api";
import "./Upload.css";

export default function Upload() {
  const navigate = useNavigate();
  const { isLoggedIn } = useAuthStore();
  const fileRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string>("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [tags, setTags] = useState("");
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState("");

  if (!isLoggedIn) {
    return (
      <div className="upload-page">
        <div className="upload-empty">
          <UploadIcon size={48} strokeWidth={1.5} />
          <h3>Upload Video</h3>
          <p>Log in to share your content</p>
          <button className="upload-login-btn" onClick={() => navigate("/login")}>
            Log In
          </button>
        </div>
      </div>
    );
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selected = e.target.files?.[0];
    if (!selected) return;

    if (!selected.type.startsWith("video/")) {
      setError("Please select a video file");
      return;
    }

    if (selected.size > 500 * 1024 * 1024) {
      setError("Video must be under 500MB");
      return;
    }

    setFile(selected);
    setError("");

    // Generate preview thumbnail
    const url = URL.createObjectURL(selected);
    setPreview(url);

    // Auto-fill title from filename
    if (!title) {
      const name = selected.name.replace(/\.[^.]+$/, "").replace(/[_-]/g, " ");
      setTitle(name);
    }
  };

  const handleUpload = async () => {
    if (!file) return;
    if (!title.trim()) {
      setError("Please enter a title");
      return;
    }

    setUploading(true);
    setProgress(0);
    setError("");

    try {
      const formData = new FormData();
      formData.append("video", file);
      formData.append("title", title.trim());
      formData.append("description", description.trim());
      formData.append("tags", tags.trim());

      await videoAPI.upload(formData, {
        onUploadProgress: (e: any) => {
          if (e.total) {
            setProgress(Math.round((e.loaded / e.total) * 100));
          }
        },
      });

      // Success
      setFile(null);
      setPreview("");
      setTitle("");
      setDescription("");
      setTags("");
      setProgress(100);

      // Navigate to profile to see uploaded video
      setTimeout(() => navigate("/profile"), 1000);
    } catch (err: any) {
      setError(err?.msg || err?.message || "Upload failed. Try again.");
    } finally {
      setUploading(false);
    }
  };

  const clearFile = () => {
    setFile(null);
    setPreview("");
    setProgress(0);
    if (fileRef.current) fileRef.current.value = "";
  };

  const formatSize = (bytes: number) => {
    if (bytes >= 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    return (bytes / 1024).toFixed(0) + " KB";
  };

  return (
    <div className="upload-page">
      <div className="upload-header">
        <h2>Upload</h2>
      </div>

      <div className="upload-content">
        {/* File picker */}
        {!file ? (
          <button className="upload-dropzone" onClick={() => fileRef.current?.click()}>
            <div className="upload-dropzone-inner">
              <UploadIcon size={40} strokeWidth={1.5} />
              <h3>Select Video</h3>
              <p>MP4, MOV, WebM up to 500MB</p>
            </div>
          </button>
        ) : (
          <div className="upload-preview">
            <div className="upload-preview-video">
              {preview && (
                <video src={preview} className="upload-preview-player" muted playsInline />
              )}
              <button className="upload-preview-remove" onClick={clearFile}>
                <X size={18} />
              </button>
              <div className="upload-preview-info">
                <Film size={14} />
                <span>{file.name}</span>
                <span className="upload-preview-size">{formatSize(file.size)}</span>
              </div>
            </div>
          </div>
        )}

        <input
          ref={fileRef}
          type="file"
          accept="video/*"
          style={{ display: "none" }}
          onChange={handleFileSelect}
        />

        {/* Form */}
        <div className="upload-form">
          <div className="upload-field">
            <label>
              <Type size={14} />
              Title
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Give your video a title..."
              maxLength={128}
              className="upload-input"
            />
            <span className="upload-count">{title.length}/128</span>
          </div>

          <div className="upload-field">
            <label>
              <ChevronRight size={14} />
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Describe your video..."
              maxLength={512}
              rows={3}
              className="upload-input upload-textarea"
            />
            <span className="upload-count">{description.length}/512</span>
          </div>

          <div className="upload-field">
            <label>
              <Hash size={14} />
              Tags
            </label>
            <input
              type="text"
              value={tags}
              onChange={(e) => setTags(e.target.value)}
              placeholder="#trending #funny #music"
              className="upload-input"
            />
          </div>

          <div className="upload-field">
            <label>
              <Music size={14} />
              Visibility
            </label>
            <select className="upload-input upload-select">
              <option value="public">Public - Anyone can see</option>
              <option value="friends">Friends Only</option>
              <option value="private">Private</option>
            </select>
          </div>
        </div>

        {/* Error */}
        {error && <div className="upload-error">{error}</div>}

        {/* Progress */}
        {uploading && (
          <div className="upload-progress">
            <div className="upload-progress-bar">
              <div className="upload-progress-fill" style={{ width: `${progress}%` }} />
            </div>
            <span>{progress}%</span>
          </div>
        )}

        {/* Submit */}
        <button
          className="upload-submit"
          onClick={handleUpload}
          disabled={!file || uploading || !title.trim()}
        >
          {uploading ? (
            <>
              <Loader size={18} className="upload-spinner" />
              Uploading...
            </>
          ) : (
            <>
              <UploadIcon size={18} />
              Post Video
            </>
          )}
        </button>
      </div>
    </div>
  );
}
