package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"google.golang.org/genai"

	"github.com/rajkumarpawar07/searchmagnet/internal/embedder"
	"github.com/rajkumarpawar07/searchmagnet/internal/format"
	"github.com/rajkumarpawar07/searchmagnet/internal/store"
)

const indexFile = "embeddings.json"

var client *genai.Client

func embeddingModel() string {
	m := os.Getenv("GEMINI_EMBEDDING_MODEL")
	if m == "" {
		return "gemini-embedding-2-preview"
	}
	return m
}

// ──────────────────────────────────────────────────────────────
// Response types
// ──────────────────────────────────────────────────────────────

type EntryResponse struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	FilePath    string `json:"file_path"`
	ContentType string `json:"content_type"`
	MIMEType    string `json:"mime_type"`
	IndexedAt   string `json:"indexed_at"`
}

type SearchResult struct {
	Rank  int     `json:"rank"`
	Score float64 `json:"score"`
	EntryResponse
}

func entryToResponse(e store.IndexEntry) EntryResponse {
	return EntryResponse{
		ID:          e.ID,
		Label:       e.Label,
		FilePath:    e.FilePath,
		ContentType: e.ContentType,
		MIMEType:    e.MIMEType,
		IndexedAt:   e.IndexedAt.Format(time.RFC3339),
	}
}

// ──────────────────────────────────────────────────────────────
// POST /api/index — upload & index a file
// ──────────────────────────────────────────────────────────────

func handleIndex(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "file is required (multipart form field 'file')"})
	}

	label := c.FormValue("label", file.Filename)

	// Save uploaded file to temp location
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, file.Filename)
	if err := c.SaveFile(file, tmpPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("saving file: %v", err)})
	}
	defer os.Remove(tmpPath)

	// Detect type
	mimeType, contentType, err := embedder.DetectType(tmpPath)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Size check
	size, _ := embedder.FileSize(tmpPath)
	if size > embedder.MaxFileSize {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("file too large: %s (max 20 MB)", format.HumanBytes(size)),
		})
	}

	// Duplicate check
	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}
	if idx.HasFilePath(file.Filename) {
		return c.Status(409).JSON(fiber.Map{"error": "file already indexed", "file": file.Filename})
	}

	// Embed
	contents, err := embedder.ContentsFromFile(tmpPath, mimeType, label)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("preparing content: %v", err)})
	}

	ctx := context.Background()
	vec, err := embedder.Embed(ctx, client, embeddingModel(), contents)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("embedding: %v", err)})
	}

	entry := store.IndexEntry{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Label:       label,
		FilePath:    file.Filename,
		ContentType: contentType,
		MIMEType:    mimeType,
		Embedding:   vec,
		IndexedAt:   time.Now(),
	}
	idx.Entries = append(idx.Entries, entry)

	if err := store.Save(idx, indexFile); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("saving index: %v", err)})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "indexed",
		"entry":   entryToResponse(entry),
		"dim":     len(vec),
	})
}

// ──────────────────────────────────────────────────────────────
// POST /api/index-text — index a text snippet
// ──────────────────────────────────────────────────────────────

