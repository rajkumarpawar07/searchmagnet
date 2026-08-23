package store

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"
)

const defaultCollection = "searchmagnet"

// ChromaStore is a Store backed by a ChromaDB HTTP server.
type ChromaStore struct {
	client     chroma.Client
	collection chroma.Collection
}

// NewChromaStore connects to ChromaDB and gets-or-creates the collection.
// It uses pre-computed embeddings (no embedding function on the collection side).
func NewChromaStore(ctx context.Context) (*ChromaStore, error) {
	url := os.Getenv("CHROMA_URL")
	if url == "" {
		url = "http://localhost:8000"
	}
	collectionName := os.Getenv("CHROMA_COLLECTION")
	if collectionName == "" {
		collectionName = defaultCollection
	}

	client, err := chroma.NewHTTPClient(chroma.WithBaseURL(url))
	if err != nil {
		return nil, fmt.Errorf("chroma: create client: %w", err)
	}

	// Get or create the collection. We don't set an embedding function because
	// we always supply pre-computed Gemini embedding vectors.
	col, err := client.GetOrCreateCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("chroma: get/create collection %q: %w", collectionName, err)
	}

	return &ChromaStore{client: client, collection: col}, nil
}

// Add inserts a new entry using its pre-computed embedding.
func (s *ChromaStore) Add(ctx context.Context, entry IndexEntry) error {
	emb := make([]embeddings.Embedding, 1)
	emb[0] = &embeddings.Float32Embedding{ArrayOfFloat32: &entry.Embedding}

	meta, _ := chroma.NewDocumentMetadataFromMap(map[string]any{
		"label":        entry.Label,
		"file_path":    entry.FilePath,
		"content_type": entry.ContentType,
		"mime_type":    entry.MIMEType,
		"indexed_at":   entry.IndexedAt.Format(time.RFC3339),
	})

	err := s.collection.Add(ctx,
		chroma.WithIDs(chroma.DocumentID(entry.ID)),
		chroma.WithEmbeddings(emb...),
		chroma.WithTexts(entry.Label),
		chroma.WithMetadatas(meta),
	)
	if err != nil {
		return fmt.Errorf("chroma: add entry %s: %w", entry.ID, err)
	}
	return nil
}

// List returns all entries in the collection without embedding vectors.
func (s *ChromaStore) List(ctx context.Context) ([]IndexEntry, error) {
	count, err := s.collection.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("chroma: count: %w", err)
	}
	if count == 0 {
		return nil, nil
	}

	results, err := s.collection.Get(ctx,
		chroma.WithInclude(chroma.IncludeMetadatas, chroma.IncludeDocuments),
		chroma.WithLimit(int(count)),
	)
	if err != nil {
		return nil, fmt.Errorf("chroma: list: %w", err)
	}

	return chromaResultsToEntries(results.GetIDs(), results.GetMetadatas()), nil
}

// Search queries ChromaDB using the pre-computed embedding vector.
func (s *ChromaStore) Search(ctx context.Context, vec []float32, topK int, typeFilter string) ([]ScoredEntry, error) {
	count, err := s.collection.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("chroma: count: %w", err)
	}
	if count == 0 {
		return nil, nil
	}
	if topK > int(count) {
		topK = int(count)
	}

	queryOpts := []chroma.CollectionQueryOption{
		chroma.WithQueryEmbeddings(&embeddings.Float32Embedding{ArrayOfFloat32: &vec}),
		chroma.WithNResults(topK),
		chroma.WithInclude(chroma.IncludeMetadatas, chroma.IncludeDistances, chroma.IncludeDocuments),
	}

	// Optional metadata filter on content_type.
	if typeFilter != "" {
		queryOpts = append(queryOpts,
			chroma.WithWhere(chroma.EqString("content_type", typeFilter)),
		)
	}

	results, err := s.collection.Query(ctx, queryOpts...)
	if err != nil {
		return nil, fmt.Errorf("chroma: query: %w", err)
	}

	ids := results.GetIDGroups()
	metas := results.GetMetadatasGroups()
	distances := results.GetDistancesGroups()

	if len(ids) == 0 {
		return nil, nil
	}

	entries := chromaResultsToEntries(ids[0], metas[0])
	scored := make([]ScoredEntry, len(entries))
	for i, e := range entries {
		dist := float32(0.0)
		if len(distances) > 0 && len(distances[0]) > i {
			dist = float32(distances[0][i])
		}
		scored[i] = ScoredEntry{
			IndexEntry: e,
			Score:      float64(1.0 - dist/2.0),
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})
	return scored, nil
}

