package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultEmbeddingModel is the Gemini model used when no override is configured.
const DefaultEmbeddingModel = "gemini-embedding-2-preview"

// appDirName is the folder created inside the user's OS config directory.
const appDirName = "SearchMagnet"

// Settings holds user-editable application preferences.
type Settings struct {
	// APIKey is the Google Gemini API key used to generate embeddings.
	APIKey string `json:"apiKey"`
	// Model is the embedding model identifier.
	Model string `json:"model"`
}

// AppPaths resolves well-known locations (config file, index file) for the app.
type AppPaths struct {
	ConfigFile  string
	IndexFile   string
	Root        string
}

// ResolveAppPaths returns the paths used by the desktop app, creating the
// per-user application directory when it does not exist yet.
func ResolveAppPaths() (*AppPaths, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config dir: %w", err)
	}
	root := filepath.Join(base, appDirName)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create app dir: %w", err)
	}
	return &AppPaths{
		Root:       root,
		ConfigFile: filepath.Join(root, "config.json"),
		IndexFile:  filepath.Join(root, "embeddings.json"),
	}, nil
}

// LoadSettings reads settings from disk. Missing files are not an error:
// the caller receives defaults with environment-variable fallbacks applied
// so that existing SearchMagnet setups keep working out of the box.
func (p *AppPaths) LoadSettings() Settings {
	settings := Settings{
		APIKey: os.Getenv("GEMINI_API_KEY"),
		Model:  defaultString(os.Getenv("GEMINI_EMBEDDING_MODEL"), DefaultEmbeddingModel),
	}
	data, err := os.ReadFile(p.ConfigFile)
	if err != nil || len(data) == 0 {
		return settings
	}
	var stored Settings
	if err := json.Unmarshal(data, &stored); err != nil {
		return settings // corrupt file → fall back to defaults
	}
	stored.APIKey = defaultString(stored.APIKey, settings.APIKey)
	stored.Model = defaultString(stored.Model, settings.Model)
	return stored
}

// SaveSettings persists the given settings to the config file.
func (p *AppPaths) SaveSettings(s Settings) error {
	if s.Model == "" {
		s.Model = DefaultEmbeddingModel
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return os.WriteFile(p.ConfigFile, data, 0o600)
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// ErrNoAPIKey is returned when an operation needs Gemini but no key is set.
var ErrNoAPIKey = errors.New("no Gemini API key configured — add one in Settings")
