package tracking

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Redact replaces the user's home folder in text with "~", so file paths in
// tracked error messages don't carry the user's account name.
func Redact(text string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return text
	}
	return redactHome(text, home, runtime.GOOS == "windows")
}

// redactHome replaces home in text, matching either path separator between
// its parts, and ignoring case on Windows.
func redactHome(text, home string, windows bool) string {
	home = strings.TrimRight(filepath.ToSlash(home), "/")
	parts := strings.Split(home, "/")
	if len(parts) < 2 {
		return text
	}
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	// The home folder ends at a separator, punctuation or the end, so
	// C:\Users\Jan doesn't match inside C:\Users\Janet.
	pattern := strings.Join(parts, `[\\/]`) + `([^\w.-]|$)`
	if windows {
		pattern = "(?i)" + pattern
	}
	return regexp.MustCompile(pattern).ReplaceAllString(text, "~${1}")
}