// FindByID returns the entry matching the ID prefix, or nil.
func (s *ChromaStore) FindByID(ctx context.Context, prefix string) (*IndexEntry, error) {
	// ChromaDB supports exact ID lookups. For prefix match we list all and filter.
	// For this scale, exact match is fine since IDs are nanosecond timestamps.
	results, err := s.collection.Get(ctx,
		chroma.WithIDs(chroma.DocumentID(prefix)),
		chroma.WithInclude(chroma.IncludeMetadatas, chroma.IncludeEmbeddings),
	)
	if err != nil {
		// Fallback: try listing and doing prefix matching.
		all, listErr := s.List(ctx)
		if listErr != nil {
			return nil, fmt.Errorf("chroma: find by id: %w", err)
		}
		for i := range all {
			if strings.HasPrefix(all[i].ID, prefix) {
				return &all[i], nil
			}
		}
		return nil, nil
	}

	ids := results.GetIDs()
	if len(ids) == 0 {
		return nil, nil
	}
	entries := chromaResultsToEntries(ids, results.GetMetadatas())
	if len(entries) == 0 {
		return nil, nil
	}

	// Attach the embedding if available.
	embs := results.GetEmbeddings()
	if len(embs) > 0 {
		e := embs[0]
		entries[0].Embedding = e.ContentAsFloat32()
	}
	return &entries[0], nil
}

// Delete removes the entry with the given ID. Returns 1 if deleted, 0 if not found.
func (s *ChromaStore) Delete(ctx context.Context, id string) (int, error) {
	// First verify it exists.
	entry, err := s.FindByID(ctx, id)
	if err != nil {
		return 0, err
	}
	if entry == nil {
		return 0, nil
	}

	err = s.collection.Delete(ctx, chroma.WithIDs(chroma.DocumentID(entry.ID)))
	if err != nil {
		return 0, fmt.Errorf("chroma: delete %s: %w", id, err)
	}
	return 1, nil
}

// HasFilePath checks if any entry has the given file_path in its metadata.
func (s *ChromaStore) HasFilePath(ctx context.Context, path string) (bool, error) {
	results, err := s.collection.Get(ctx,
		chroma.WithWhere(chroma.EqString("file_path", path)),
		chroma.WithLimit(1),
	)
	if err != nil {
		return false, fmt.Errorf("chroma: has file path: %w", err)
	}
	return len(results.GetIDs()) > 0, nil
}

// Count returns the total number of entries.
func (s *ChromaStore) Count(ctx context.Context) (int, error) {
	n, err := s.collection.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("chroma: count: %w", err)
	}
	return int(n), nil
}

// TypeCounts returns a map of content_type → count by listing all metadata.
func (s *ChromaStore) TypeCounts(ctx context.Context) (map[string]int, error) {
	entries, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, e := range entries {
		counts[e.ContentType]++
	}
	return counts, nil
}

// chromaResultsToEntries converts parallel ID + metadata slices into IndexEntry values.
func chromaResultsToEntries(ids []chroma.DocumentID, metas []chroma.DocumentMetadata) []IndexEntry {
	entries := make([]IndexEntry, len(ids))
	for i, id := range ids {
		m := metas[i]
		e := IndexEntry{ID: string(id)}
		if m != nil {
			if v, ok := m.GetString("label"); ok { e.Label = v }
			if v, ok := m.GetString("file_path"); ok { e.FilePath = v }
			if v, ok := m.GetString("content_type"); ok { e.ContentType = v }
			if v, ok := m.GetString("mime_type"); ok { e.MIMEType = v }
			if v, ok := m.GetString("indexed_at"); ok {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					e.IndexedAt = t
				}
			}
		}
		entries[i] = e
	}
	return entries
}
