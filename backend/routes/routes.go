package routes

import (
	"backend/metrics"
	"backend/utils"
	"backend/worker"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	streamTimeout  = 3 * time.Minute
	titleTimeout   = 30 * time.Second
	imageTimeout   = 30 * time.Second
	maxRetries     = 2
	retryDelay     = 2 * time.Second
)

var StreamPool = worker.NewPool(5)

var jsonLog = json.NewEncoder(os.Stdout)

type routeLog struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Event   string `json:"event"`
	URL     string `json:"url,omitempty"`
	Format  string `json:"format,omitempty"`
	Message string `json:"msg,omitempty"`
	Error   string `json:"error,omitempty"`
}

func logEvent(level, event, url, format, msg, errStr string) {
	jsonLog.Encode(routeLog{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   level,
		Event:   event,
		URL:     url,
		Format:  format,
		Message: msg,
		Error:   errStr,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func getContentType(format string) string {
	switch format {
	case "video":
		return "video/mp4"
	case "audio":
		return "audio/mpeg"
	case "image":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func ValidateURL(c *gin.Context) {
	urlParam := c.Query("url")
	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}
	result := utils.ValidateURL(urlParam)
	c.JSON(http.StatusOK, result)
}

func StreamMedia(c *gin.Context) {
	urlParam := c.Query("url")
	formatParam := c.Query("format")

	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	if formatParam != "video" && formatParam != "audio" && formatParam != "image" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format parameter must be 'video', 'audio', or 'image'"})
		return
	}

	result := utils.ValidateURL(urlParam)
	if !result.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or unsupported URL"})
		return
	}

	if result.Platform != "youtube" && result.Platform != "instagram" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only YouTube and Instagram URLs are supported"})
		return
	}

	if result.Platform == "instagram" && formatParam == "image" {
		handleInstagramImageStream(c, urlParam)
		return
	}

	if err := StreamPool.Acquire(); err != nil {
		metrics.Inc("stream_rejected_pool_full")
		logEvent("warn", "pool_full", urlParam, formatParam, "worker pool at capacity", "")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Server is busy, please try again shortly"})
		return
	}
	defer StreamPool.Release()

	metrics.Inc("stream_started")

	title, err := getVideoTitleWithRetry(urlParam)
	if err != nil {
		errStr := err.Error()
		logEvent("warn", "title_fetch_failed", urlParam, formatParam, "", errStr)
		if strings.Contains(errStr, "private") || strings.Contains(errStr, "login required") {
			metrics.Inc("stream_error_private")
			c.JSON(http.StatusForbidden, gin.H{"error": "This content is private or requires login"})
			return
		} else if strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "not available") {
			metrics.Inc("stream_error_unavailable")
			c.JSON(http.StatusNotFound, gin.H{"error": "Content not found or unavailable"})
			return
		}
		title = "media"
	}

	filename := sanitizeFilename(title)
	if formatParam == "audio" {
		filename += ".mp3"
	} else {
		filename += ".mp4"
	}

	ctx, cancel := context.WithTimeout(context.Background(), streamTimeout)
	defer cancel()

	ytdlpArgs := buildYtdlpArgs(result.Platform, formatParam, urlParam)
	cmd := exec.CommandContext(ctx, "yt-dlp", ytdlpArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logEvent("error", "pipe_error", urlParam, formatParam, "stdout pipe", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize streaming"})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logEvent("error", "pipe_error", urlParam, formatParam, "stderr pipe", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize streaming"})
		return
	}

	if err := cmd.Start(); err != nil {
		logEvent("error", "cmd_start_failed", urlParam, formatParam, "", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start media processing"})
		return
	}

	go func() {
		<-c.Request.Context().Done()
		logEvent("info", "client_disconnected", urlParam, formatParam, "killing process", "")
		cancel()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", getContentType(formatParam))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")

	start := time.Now()

	c.Stream(func(w io.Writer) bool {
		_, copyErr := io.Copy(w, stdout)
		if copyErr != nil {
			logEvent("error", "stream_copy_error", urlParam, formatParam, "", copyErr.Error())
		}
		return false
	})

	if err := cmd.Wait(); err != nil {
		stderrOutput, _ := io.ReadAll(stderr)
		stderrStr := string(stderrOutput)
		logEvent("error", "cmd_wait_error", urlParam, formatParam, stderrStr, err.Error())
		metrics.Inc("stream_error_process")

		if ctx.Err() == context.DeadlineExceeded {
			logEvent("warn", "stream_timeout", urlParam, formatParam, "deadline exceeded", "")
			metrics.Inc("stream_timeout")
		}
	} else {
		dur := time.Since(start).Milliseconds()
		logEvent("info", "stream_complete", urlParam, formatParam, fmt.Sprintf("duration_ms=%d", dur), "")
		metrics.Inc("stream_success")
	}
}

func buildYtdlpArgs(platform, format, url string) []string {
	var args []string
	if platform == "instagram" {
		args = append(args, "--no-check-certificates")
	}
	if format == "video" {
		args = append(args, "-f", "best", "-o", "-", url)
	} else {
		args = append(args, "-x", "--audio-format", "mp3", "-o", "-", url)
	}
	return args
}

func getVideoTitleWithRetry(url string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
			logEvent("info", "title_retry", url, "", fmt.Sprintf("attempt %d", attempt+1), "")
		}
		title, err := getVideoTitle(url)
		if err == nil {
			return title, nil
		}
		lastErr = err
		errStr := err.Error()
		if strings.Contains(errStr, "private") ||
			strings.Contains(errStr, "login required") ||
			strings.Contains(errStr, "unavailable") ||
			strings.Contains(errStr, "not available") {
			return "", err
		}
	}
	return "", lastErr
}

func getVideoTitle(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), titleTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-title", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

func handleInstagramImageStream(c *gin.Context, url string) {
	imageURL, err := extractInstagramImageURL(url)
	if err != nil {
		logEvent("error", "instagram_extract_failed", url, "image", "", err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Failed to extract image from Instagram post"})
		return
	}

	client := &http.Client{Timeout: imageTimeout}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		logEvent("error", "instagram_request_create", url, "image", "", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download image"})
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		logEvent("error", "instagram_download_failed", url, "image", "", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download image"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logEvent("warn", "instagram_image_status", url, "image", fmt.Sprintf("status=%d", resp.StatusCode), "")
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found or unavailable"})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", `attachment; filename="instagram_image.jpg"`)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-cache")

	if resp.ContentLength > 0 {
		c.Header("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	}

	if _, err = io.Copy(c.Writer, resp.Body); err != nil {
		logEvent("error", "instagram_stream_error", url, "image", "", err.Error())
		return
	}

	logEvent("info", "instagram_image_complete", url, "image", "", "")
	metrics.Inc("stream_success")
}

func extractInstagramImageURL(url string) (string, error) {
	client := &http.Client{
		Timeout: imageTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	ogImageRegex := regexp.MustCompile(`<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
	matches := ogImageRegex.FindSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find og:image meta tag")
	}

	imageURL := string(matches[1])
	if imageURL == "" {
		return "", fmt.Errorf("empty image URL")
	}

	return imageURL, nil
}
