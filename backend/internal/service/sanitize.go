package service

import "regexp"

var (
	ipRegex           = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\b`)
	secretKVRegex     = regexp.MustCompile(`(?i)\b(password|passwd|token|secret|api[_-]?key)\b\s*[:=]\s*([^\s,;]+)`)
	bearerTokenRegex  = regexp.MustCompile(`(?i)bearer\s+([a-z0-9\-._~+/]+=*)`)
	basicAuthURLRegex = regexp.MustCompile(`(?i)(https?://[^\s:@]+:)([^\s@]+)(@)`)
)

func sanitizeSensitiveText(input string) string {
	if input == "" {
		return input
	}
	result := ipRegex.ReplaceAllString(input, "x.x.x.x")
	result = secretKVRegex.ReplaceAllString(result, "$1=***")
	result = bearerTokenRegex.ReplaceAllString(result, "Bearer ***")
	result = basicAuthURLRegex.ReplaceAllString(result, "$1***$3")
	return result
}
