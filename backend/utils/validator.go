package utils

import (
	"net/url"
	"strings"
)

type ValidationResult struct {
	Valid    bool   `json:"valid"`
	Platform string `json:"platform"`
}

var dangerousChars = []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r", "\\"}

func containsDangerousChars(s string) bool {
	for _, ch := range dangerousChars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}

func ValidateURL(urlStr string) ValidationResult {
	invalid := ValidationResult{Valid: false, Platform: "unknown"}

	if urlStr == "" || len(urlStr) > 2048 {
		return invalid
	}

	if containsDangerousChars(urlStr) {
		return invalid
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return invalid
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return invalid
	}

	if parsedURL.Host == "" {
		return invalid
	}

	host := strings.ToLower(parsedURL.Host)
	host = strings.TrimPrefix(host, "www.")

	if strings.Contains(host, "..") || strings.ContainsAny(host, " \t") {
		return invalid
	}

	switch {
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com"):
		return ValidationResult{Valid: true, Platform: "youtube"}
	case host == "youtu.be":
		return ValidationResult{Valid: true, Platform: "youtube"}
	case host == "instagram.com" || strings.HasSuffix(host, ".instagram.com"):
		return ValidationResult{Valid: true, Platform: "instagram"}
	default:
		return invalid
	}
}
