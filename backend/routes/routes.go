package routes

import (
	"backend/utils"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
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
	if !result.Valid || result.Platform != "youtube" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only valid YouTube URLs are supported",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	var filename string

	if formatParam == "video" {
		cmd = exec.CommandContext(ctx, "yt-dlp", "-f", "best", "-o", "-", urlParam)
		filename = "media.mp4"
	} else {
		cmd = exec.CommandContext(ctx, "yt-dlp", "-x", "--audio-format", "mp3", "-o", "-", urlParam)
		filename = "media.mp3"
	}

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
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Transfer-Encoding", "chunked")

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
		log.Printf("yt-dlp error: %v, stderr: %s", err, string(stderrOutput))

		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Process timeout exceeded")
		}
	}

	log.Printf("Successfully streamed %s for URL: %s", formatParam, urlParam)
}
