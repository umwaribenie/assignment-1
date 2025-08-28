package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// GenerateSlug generates a URL-friendly slug from first name and last name
func GenerateSlug(firstName, lastName string) string {
	// Combine first and last name
	fullName := strings.ToLower(firstName + " " + lastName)
	
	// Replace spaces with hyphens
	slug := strings.ReplaceAll(fullName, " ", "-")
	
	// Remove special characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	slug = reg.ReplaceAllString(slug, "")
	
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	
	// Add timestamp to make it unique
	timestamp := time.Now().Unix()
	slug = fmt.Sprintf("%s-%d", slug, timestamp)
	
	return slug
}