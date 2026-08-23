package main

import (
	"context"
	"log"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"google.golang.org/genai"

	"github.com/rajkumarpawar07/searchmagnet/internal/store"
)

// App is the central object of the desktop app. It owns configuration, the
// vector store and the Gemini client, and exposes the API surface that the
// frontend calls through Wails bindings.
type App struct {
	ctx      context.Context
	paths    *AppPaths
	settings Settings
	db       store.Store

	clientMu sync.Mutex
	genai    *genai.Client

	indexer *Indexer
}

// NewApp constructs the app shell before the window opens. Heavy resources
// (store, Gemini client) are created in startup so failures can be surfaced
// inside the UI instead of aborting the process.
func NewApp() *App {
	return &App{}
}

// startup runs when the window has been created.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	paths, err := ResolveAppPaths()
	if err != nil {
		log.Fatalf("desktop app: %v", err)
	}
	a.paths = paths
	a.settings = paths.LoadSettings()

	a.db = store.NewJSONStore(paths.IndexFile)
	a.indexer = NewIndexer(a.db, a.geminiClient, a.modelName, a.emitProgress)
}

// shutdown releases resources when the window closes.
func (a *App) shutdown(context.Context) {}

// modelName returns the configured embedding model.
func (a *App) modelName() string {
	return a.settings.Model
}

// geminiClient lazily creates (and caches) a Gemini client using the current
// settings. It returns ErrNoAPIKey when no key has been configured yet.
func (a *App) geminiClient(ctx context.Context) (*genai.Client, error) {
	a.clientMu.Lock()
	defer a.clientMu.Unlock()

	if a.settings.APIKey == "" {
		return nil, ErrNoAPIKey
	}
	if a.genai == nil {
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: a.settings.APIKey})
		if err != nil {
			return nil, err
		}
		a.genai = client
	}
	return a.genai, nil
}

// resetGeminiClient drops the cached client so the next call re-authenticates.
func (a *App) resetGeminiClient() {
	a.clientMu.Lock()
	defer a.clientMu.Unlock()
	a.genai = nil
}

// emitProgress forwards indexer progress to the frontend.
func (a *App) emitProgress(event IndexProgressEvent) {
	runtime.EventsEmit(a.ctx, "indexer:progress", event)
}
