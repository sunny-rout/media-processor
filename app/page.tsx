"use client";

import { useState, FormEvent } from "react";

const BACKEND_URL = "http://localhost:8080";

const validateYouTubeUrl = (url: string): boolean => {
  const patterns = [
    /^(https?:\/\/)?(www\.)?(youtube\.com|youtu\.be)\/.+$/i,
    /^(https?:\/\/)?(www\.)?youtube\.com\/watch\?v=.+$/i,
    /^(https?:\/\/)?(www\.)?youtu\.be\/.+$/i,
  ];
  return patterns.some((pattern) => pattern.test(url));
};

export default function Home() {
  const [url, setUrl] = useState("");
  const [format, setFormat] = useState<"video" | "audio">("video");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [downloadProgress, setDownloadProgress] = useState(0);

  const handleDownload = async (e: FormEvent) => {
    e.preventDefault();

    const trimmedUrl = url.trim();

    if (!trimmedUrl) {
      setError("Please enter a YouTube URL");
      return;
    }

    if (!validateYouTubeUrl(trimmedUrl)) {
      setError("Please enter a valid YouTube or Youtu.be URL");
      return;
    }

    setError("");
    setIsLoading(true);

    try {
      const streamUrl = `${BACKEND_URL}/api/stream?url=${encodeURIComponent(trimmedUrl)}&format=${format}`;

      const response = await fetch(streamUrl, {
        method: "GET",
        headers: {
          "Accept": format === "video" ? "video/mp4" : "audio/mpeg",
        },
      });

      if (!response.ok) {
        const contentType = response.headers.get("content-type");
        if (contentType && contentType.includes("application/json")) {
          const errorData = await response.json();
          throw new Error(errorData.error || "Failed to download media");
        }
        throw new Error(`Server error: ${response.status}`);
      }

      const contentLength = response.headers.get("Content-Length");
      const total = contentLength ? parseInt(contentLength, 10) : 0;
      const contentDisposition = response.headers.get("Content-Disposition");
      const filenameMatch = contentDisposition?.match(/filename="(.+)"/);
      const filename = filenameMatch ? filenameMatch[1] : (format === "video" ? "video.mp4" : "audio.mp3");

      if (!response.body) {
        throw new Error("Streaming not supported in this browser");
      }

      const reader = response.body.getReader();
      const chunks: BlobPart[] = [];
      let received = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
        received += value.length;
        if (total > 0) {
          const progress = Math.round((received / total) * 100);
          setDownloadProgress(progress);
        }
      }

      const blob = new Blob(chunks, {
        type: format === "video" ? "video/mp4" : "audio/mpeg",
      });
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = downloadUrl;
      link.download = filename;
      setDownloadProgress(0);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);

      setUrl("");
      setDownloadProgress(0);
    } catch (err) {
      setDownloadProgress(0);
      if (err instanceof TypeError && err.message.includes("fetch")) {
        setError("Cannot connect to backend. Make sure the server is running on port 8080");
      } else {
        setError(err instanceof Error ? err.message : "An unexpected error occurred");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && !isLoading) {
      handleDownload(e as unknown as FormEvent);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 px-4 py-8">
      <div className="w-full max-w-md">
        <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 space-y-6">
          <div className="text-center space-y-2">
            <h1 className="text-3xl font-bold text-slate-900 dark:text-white">
              YouTube Downloader
            </h1>
            <p className="text-slate-600 dark:text-slate-400">
              Download videos and audio from YouTube
            </p>
          </div>

          <form onSubmit={handleDownload} className="space-y-4">
            <div>
              <label
                htmlFor="url"
                className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
              >
                YouTube URL
              </label>
              <input
                id="url"
                type="text"
                value={url}
                onChange={(e) => {
                  setUrl(e.target.value);
                  if (error) setError("");
                }}
                onKeyPress={handleKeyPress}
                placeholder="https://www.youtube.com/watch?v=..."
                className="w-full px-4 py-3 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition"
                disabled={isLoading}
                autoComplete="off"
              />
            </div>

            <div>
              <label
                htmlFor="format"
                className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2"
              >
                Format
              </label>
              <select
                id="format"
                value={format}
                onChange={(e) => setFormat(e.target.value as "video" | "audio")}
                className="w-full px-4 py-3 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition cursor-pointer"
                disabled={isLoading}
              >
                <option value="video">Video (MP4)</option>
                <option value="audio">Audio (MP3)</option>
              </select>
            </div>

            {downloadProgress > 0 && (
              <div className="w-full">
                <div className="flex justify-between text-sm text-slate-600 dark:text-slate-400 mb-1">
                  <span>Downloading...</span>
                  <span>{downloadProgress}%</span>
                </div>
                <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2">
                  <div
                    className="bg-blue-600 dark:bg-blue-500 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${downloadProgress}%` }}
                  />
                </div>
              </div>
            )}

            {error && (
              <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
                <p className="text-sm text-red-600 dark:text-red-400">
                  {error}
                </p>
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading || !url.trim()}
              className="w-full py-3 px-6 rounded-lg bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 text-white font-medium focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-slate-800 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
            >
              {isLoading ? (
                <>
                  <svg
                    className="animate-spin h-5 w-5 text-white"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  <span>Processing... {downloadProgress > 0 ? `${downloadProgress}%` : ""}</span>
                </>
              ) : (
                <>
                  <svg
                    className="h-5 w-5"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                    />
                  </svg>
                  <span>Download</span>
                </>
              )}
            </button>
          </form>

          <div className="pt-4 border-t border-slate-200 dark:border-slate-700">
            <p className="text-xs text-center text-slate-500 dark:text-slate-400">
              Supports YouTube and Youtu.be links
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
