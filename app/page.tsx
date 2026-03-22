"use client";

import { useState } from "react";

export default function Home() {
  const [url, setUrl] = useState("");
  const [format, setFormat] = useState<"video" | "audio">("video");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleDownload = async () => {
    if (!url.trim()) {
      setError("Please enter a YouTube URL");
      return;
    }

    setError("");
    setIsLoading(true);

    try {
      const backendUrl = `http://localhost:8080/api/stream?url=${encodeURIComponent(url)}&format=${format}`;

      const response = await fetch(backendUrl);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Failed to download media");
      }

      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = downloadUrl;
      a.download = format === "video" ? "media.mp4" : "media.mp3";
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(downloadUrl);
      document.body.removeChild(a);

      setUrl("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 px-4">
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

          <div className="space-y-4">
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
                onChange={(e) => setUrl(e.target.value)}
                placeholder="https://www.youtube.com/watch?v=..."
                className="w-full px-4 py-3 rounded-lg border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:focus:ring-blue-400 focus:border-transparent transition"
                disabled={isLoading}
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

            {error && (
              <div className="p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
                <p className="text-sm text-red-600 dark:text-red-400">
                  {error}
                </p>
              </div>
            )}

            <button
              onClick={handleDownload}
              disabled={isLoading}
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
                    ></circle>
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    ></path>
                  </svg>
                  <span>Processing...</span>
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
          </div>

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
