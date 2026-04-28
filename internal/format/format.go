// Package format provides terminal formatting helpers (colors, sizing, truncation).
package format

import "fmt"

// ANSI color/style codes for terminal output.
const (
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Cyan    = "\033[36m"
	Red     = "\033[31m"
	Magenta = "\033[35m"
	Reset   = "\033[0m"
)

// HumanBytes converts a byte count into a human-readable string (e.g. "1.4 MB").
func HumanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Truncate shortens a string to n characters, appending "…" if truncated.
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// TypeIcon returns a unicode icon for a content type.
func TypeIcon(contentType string) string {
	switch contentType {
	case "image":
		return "🖼️"
	case "video":
		return "🎬"
	case "audio":
		return "🎵"
	case "pdf":
		return "📄"
	case "text":
		return "📝"
	default:
		return "📦"
	}
}
