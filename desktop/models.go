package main

import (
	"os"
	"time"

	"github.com/rajkumarpawar07/searchmagnet/internal/store"
)

// EntryDTO is the frontend representation of an indexed item.
// Embedding vectors are intentionally excluded — the UI never needs them.
type EntryDTO struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	FilePath    string `json:"filePath"`
	ContentType string `json:"contentType"` // "text" | "image" | "video" | "audio" | "pdf"
	MIMEType    string `json:"mimeType"`
	SizeBytes   int64  `json:"sizeBytes"`
	IndexedAt   string `json:"indexedAt"`
}

// SearchHitDTO is a search result: the entry plus its similarity metadata.
type SearchHitDTO struct {
	Rank  int     `json:"rank"`
	Score float64 `json:"score"`
	Entry EntryDTO `json:"entry"`
}

// IndexSummary reports what happened during one indexing run.
type IndexSummary struct {
	Indexed int      `json:"indexed"`
	Skipped int      `json:"skipped"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors,omitempty"`
}

// IndexProgressEvent is emitted over Wails events while indexing runs.
type IndexProgressEvent struct {
	File      string `json:"file"`
	Completed int    `json:"completed"` // files finished so far in this run
	Total     int    `json:"total"`     // number of files in this run
	Status    string `json:"status"`    // "done" | "skipped" | "failed"
	Error     string `json:"error,omitempty"`
	EntryID   string `json:"entryId,omitempty"`
}

// StatsDTO powers the sidebar "Index Stats" card and Settings diagnostics.
type StatsDTO struct {
	TotalEntries int            `json:"totalEntries"`
	TypeCounts   map[string]int `json:"typeCounts"`
	TotalBytes   int64          `json:"totalBytes"`
	LastUpdated  string         `json:"lastUpdated"`
	Model        string         `json:"model"`
	IndexFile    string         `json:"indexFile"`
	HasAPIKey    bool           `json:"hasApiKey"`
}

// newEntryDTO converts a store entry for the frontend and attaches the
// on-disk file size (0 when the file no longer exists or is text-only).
func newEntryDTO(e store.IndexEntry) EntryDTO {
	return EntryDTO{
		ID:          e.ID,
		Label:       e.Label,
		FilePath:    e.FilePath,
		ContentType: e.ContentType,
		MIMEType:    e.MIMEType,
		SizeBytes:   fileSizeOf(e.FilePath),
		IndexedAt:   formatTime(e.IndexedAt),
	}
}

// newSearchHitDTO converts a scored store entry into its DTO form.
func newSearchHitDTO(rank int, hit store.ScoredEntry) SearchHitDTO {
	return SearchHitDTO{
		Rank:  rank,
		Score: hit.Score,
		Entry: newEntryDTO(hit.IndexEntry),
	}
}

// fileSizeOf stats path, returning 0 instead of an error for missing files.
func fileSizeOf(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
