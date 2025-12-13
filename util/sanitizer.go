package util

import (
	"net/url"
	"strings"
)

// MaskDBURL takes a database URL and returns a new string with the password masked.
func MaskDBURL(dbURL string) string {
	// Handle cases where the URL might be empty or not a valid URL
	if dbURL == "" {
		return ""
	}

	// For simple DSNs that are not full URLs (like just a filename)
	if !strings.Contains(dbURL, "://") {
		return dbURL // It's likely a simple file path like "forge.db"
	}

	parsedURL, err := url.Parse(dbURL)
	if err != nil {
		// If parsing fails, return a generic masked message to avoid leaking info
		return "invalid-db-url-format"
	}

	if parsedURL.User != nil {
		if _, isSet := parsedURL.User.Password(); isSet {
			// Create a new Userinfo with a masked password
			newUserInfo := url.UserPassword(parsedURL.User.Username(), "******")
			parsedURL.User = newUserInfo
			return parsedURL.String()
		}
	}

	// Return the original URL if no password is set
	return dbURL
}
