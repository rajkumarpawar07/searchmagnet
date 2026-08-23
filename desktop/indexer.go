package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/genai"

	"github.com/rajkumarpawar07/searchmagnet/internal/embedder"
	"github.com/rajkumarpawar07/searchmagnet/internal/store"
)

// indexerWorkers is the number of files embedded concurrently.
const indexerWorkers = 4

// clientProvider returns a ready-to-use Gemini client for the current settings.
type clientProvider func(context.Context) (*genai.Client, error)

// Indexer turns files and text into embeddings and writes them to the store.
// Progress is streamed to the frontend via Wails "indexer:progress" events.
type Indexer struct {
	db       store.Store
	client   clientProvider
	model    func() string
	emit     func(event IndexProgressEvent)
}

// NewIndexer wires an Indexer together.
//
//	client — provides an authenticated Gemini client (or ErrNoAPIKey).
//	model  — resolves the embedding model on each call so setting changes apply live.
//	emit   — receives one progress event per processed file.
func NewIndexer(db store.Store, client clientProvider, model func() string, emit func(IndexProgressEvent)) *Indexer {
	return &Indexer{db: db, client: client, model: model, emit: emit}
}

// IndexFiles indexes the given paths concurrently and returns a summary.
// Unsupported extensions and duplicates count as skipped; embedding failures
// count as failed and are collected into Summary.Errors.
func (ix *Indexer) IndexFiles(ctx context.Context, paths []string) IndexSummary {
	total := len(paths)
	if total == 0 {
		return IndexSummary{}
	}

	jobs := make(chan string)
	results := make(chan IndexProgressEvent, total)

	var workers sync.WaitGroup
	for i := 0; i < indexerWorkers; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for path := range jobs {
				results <- ix.indexFile(ctx, path)
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, path := range paths {
			select {
			case jobs <- path:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		workers.Wait()
		close(results)
	}()

	summary := IndexSummary{Errors: []string{}}
	completed := 0
	for event := range results {
		completed++
		event.Completed = completed
		event.Total = total
		ix.emit(event)

		switch event.Status {
		case statusDone:
			summary.Indexed++
		case statusSkipped:
			summary.Skipped++
		default:
			summary.Failed++
			summary.Errors = append(summary.Errors,
				fmt.Sprintf("%s: %s", event.File, event.Error))
		}
	}
	return summary
}

const (
	statusDone    = "done"
	statusSkipped = "skipped"
	statusFailed  = "failed"
)

// indexFile runs the full pipeline for one file: detect type → dedupe check →
// embed → persist. The returned event carries a skip/failure reason when needed.
func (ix *Indexer) indexFile(ctx context.Context, path string) IndexProgressEvent {
	file := filepath.Base(path)
	fail := func(err error) IndexProgressEvent {
		return IndexProgressEvent{File: file, Error: err.Error()}
	}

	mimeType, contentType, err := embedder.DetectType(path)
	if err != nil {
		return IndexProgressEvent{File: file, Status: statusSkipped, Error: err.Error()}
	}

	if exists, err := ix.db.HasFilePath(ctx, path); err == nil && exists {
		return IndexProgressEvent{File: file, Status: statusSkipped, Error: "already indexed"}
	}

	entry, err := ix.embedAndStore(ctx, path, mimeType, contentType)
	if err != nil {
		return fail(err)
	}

	return IndexProgressEvent{File: file, Status: statusDone, EntryID: entry.ID}
}

// embedAndStore generates the embedding vector and saves the resulting entry.
func (ix *Indexer) embedAndStore(ctx context.Context, path, mimeType, contentType string) (*store.IndexEntry, error) {
	client, err := ix.client(ctx)
	if err != nil {
		return nil, err
	}

	label := filepath.Base(path)
	contents, err := embedder.ContentsFromFile(path, mimeType, label)
	if err != nil {
		return nil, err
	}

	vec, err := embedder.Embed(ctx, client, ix.model(), contents)
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}

	entry := store.IndexEntry{
		ID:          fmt.Sprintf("%d-%s", time.Now().UnixNano(), label),
		Label:       label,
		FilePath:    path, // absolute path so previews and "open file" work
		ContentType: contentType,
		MIMEType:    mimeType,
		Embedding:   vec,
		IndexedAt:   time.Now(),
	}
	if err := ix.db.Add(ctx, entry); err != nil {
		return nil, fmt.Errorf("saving entry: %w", err)
	}
	return &entry, nil
}

// IndexText embeds a text snippet and stores it. Returns the stored ID.
func (ix *Indexer) IndexText(ctx context.Context, text string) (string, error) {
	client, err := ix.client(ctx)
	if err != nil {
		return "", err
	}

	trimmed := trimToLength(text, 80)
	vec, err := embedder.Embed(ctx, client, ix.model(), embedder.ContentsFromText(text))
	if err != nil {
		return "", fmt.Errorf("embedding: %w", err)
	}

	entry := store.IndexEntry{
		ID:          fmt.Sprintf("%d-text", time.Now().UnixNano()),
		Label:       trimmed,
		ContentType: "text",
		Embedding:   vec,
		IndexedAt:   time.Now(),
	}
	if err := ix.db.Add(ctx, entry); err != nil {
		return "", fmt.Errorf("saving entry: %w", err)
	}
	return entry.ID, nil
}

// trimToLength shortens s to n characters for use as a label.
func trimToLength(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
