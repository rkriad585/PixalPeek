package qrengine

import (
	"net/url"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isPhoneNumber(s string) bool {
	c := s
	if strings.HasPrefix(c, "+") {
		c = c[1:]
	}
	if len(c) < 5 || len(c) > 16 {
		return false
	}
	digits := 0
	for _, r := range c {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == ' ' || r == '-' || r == '(' || r == ')' || r == '.':
		default:
			return false
		}
	}
	return digits >= 5
}

func DetectContentType(content string) ContentType {
	c := strings.TrimSpace(content)
	if c == "" {
		return TypeText
	}
	upper := strings.ToUpper(c)
	switch {
	case strings.HasPrefix(upper, "WIFI:"):
		return TypeWiFi
	case strings.HasPrefix(upper, "BEGIN:VCARD"):
		return TypeVCard
	case strings.HasPrefix(upper, "BEGIN:VEVENT"):
		return TypeEvent
	case strings.HasPrefix(upper, "GEO:"):
		return TypeGeo
	case strings.HasPrefix(upper, "MATMSG:") || strings.HasPrefix(upper, "MAILTO:"):
		return TypeEmail
	case strings.HasPrefix(upper, "SMSTO:") || strings.HasPrefix(upper, "SMS:"):
		return TypeSMS
	case strings.HasPrefix(upper, "TEL:"):
		return TypePhone
	case strings.HasPrefix(upper, "HTTP://"), strings.HasPrefix(upper, "HTTPS://"):
		return TypeURL
	}

	if strings.HasPrefix(strings.ToLower(c), "www.") {
		return TypeURL
	}
	if u, err := url.ParseRequestURI(c); err == nil && u.Scheme != "" && u.Host != "" {
		return TypeURL
	}
	if emailRegex.MatchString(c) {
		return TypeEmail
	}
	if isPhoneNumber(c) {
		return TypePhone
	}
	return TypeText
}
