package store

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const defaultIndexFile = "embeddings.json"

// NewStore creates the appropriate Store backend based on the STORE_BACKEND
// environment variable.
//
//	STORE_BACKEND=json   (default) → JSONStore backed by embeddings.json
//	STORE_BACKEND=chroma           → ChromaStore backed by a ChromaDB HTTP server
func NewStore(ctx context.Context) (Store, error) {
	backend := strings.ToLower(os.Getenv("STORE_BACKEND"))
	switch backend {
	case "chroma":
		s, err := NewChromaStore(ctx)
		if err != nil {
			return nil, fmt.Errorf("init chroma store: %w", err)
		}
		return s, nil
	default: // "json" or empty
		path := os.Getenv("JSON_INDEX_PATH")
		if path == "" {
			path = defaultIndexFile
		}
		return NewJSONStore(path), nil
	}
}
