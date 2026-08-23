package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/rajkumarpawar07/searchmagnet/internal/embedder"
)

// ──────────────────────────────────────────────────────────────
// Search
// ──────────────────────────────────────────────────────────────

// Search runs a semantic query against the index. typeFilter restricts the
// results to one content type ("" means all types).
func (a *App) Search(query, typeFilter string, topK int) ([]SearchHitDTO, error) {
	if query == "" {
		return []SearchHitDTO{}, nil
	}
	if topK <= 0 {
		topK = 12
	}

	client, err := a.geminiClient(a.ctx)
	if err != nil {
		return nil, err
	}

	vec, err := embedder.Embed(a.ctx, client, a.modelName(), embedder.ContentsFromText(query))
	if err != nil {
		return nil, fmt.Errorf("embedding query: %w", err)
	}

	hits, err := a.db.Search(a.ctx, vec, topK, typeFilter)
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}

	out := make([]SearchHitDTO, len(hits))
	for i, hit := range hits {
		out[i] = newSearchHitDTO(i+1, hit)
	}
	return out, nil
}

// Similar returns entries semantically close to an existing entry.
func (a *App) Similar(id string, topK int) ([]SearchHitDTO, error) {
	if topK <= 0 {
		topK = 6
	}

	target, err := a.db.FindByID(a.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("finding entry: %w", err)
	}
	if target == nil {
		return nil, errors.New("entry not found")
	}

	hits, err := a.db.Search(a.ctx, target.Embedding, topK+1, "")
	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}

	out := make([]SearchHitDTO, 0, topK)
	for _, hit := range hits {
		if hit.ID == target.ID {
			continue // skip the source entry itself
		}
		if len(out) >= topK {
			break
		}
		out = append(out, newSearchHitDTO(len(out)+1, hit))
	}
	return out, nil
}

// ──────────────────────────────────────────────────────────────
// Indexing
// ──────────────────────────────────────────────────────────────

// PickFilesAndIndex opens a multi-select file dialog and indexes everything chosen.
func (a *App) PickFilesAndIndex() (IndexSummary, error) {
	paths, err := wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Choose files to index",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Supported media (*.png;*.jpg;*.mp4;*.wav;*.pdf;…)", Pattern: supportedPattern()},
			{DisplayName: "All files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return IndexSummary{}, err
	}
	return a.indexPaths(paths), nil
}

// PickFolderAndIndex lets the user choose a folder and indexes every
// supported file inside it (recursively).
func (a *App) PickFolderAndIndex() (IndexSummary, error) {
	folder, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Choose a folder to index",
	})
	if err != nil {
		return IndexSummary{}, err
	}
	if folder == "" {
		return IndexSummary{}, nil
	}
	return a.indexFolder(folder)
}

// indexPaths runs the indexer over the given absolute paths.
func (a *App) indexPaths(paths []string) IndexSummary {
	return a.indexer.IndexFiles(a.ctx, paths)
}

// indexFolder walks root and feeds every supported file to the indexer.
func (a *App) indexFolder(root string) (IndexSummary, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // unreadable subtree — skip rather than abort the run
		}
		if !d.IsDir() && embedder.IsSupportedExt(path) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return IndexSummary{}, fmt.Errorf("scanning folder: %w", err)
	}
	return a.indexer.IndexFiles(a.ctx, paths), nil
}

// OnFilesDropped is called by the frontend when files are dragged onto the window.
func (a *App) OnFilesDropped(paths []string) IndexSummary {
	return a.indexPaths(paths)
}

// IndexTextSnippet embeds and stores a piece of text.
func (a *App) IndexTextSnippet(text string) (EntryDTO, error) {
	id, err := a.indexer.IndexText(a.ctx, text)
	if err != nil {
		return EntryDTO{}, err
	}
	entry, err := a.db.FindByID(a.ctx, id)
	if err != nil || entry == nil {
		return EntryDTO{}, fmt.Errorf("stored entry could not be read back: %w", err)
	}
	return newEntryDTO(*entry), nil
}

