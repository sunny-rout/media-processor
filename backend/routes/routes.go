package routes

import (
	"backend/utils"
	"net/http"

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
