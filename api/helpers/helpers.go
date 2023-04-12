package helpers

import (
	"os"
	"strings"
)

func EnforceHttp(url string) string {
	if url[:4] != "http" {
		return "https://" + url[4:]
	}
	return url
}

func RemoveDomainError(url string) bool {
	domain := os.Getenv("DOMAIN")

	if url == domain {
		return false
	}

	validUrl := strings.Replace(url, "http://", "", 1)
	validUrl = strings.Replace(validUrl, "https://", "", 1)
	validUrl = strings.Replace(validUrl, "www.", "", 1)
	validUrl = strings.Split(validUrl, "/")[0]

	if validUrl == domain {
		return false
	}

	return true
}