// ──────────────────────────────────────────────────────────────
// Library
// ──────────────────────────────────────────────────────────────

// ListEntries returns everything in the index, newest first.
func (a *App) ListEntries() ([]EntryDTO, error) {
	entries, err := a.db.List(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("loading index: %w", err)
	}

	out := make([]EntryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, newEntryDTO(e))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].IndexedAt > out[j].IndexedAt })
	return out, nil
}

// RecentEntries returns the newest entries for the dashboard grid.
func (a *App) RecentEntries(limit int) ([]EntryDTO, error) {
	all, err := a.ListEntries()
	if err != nil || limit <= 0 {
		return all, err
	}
	if limit > len(all) {
		limit = len(all)
	}
	return all[:limit], nil
}

// DeleteEntry removes entries whose ID matches the given prefix.
func (a *App) DeleteEntry(id string) (int, error) {
	if id == "" {
		return 0, errors.New("id is required")
	}
	return a.db.Delete(a.ctx, id)
}

// ──────────────────────────────────────────────────────────────
// Stats & settings
// ──────────────────────────────────────────────────────────────

// GetStats powers the sidebar stats card and settings diagnostics.
func (a *App) GetStats() (StatsDTO, error) {
	count, err := a.db.Count(a.ctx)
	if err != nil {
		return StatsDTO{}, fmt.Errorf("counting entries: %w", err)
	}
	typeCounts, err := a.db.TypeCounts(a.ctx)
	if err != nil {
		return StatsDTO{}, fmt.Errorf("counting types: %w", err)
	}

	stats := StatsDTO{
		TotalEntries: count,
		TypeCounts:   typeCounts,
		Model:        a.modelName(),
		IndexFile:    a.paths.IndexFile,
		HasAPIKey:    a.settings.APIKey != "",
	}

	var lastIndexed time.Time
	if entries, err := a.db.List(a.ctx); err == nil {
		for _, e := range entries {
			stats.TotalBytes += fileSizeOf(e.FilePath)
			if e.IndexedAt.After(lastIndexed) {
				lastIndexed = e.IndexedAt
			}
		}
	}
	stats.LastUpdated = formatTime(lastIndexed)
	return stats, nil
}

// GetSettings returns current user-editable settings.
func (a *App) GetSettings() Settings {
	return a.settings
}

// SaveSettings persists settings and re-authenticates the Gemini client.
func (a *App) SaveSettings(apiKey, model string) error {
	a.settings.APIKey = apiKey
	a.settings.Model = defaultString(model, DefaultEmbeddingModel)

	a.resetGeminiClient()
	if err := a.paths.SaveSettings(a.settings); err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}

	// Fail fast on an obviously bad key so users get immediate feedback.
	_, err := a.geminiClient(a.ctx)
	return err
}

// ──────────────────────────────────────────────────────────────
// File helpers
// ──────────────────────────────────────────────────────────────

// OpenFile launches the OS default application for path.
func (a *App) OpenFile(path string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

// RevealFile shows path in the platform file manager.
func (a *App) RevealFile(path string) error {
	switch goruntime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", path).Start()
	case "darwin":
		return exec.Command("open", "-R", path).Start()
	default:
		dir := filepath.Dir(path)
		return exec.Command("xdg-open", dir).Start()
	}
}

// PreviewURL exposes the URL scheme used to load indexed file contents
// inside the UI (images, videos, audio, PDFs).
func (a *App) PreviewURL(path string) string {
	return "/preview?path=" + url.QueryEscape(path)
}

// ──────────────────────────────────────────────────────────────
// small helpers
// ──────────────────────────────────────────────────────────────

func supportedPattern() string {
	pattern := ""
	for ext := range embedder.SupportedTypes {
		pattern += "*" + ext + ";"
	}
	return pattern
}
