package qrengine

import (
	"net"
	"net/url"
	"strings"
)

type SafetyAssessment struct {
	IsSuspicious bool     `json:"is_suspicious"`
	Score        int      `json:"score"`
	Reasons      []string `json:"reasons"`
	ResolvedHost string   `json:"resolved_host"`
	IsShortener  bool     `json:"is_shortener"`
}

var knownShorteners = map[string]bool{
	"bit.ly": true, "tinyurl.com": true, "t.co": true, "goo.gl": true,
	"is.gd": true, "buff.ly": true, "ow.ly": true, "cutt.ly": true,
	"rb.gy": true, "shorturl.at": true, "tiny.cc": true, "v.gd": true,
	"lnkd.in": true, "rebrand.ly": true, "bl.ink": true, "t.ly": true,
	"dub.sh": true, "surl.li": true,
}

var suspiciousTLDs = []string{
	".zip", ".mov", ".top", ".gq", ".tk", ".cf", ".buzz", ".xyz",
	".click", ".link", ".loan", ".racing", ".review", ".stream",
	".win", ".bid", ".download", ".accountant", ".science", ".party",
}

func AssessURLSecurity(rawURL string) SafetyAssessment {
	assessment := SafetyAssessment{
		IsSuspicious: false,
		Score:        0,
		Reasons:      make([]string, 0),
	}

	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return assessment
	}

	hostname := strings.ToLower(u.Hostname())
	assessment.ResolvedHost = hostname

	if knownShorteners[hostname] {
		assessment.IsShortener = true
		assessment.Score += 35
		assessment.Reasons = append(assessment.Reasons, "obfuscated target: uses URL shortener domain ("+hostname+")")
	}

	if net.ParseIP(hostname) != nil {
		assessment.Score += 50
		assessment.Reasons = append(assessment.Reasons, "suspicious direct IP address instead of domain name")
	}

	if strings.HasPrefix(hostname, "xn--") || strings.Contains(hostname, ".xn--") {
		assessment.Score += 40
		assessment.Reasons = append(assessment.Reasons, "punycode/IDN character spoofing detected")
	}

	subdomainCount := len(strings.Split(hostname, "."))
	if subdomainCount > 4 {
		assessment.Score += 25
		assessment.Reasons = append(assessment.Reasons, "high subdomain depth (possible masquerading)")
	}

	if u.Scheme == "http" {
		query := strings.ToLower(u.RawQuery)
		if strings.Contains(query, "token") || strings.Contains(query, "auth") ||
			strings.Contains(query, "password") || strings.Contains(query, "session") ||
			strings.Contains(query, "key") || strings.Contains(query, "secret") {
			assessment.Score += 45
			assessment.Reasons = append(assessment.Reasons, "insecure HTTP scheme transmitting sensitive auth parameters")
		}
	}

	for _, tld := range suspiciousTLDs {
		if strings.HasSuffix(hostname, tld) {
			assessment.Score += 20
			assessment.Reasons = append(assessment.Reasons, "high-risk top-level domain ("+tld+")")
			break
		}
	}

	if strings.Contains(u.RawQuery, "@") || strings.Contains(hostname, "@") {
		assessment.Score += 30
		assessment.Reasons = append(assessment.Reasons, "contains @ in URL (possible credential phishing)")
	}

	if strings.Count(hostname, "-") > 3 && subdomainCount <= 2 {
		assessment.Score += 15
		assessment.Reasons = append(assessment.Reasons, "excessive hyphens in domain (possible typosquatting)")
	}

	if assessment.Score >= 35 {
		assessment.IsSuspicious = true
	}

	return assessment
}

func AssessURLSecurityForResults(results []DecodeResult) []DecodeResult {
	for i := range results {
		if results[i].ContentType == TypeURL {
			assessment := AssessURLSecurity(results[i].Content)
			results[i].Security = &assessment
		}
	}
	return results
}
