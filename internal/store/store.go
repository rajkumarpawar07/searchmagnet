// Package store handles persistence and similarity math for the embedding index.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
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

// Index holds all indexed entries.
type Index struct {
	Entries []IndexEntry `json:"entries"`
}

// Load reads the index from a JSON file. Returns an empty index if the file doesn't exist.
func Load(path string) (*Index, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Index{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing index: %w", err)
	}
	return &idx, nil
}

// Save writes the index to a JSON file.
func Save(idx *Index, path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// HasFilePath checks whether any entry already references the given file path.
func (idx *Index) HasFilePath(path string) bool {
	for _, e := range idx.Entries {
		if e.FilePath == path {
			return true
		}
	}
	return false
}

// DeleteByIDPrefix removes entries whose ID starts with (or ends with, after "…") prefix.
// Returns the number of entries deleted.
func (idx *Index) DeleteByIDPrefix(prefix string) int {
	before := len(idx.Entries)
	filtered := idx.Entries[:0]
	for _, e := range idx.Entries {
		if !strings.HasSuffix(e.ID, strings.TrimPrefix(prefix, "…")) &&
			!strings.HasPrefix(e.ID, prefix) {
			filtered = append(filtered, e)
		}
	}
	idx.Entries = filtered
	return before - len(idx.Entries)
}

// FindByID returns a pointer to the entry matching the given ID prefix, or nil.
func (idx *Index) FindByID(prefix string) *IndexEntry {
	for i, e := range idx.Entries {
		if strings.HasPrefix(e.ID, prefix) ||
			strings.HasSuffix(e.ID, strings.TrimPrefix(prefix, "…")) {
			return &idx.Entries[i]
		}
	}
	return nil
}

// CosineSimilarity computes the cosine similarity between two float32 vectors.
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
