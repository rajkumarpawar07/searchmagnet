// Package embedder handles multimodal content preparation and Gemini embedding generation.
package embedder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

// MaxFileSize is the maximum file size the Gemini embedding model accepts (20 MB).
const MaxFileSize = 20 * 1024 * 1024 // 20 MB

// FileTypeInfo holds MIME and content type for a file extension.
type FileTypeInfo struct {
	MIME        string
	ContentType string
}

// SupportedTypes maps file extensions to their type information.
var SupportedTypes = map[string]FileTypeInfo{
	// Images
	".png":  {"image/png", "image"},
	".jpg":  {"image/jpeg", "image"},
	".jpeg": {"image/jpeg", "image"},
	".gif":  {"image/gif", "image"},
	".webp": {"image/webp", "image"},
	".bmp":  {"image/bmp", "image"},
	".tiff": {"image/tiff", "image"},
	// Video
	".mp4":  {"video/mp4", "video"},
	".mov":  {"video/quicktime", "video"},
	".avi":  {"video/avi", "video"},
	".mkv":  {"video/x-matroska", "video"},
	".webm": {"video/webm", "video"},
	// Audio
	".mp3":  {"audio/mpeg", "audio"},
	".wav":  {"audio/wav", "audio"},
	".ogg":  {"audio/ogg", "audio"},
	".flac": {"audio/flac", "audio"},
	".aac":  {"audio/aac", "audio"},
	// Documents
	".pdf": {"application/pdf", "pdf"},
}

// DetectType returns the MIME type and content type for a file based on its extension.
func DetectType(path string) (mimeType, contentType string, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	info, ok := SupportedTypes[ext]
	if !ok {
		return "", "", fmt.Errorf("unsupported file extension %q", ext)
	}
	return info.MIME, info.ContentType, nil
}

// IsSupportedExt checks if a file's extension is in the supported types map.
func IsSupportedExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := SupportedTypes[ext]
	return ok
}

// FileSize returns the size of a file in bytes.
func FileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ContentsFromFile creates genai.Content from a file path.
// Returns an error if the file exceeds MaxFileSize (20 MB).
func ContentsFromFile(path, mimeType, label string) ([]*genai.Content, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}
	if info.Size() > MaxFileSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d bytes / 20 MB)", info.Size(), MaxFileSize)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Label as a separate text Part (protobuf oneof — Text and InlineData can't share a Part)
	parts := []*genai.Part{
		{Text: label},
		{InlineData: &genai.Blob{MIMEType: mimeType, Data: raw}},
	}
	return []*genai.Content{{Role: genai.RoleUser, Parts: parts}}, nil
}

// ContentsFromText creates genai.Content from a text string.
func ContentsFromText(text string) []*genai.Content {
	return []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}
}

// Embed generates an embedding vector for the given content.
func Embed(ctx context.Context, client *genai.Client, model string, contents []*genai.Content) ([]float32, error) {
	result, err := client.Models.EmbedContent(ctx, model, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("embed content: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return result.Embeddings[0].Values, nil
}
