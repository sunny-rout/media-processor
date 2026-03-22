package routes

import (
	"backend/utils"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func getContentType(format string) string {
	switch format {
	case "video":
		return "video/mp4"
	case "audio":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

func ValidateURL(c *gin.Context) {
	urlParam := c.Query("url")

	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL parameter is required",
		})
		return
	}

	result := utils.ValidateURL(urlParam)

	c.JSON(http.StatusOK, result)
}

func StreamMedia(c *gin.Context) {
	urlParam := c.Query("url")
	formatParam := c.Query("format")

	if urlParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL parameter is required",
		})
		return
	}

	if formatParam != "video" && formatParam != "audio" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format parameter must be 'video' or 'audio'",
		})
		return
	}

	result := utils.ValidateURL(urlParam)
	if !result.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid URL format",
		})
		return
	}

	if result.Platform != "youtube" && result.Platform != "instagram" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only YouTube and Instagram URLs are supported",
		})
		return
	}

	// Get video title for filename and validate accessibility
	title, err := getVideoTitle(urlParam)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "private") || strings.Contains(errStr, "login required") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "This content is private or requires login",
			})
			return
		} else if strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "not available") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Content not found or unavailable",
			})
			return
		}
		log.Printf("Warning: could not get video title: %v", err)
		title = "media"
	}

	filename := sanitizeFilename(title)
	if formatParam == "audio" {
		filename += ".mp3"
	} else {
		filename += ".mp4"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	var ytdlpArgs []string

	if result.Platform == "instagram" {
		ytdlpArgs = []string{"--no-check-certificates"}
		if formatParam == "video" {
			ytdlpArgs = append(ytdlpArgs, "-f", "best", "-o", "-", urlParam)
		} else {
			ytdlpArgs = append(ytdlpArgs, "-x", "--audio-format", "mp3", "-o", "-", urlParam)
		}
	} else {
		if formatParam == "video" {
			ytdlpArgs = []string{"-f", "best", "-o", "-", urlParam}
		} else {
			ytdlpArgs = []string{"-x", "--audio-format", "mp3", "-o", "-", urlParam}
		}
	}

	cmd = exec.CommandContext(ctx, "yt-dlp", ytdlpArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating stdout pipe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize streaming",
		})
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Error creating stderr pipe: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize streaming",
		})
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting yt-dlp: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start media processing",
		})
		return
	}

	go func() {
		<-c.Request.Context().Done()
		log.Println("Client disconnected, cancelling yt-dlp process")
		cancel()
		cmd.Process.Kill()
	}()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Type", getContentType(formatParam))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("Cache-Control", "no-cache")

	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, stdout)
		if err != nil {
			log.Printf("Error streaming data: %v", err)
			return false
		}
		return false
	})

	if err := cmd.Wait(); err != nil {
		stderrOutput, _ := io.ReadAll(stderr)
		stderrStr := string(stderrOutput)
		log.Printf("yt-dlp error: %v, stderr: %s", err, stderrStr)

		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Process timeout exceeded")
		}

		if strings.Contains(stderrStr, "private") || strings.Contains(stderrStr, "login required") {
			log.Println("Private content detected")
		} else if strings.Contains(stderrStr, "Video unavailable") || strings.Contains(stderrStr, "not available") {
			log.Println("Content unavailable")
		}
	}

	log.Printf("Successfully streamed %s for URL: %s", formatParam, urlParam)
}

func getVideoTitle(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp", "--get-title", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func sanitizeFilename(name string) string {
	// Remove characters that are unsafe for filenames
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
