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

var client *genai.Client
var db store.Store

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
	ctx := context.Background()
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
	exists, err := db.HasFilePath(ctx, file.Filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("checking index: %v", err)})
	}
	if exists {
		return c.Status(409).JSON(fiber.Map{"error": "file already indexed", "file": file.Filename})
	}

	// Embed
	contents, err := embedder.ContentsFromFile(tmpPath, mimeType, label)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("preparing content: %v", err)})
	}

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

	if err := db.Add(ctx, entry); err != nil {
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
	ctx := context.Background()
	type req struct {
		Text string `json:"text"`
	}
	var body req
	if err := c.BodyParser(&body); err != nil || body.Text == "" {
		return c.Status(400).JSON(fiber.Map{"error": "JSON body with 'text' field is required"})
	}

	contents := embedder.ContentsFromText(body.Text)
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

	if err := db.Add(ctx, entry); err != nil {
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
	ctx := context.Background()
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query parameter 'q' is required"})
	}

	topK := c.QueryInt("top", 5)
	typeFilter := c.Query("type", "")

	count, err := db.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("checking index: %v", err)})
	}
	if count == 0 {
		return c.JSON(fiber.Map{"results": []SearchResult{}, "message": "index is empty"})
	}

	queryVec, err := embedder.Embed(ctx, client, embeddingModel(), embedder.ContentsFromText(query))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("embedding query: %v", err)})
	}

	results, err := db.Search(ctx, queryVec, topK, typeFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("search: %v", err)})
	}

	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{
			Rank:          i + 1,
			Score:         r.Score,
			EntryResponse: entryToResponse(r.IndexEntry),
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
	ctx := context.Background()
	entries, err := db.List(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("loading index: %v", err)})
	}

	out := make([]EntryResponse, len(entries))
	for i, e := range entries {
		out[i] = entryToResponse(e)
	}

	return c.JSON(fiber.Map{
		"entries": out,
		"total":   len(out),
	})
}

// ──────────────────────────────────────────────────────────────
// DELETE /api/entries/:id
// ──────────────────────────────────────────────────────────────

func handleDelete(c *fiber.Ctx) error {
	ctx := context.Background()
	idPrefix := c.Params("id")
	if idPrefix == "" {
		return c.Status(400).JSON(fiber.Map{"error": "id parameter is required"})
	}

	deleted, err := db.Delete(ctx, idPrefix)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("deleting index: %v", err)})
	}
	if deleted == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "no entry found matching that ID"})
	}

	return c.JSON(fiber.Map{"deleted": deleted})
}

// ──────────────────────────────────────────────────────────────
// GET /api/similar/:id?top=5
// ──────────────────────────────────────────────────────────────

func handleSimilar(c *fiber.Ctx) error {
	ctx := context.Background()
	idPrefix := c.Params("id")
	topK := c.QueryInt("top", 5)

	target, err := db.FindByID(ctx, idPrefix)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("finding target: %v", err)})
	}
	if target == nil {
		return c.Status(404).JSON(fiber.Map{"error": "no entry found matching that ID"})
	}
	if len(target.Embedding) == 0 {
		return c.Status(500).JSON(fiber.Map{"error": "target entry has no embedding vector available for similarity search"})
	}

	// Request topK+1 because the result will likely include the target itself
	results, err := db.Search(ctx, target.Embedding, topK+1, "")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("search: %v", err)})
	}

	var out []SearchResult
	rank := 1
	for _, r := range results {
		if r.ID == target.ID {
			continue // exclude the target itself
		}
		if len(out) >= topK {
			break
		}
		out = append(out, SearchResult{
			Rank:          rank,
			Score:         r.Score,
			EntryResponse: entryToResponse(r.IndexEntry),
		})
		rank++
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
	ctx := context.Background()
	count, err := db.Count(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("count: %v", err)})
	}

	typeCounts, err := db.TypeCounts(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("type counts: %v", err)})
	}

	// Supported extensions
	var exts []string
	for ext := range embedder.SupportedTypes {
		exts = append(exts, ext)
	}
	sort.Strings(exts)

	return c.JSON(fiber.Map{
		"total_entries":        count,
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
	ctx := context.Background()

	// Init DB
	var err error
	db, err = store.NewStore(ctx)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}

	// Init Gemini client
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
