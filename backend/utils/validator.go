package utils

import (
	"net/url"
	"strings"
)

type ValidationResult struct {
	Valid    bool   `json:"valid"`
	Platform string `json:"platform"`
}

func ValidateURL(urlStr string) ValidationResult {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ValidationResult{
			Valid:    false,
			Platform: "unknown",
		}
	}

	if parsedURL.Scheme == "" {
		parsedURL, err = url.Parse("https://" + urlStr)
		if err != nil {
			return ValidationResult{
				Valid:    false,
				Platform: "unknown",
			}
		}
	}

	host := strings.ToLower(parsedURL.Host)
	host = strings.TrimPrefix(host, "www.")

	switch {
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com"):
		return ValidationResult{
			Valid:    true,
			Platform: "youtube",
		}
	case host == "youtu.be":
		return ValidationResult{
			Valid:    true,
			Platform: "youtube",
		}
	case host == "instagram.com" || strings.HasSuffix(host, ".instagram.com"):
		return ValidationResult{
			Valid:    true,
			Platform: "instagram",
		}
	default:
		return ValidationResult{
			Valid:    false,
			Platform: "unknown",
		}
	}
}
