package main

import (
	"context"
	"net/http"
	"time"
)

// previewTimeout bounds each lookup against the index store.
const previewTimeout = 5 * time.Second

// ServeHTTP implements the Wails asset-server fallback handler.
//
// Requests for embedded frontend assets never reach this method; anything
// else — e.g. /preview?path=... — does. Only paths that are present in the
// index may be served, so the UI can display real thumbnails and media
// without exposing arbitrary disk contents.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/preview" {
		http.NotFound(w, r)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), previewTimeout)
	defer cancel()

	indexed, err := a.db.HasFilePath(ctx, path)
	if err != nil {
		http.Error(w, "index unavailable", http.StatusInternalServerError)
		return
	}
	if !indexed {
		http.Error(w, "file is not in the search index", http.StatusForbidden)
		return
	}

	// ServeFile supports Range requests, so video/audio seeking works.
	http.ServeFile(w, r, path)
}
