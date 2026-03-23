package utils

import (
	"regexp"
	"strings"
	"fmt"
	"math/rand"
	"time"
)

func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.TrimSpace(slug)

	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = reg.ReplaceAllString(slug,"")

	slug = regexp.MustCompile(`[\s]+`).ReplaceAllString(slug,"-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug,"-")
	slug = strings.Trim(slug,"-")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	suffix := fmt.Sprintf("%06d",r.Intn(999999))
	slug = fmt.Sprintf("%s-%s",slug,suffix)

	return slug
}

func GenerateExcerpt(content string, maxLength int) string {
	reg := regexp.MustCompile(`<[^>]*>`)
	plain := reg.ReplaceAllString(content,"")

	reg = regexp.MustCompile(`[{}\[\]":]`)
	plain = reg.ReplaceAllString(plain," ")

	plain = regexp.MustCompile(`\s+`).ReplaceAllString(plain," ")
	plain = strings.TrimSpace(plain)

	if len(plain) <= maxLength {
		return plain
	}

	truncated := plain[:maxLength]
	lastSpace := strings.LastIndex(truncated," ")
	if lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}
