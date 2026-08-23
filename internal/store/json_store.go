package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// jsonIndex is the on-disk representation.
type jsonIndex struct {
	Entries []IndexEntry `json:"entries"`
}

// JSONStore persists entries to a local JSON file.
// All reads load the full file; all writes re-serialize the full file.
// This is intentional for a local/small-scale store (<500 items).
type JSONStore struct {
	mu   sync.RWMutex
	path string
}

// NewJSONStore creates a JSONStore backed by the given file path.
func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

// load reads and parses the JSON file. Caller must hold at least a read lock.
func (s *JSONStore) load() (*jsonIndex, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return &jsonIndex{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading index: %w", err)
	}
	var idx jsonIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing index: %w", err)
	}
	return &idx, nil
}

// save serializes and writes the index to disk. Caller must hold a write lock.
func (s *JSONStore) save(idx *jsonIndex) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}
	return os.WriteFile(s.path, data, 0644)
}

// Add inserts a new entry.
func (s *JSONStore) Add(_ context.Context, entry IndexEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.load()
	if err != nil {
		return err
	}
	idx.Entries = append(idx.Entries, entry)
	return s.save(idx)
}

// List returns all entries (includes embeddings for the JSON backend).
func (s *JSONStore) List(_ context.Context) ([]IndexEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	// Return copies without embeddings to keep the API consistent.
	out := make([]IndexEntry, len(idx.Entries))
	for i, e := range idx.Entries {
		out[i] = IndexEntry{
			ID:          e.ID,
			Label:       e.Label,
			FilePath:    e.FilePath,
			ContentType: e.ContentType,
			MIMEType:    e.MIMEType,
			IndexedAt:   e.IndexedAt,
		}
	}
	return out, nil
}

// Search performs brute-force cosine similarity against all entries.
func (s *JSONStore) Search(_ context.Context, vec []float32, topK int, typeFilter string) ([]ScoredEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return nil, err
	}

	var scored []ScoredEntry
	for _, e := range idx.Entries {
		if typeFilter != "" && e.ContentType != typeFilter {
			continue
		}
		scored = append(scored, ScoredEntry{
			IndexEntry: e,
			Score:      CosineSimilarity(vec, e.Embedding),
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	if topK > len(scored) {
		topK = len(scored)
	}
	return scored[:topK], nil
}

// FindByID returns the entry matching the ID prefix.
func (s *JSONStore) FindByID(_ context.Context, prefix string) (*IndexEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	for i, e := range idx.Entries {
		if strings.HasPrefix(e.ID, prefix) ||
			strings.HasSuffix(e.ID, strings.TrimPrefix(prefix, "…")) {
			cp := idx.Entries[i]
			return &cp, nil
		}
	}
	return nil, nil
}

// Delete removes entries matching the ID prefix. Returns the number deleted.
func (s *JSONStore) Delete(_ context.Context, prefix string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx, err := s.load()
	if err != nil {
		return 0, err
	}
	before := len(idx.Entries)
	filtered := idx.Entries[:0]
	for _, e := range idx.Entries {
		if !strings.HasSuffix(e.ID, strings.TrimPrefix(prefix, "…")) &&
			!strings.HasPrefix(e.ID, prefix) {
			filtered = append(filtered, e)
		}
	}
	idx.Entries = filtered
	if err := s.save(idx); err != nil {
		return 0, err
	}
	return before - len(idx.Entries), nil
}

// HasFilePath checks for an existing entry with the given file path.
func (s *JSONStore) HasFilePath(_ context.Context, path string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return false, err
	}
	for _, e := range idx.Entries {
		if e.FilePath == path {
			return true, nil
		}
	}
	return false, nil
}

// Count returns the total number of entries.
func (s *JSONStore) Count(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return 0, err
	}
	return len(idx.Entries), nil
}

// TypeCounts returns a map of content_type → count.
func (s *JSONStore) TypeCounts(_ context.Context) (map[string]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, e := range idx.Entries {
		counts[e.ContentType]++
	}
	return counts, nil
}

// --- Legacy helpers used by the CLI directly (for zero-refactor compatibility) ---

// LegacyLoad reads the JSON index from a file path.
// Deprecated: prefer NewJSONStore and the Store interface.
func LegacyLoad(path string) (*jsonIndex, error) {
	s := &JSONStore{path: path}
	return s.load()
}

// LegacySave writes the JSON index to a file path.
// Deprecated: prefer NewJSONStore and the Store interface.
func LegacySave(idx *jsonIndex, path string) error {
	s := &JSONStore{path: path}
	return s.save(idx)
}

// LegacyEntries exposes the entries slice from a jsonIndex (for migration).
func LegacyEntries(idx *jsonIndex) []IndexEntry {
	if idx == nil {
		return nil
	}
	return idx.Entries
}

// LegacyIndexedAt is a zero-safe helper for migration.
func LegacyIndexedAt(e IndexEntry) string {
	if e.IndexedAt.IsZero() {
		return time.Now().Format(time.RFC3339)
	}
	return e.IndexedAt.Format(time.RFC3339)
}
