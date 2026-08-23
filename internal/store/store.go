// Package store defines the Store interface and shared types for the embedding index.
package store

import (
	"context"
	"math"
	"time"
)

// IndexEntry represents a single indexed item with its embedding vector.
type IndexEntry struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	FilePath    string    `json:"file_path"`
	ContentType string    `json:"content_type"` // "text" | "image" | "video" | "audio" | "pdf"
	MIMEType    string    `json:"mime_type"`
	Embedding   []float32 `json:"embedding"`
	IndexedAt   time.Time `json:"indexed_at"`
}

// ScoredEntry is an IndexEntry with its similarity score attached.
type ScoredEntry struct {
	IndexEntry
	Score float64
}

// Store is the storage backend interface used by the server and CLI.
type Store interface {
	// Add inserts a new entry into the index.
	Add(ctx context.Context, entry IndexEntry) error

	// List returns all entries in the index (no embedding vectors).
	List(ctx context.Context) ([]IndexEntry, error)

	// Search finds the top-K most similar entries to the given vector.
	// typeFilter is an optional content_type to restrict results.
	Search(ctx context.Context, vec []float32, topK int, typeFilter string) ([]ScoredEntry, error)

	// FindByID returns the entry whose ID starts with the given prefix, or nil.
	FindByID(ctx context.Context, id string) (*IndexEntry, error)

	// Delete removes entries whose ID matches (prefix match). Returns deleted count.
	Delete(ctx context.Context, id string) (int, error)

	// HasFilePath returns true if any entry references the given file path.
	HasFilePath(ctx context.Context, path string) (bool, error)

	// Count returns the total number of indexed entries.
	Count(ctx context.Context) (int, error)

	// TypeCounts returns a map of content_type → count.
	TypeCounts(ctx context.Context) (map[string]int, error)
}

// CosineSimilarity computes cosine similarity between two float32 vectors.
// Returns 0 if lengths differ or either vector is zero.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