func handleIndexText(c *fiber.Ctx) error {
	type req struct {
		Text string `json:"text"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil || body.Text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "JSON body with 'text' field is required"})
	}

	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	contents := embedder.ContentsFromText(body.Text)
	ctx := context.Background()
	vec, err := embedder.Embed(ctx, client, embeddingModel(), contents)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("embedding: %v", err)})
	}

	entry := store.IndexEntry{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		Label:       body.Text,
		ContentType: "text",
		Embedding:   vec,
		IndexedAt:   time.Now(),
	}
	idx.Entries = append(idx.Entries, entry)

	if err := store.Save(idx, indexFile); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("saving index: %v", err)})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "indexed",
		"entry":   entryToResponse(entry),
		"dim":     len(vec),
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/search?q=...&top=5&type=image
// ──────────────────────────────────────────────────────────────

func handleSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query parameter 'q' is required"})
	}

	topK := c.QueryInt("top", 5)
	typeFilter := c.Query("type", "")

	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}
	if len(idx.Entries) == 0 {
		return c.JSON(fiber.Map{"results": []SearchResult{}, "message": "index is empty"})
	}

	ctx := context.Background()
	queryVec, err := embedder.Embed(ctx, client, embeddingModel(), embedder.ContentsFromText(query))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("embedding query: %v", err)})
	}

	type scored struct {
		entry store.IndexEntry
		score float64
	}
	var results []scored
	for _, e := range idx.Entries {
		if typeFilter != "" && e.ContentType != typeFilter {
			continue
		}
		results = append(results, scored{e, store.CosineSimilarity(queryVec, e.Embedding)})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if topK > len(results) {
		topK = len(results)
	}

	out := make([]SearchResult, topK)
	for i, r := range results[:topK] {
		out[i] = SearchResult{
			Rank:          i + 1,
			Score:         r.score,
			EntryResponse: entryToResponse(r.entry),
		}
	}

	return c.JSON(fiber.Map{
		"query":   query,
		"results": out,
		"total":   len(results),
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/list
// ──────────────────────────────────────────────────────────────

func handleList(c *fiber.Ctx) error {
	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	entries := make([]EntryResponse, len(idx.Entries))
	for i, e := range idx.Entries {
		entries[i] = entryToResponse(e)
	}

	return c.JSON(fiber.Map{
		"entries": entries,
		"total":   len(entries),
	})
}

// ──────────────────────────────────────────────────────────────
// DELETE /api/entries/:id
// ──────────────────────────────────────────────────────────────

func handleDelete(c *fiber.Ctx) error {
	idPrefix := c.Params("id")
	if idPrefix == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id parameter is required"})
	}

	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	deleted := idx.DeleteByIDPrefix(idPrefix)
	if deleted == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "no entry found matching that ID"})
	}

	if err := store.Save(idx, indexFile); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("saving index: %v", err)})
	}

	return c.JSON(fiber.Map{"deleted": deleted})
}

// ──────────────────────────────────────────────────────────────
// GET /api/similar/:id?top=5
// ──────────────────────────────────────────────────────────────

func handleSimilar(c *fiber.Ctx) error {
	idPrefix := c.Params("id")
	topK := c.QueryInt("top", 5)

	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	target := idx.FindByID(idPrefix)
	if target == nil {
		return c.Status(404).JSON(fiber.Map{"error": "no entry found matching that ID"})
	}

	type scored struct {
		entry store.IndexEntry
		score float64
	}
	var results []scored
	for _, e := range idx.Entries {
		if e.ID == target.ID {
			continue
		}
		results = append(results, scored{e, store.CosineSimilarity(target.Embedding, e.Embedding)})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if topK > len(results) {
		topK = len(results)
	}

	out := make([]SearchResult, topK)
	for i, r := range results[:topK] {
		out[i] = SearchResult{
			Rank:          i + 1,
			Score:         r.score,
			EntryResponse: entryToResponse(r.entry),
		}
	}

	return c.JSON(fiber.Map{
		"source":  entryToResponse(*target),
		"results": out,
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/stats
// ──────────────────────────────────────────────────────────────

func handleStats(c *fiber.Ctx) error {
	idx, err := store.Load(indexFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	typeCounts := make(map[string]int)
	for _, e := range idx.Entries {
		typeCounts[e.ContentType]++
	}

	// Supported extensions
	var exts []string
	for ext := range embedder.SupportedTypes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	return c.JSON(fiber.Map{
		"total_entries":        len(idx.Entries),
		"entries_by_type":      typeCounts,
		"supported_extensions": exts,
		"max_file_size":        format.HumanBytes(embedder.MaxFileSize),
		"embedding_model":      embeddingModel(),
	})
}

// ──────────────────────────────────────────────────────────────
// GET /api/file?path=...
// ──────────────────────────────────────────────────────────────

func handleFile(c *fiber.Ctx) error {
	filePath := c.Query("path")
	if filePath == "" {
		return c.Status(400).JSON(fiber.Map{"error": "path parameter is required"})
	}

	// For a local tool, we'll serve the file directly. 
	// In production, this would need serious path validation/security.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "file not found on disk"})
	}

	return c.SendFile(filePath)
}

// ──────────────────────────────────────────────────────────────
// Main
// ──────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	// Init Gemini client
	ctx := context.Background()
	var err error
	client, err = genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("create genai client: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app := fiber.New(fiber.Config{
		AppName:   "SearchMagnet API",
		BodyLimit: 25 * 1024 * 1024, // 25 MB upload limit
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New()) // allow all origins for dev

	// Health check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "SearchMagnet API",
			"version": "1.0.0",
			"endpoints": []string{
				"POST   /api/index       - Upload & index a file",
				"POST   /api/index-text  - Index a text snippet",
				"GET    /api/search      - Search (q, top, type)",
				"GET    /api/list        - List all entries",
				"DELETE /api/entries/:id  - Delete an entry",
				"GET    /api/similar/:id - Find similar entries",
				"GET    /api/stats       - Index statistics",
			},
		})
	})

	// API routes
	api := app.Group("/api")
	api.Post("/index", handleIndex)
	api.Post("/index-text", handleIndexText)
	api.Get("/search", handleSearch)
	api.Get("/list", handleList)
	api.Delete("/entries/:id", handleDelete)
	api.Get("/similar/:id", handleSimilar)
	api.Get("/stats", handleStats)
	api.Get("/file", handleFile)

	// Serve static web UI (future)
	if _, err := os.Stat("web/dist"); err == nil {
		app.Static("/app", "./web/dist")
	}

	banner := fmt.Sprintf(`
%s╔══════════════════════════════════════════════╗
║   SearchMagnet API Server                    ║
║   %s→%s http://localhost:%s                     ║
╚══════════════════════════════════════════════╝%s
`, format.Bold, format.Cyan, format.Reset+format.Bold,
		padPort(port), format.Reset)
	fmt.Print(banner)

	log.Fatal(app.Listen(":" + port))
}

func padPort(port string) string {
	// Pad to keep the box aligned
	for len(port) < 4 {
		port += " "
	}
	return port
}
