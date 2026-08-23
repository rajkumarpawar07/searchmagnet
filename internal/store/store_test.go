package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEmptyIndex(t *testing.T) {
	ctx := context.Background()
	store := NewJSONStore("nonexistent_file_for_test.json")
	
	count, err := store.Count(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 entries, got %d", count)
	}
}

func TestSaveAndReload(t *testing.T) {
	ctx := context.Background()
	tmp := filepath.Join(t.TempDir(), "test_index.json")

	store := NewJSONStore(tmp)

	entry := IndexEntry{
		ID:          "test-1",
		Label:       "test label",
		FilePath:    "/fake/path.png",
		ContentType: "image",
		MIMEType:    "image/png",
		Embedding:   []float32{0.1, 0.2, 0.3},
		IndexedAt:   time.Now(),
	}

	if err := store.Add(ctx, entry); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Create a new store instance to test reloading from disk
	loadedStore := NewJSONStore(tmp)
	entries, err := loadedStore.List(ctx)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Label != "test label" {
		t.Fatalf("expected label 'test label', got %q", entries[0].Label)
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	ctx := context.Background()
	tmp := filepath.Join(t.TempDir(), "corrupt.json")
	os.WriteFile(tmp, []byte("{bad json"), 0644)

	store := NewJSONStore(tmp)
	_, err := store.List(ctx)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestHasFilePath(t *testing.T) {
	ctx := context.Background()
	tmp := filepath.Join(t.TempDir(), "test_index.json")
	store := NewJSONStore(tmp)

	_ = store.Add(ctx, IndexEntry{FilePath: "/a/b.png", Embedding: []float32{0}})
	_ = store.Add(ctx, IndexEntry{FilePath: "/c/d.mp4", Embedding: []float32{0}})

	has, err := store.HasFilePath(ctx, "/a/b.png")
	if err != nil || !has {
		t.Fatal("expected to find /a/b.png")
	}
	has, err = store.HasFilePath(ctx, "/x/y.pdf")
	if err != nil || has {
		t.Fatal("should not find /x/y.pdf")
	}
}

func TestDeleteByIDPrefix(t *testing.T) {
	ctx := context.Background()
	tmp := filepath.Join(t.TempDir(), "test_index.json")
	store := NewJSONStore(tmp)

	_ = store.Add(ctx, IndexEntry{ID: "abc123", Embedding: []float32{0}})
	_ = store.Add(ctx, IndexEntry{ID: "abc456", Embedding: []float32{0}})
	_ = store.Add(ctx, IndexEntry{ID: "def789", Embedding: []float32{0}})

	n, err := store.Delete(ctx, "abc")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 deleted, got %d", n)
	}
	
	count, _ := store.Count(ctx)
	if count != 1 {
		t.Fatalf("expected 1 remaining, got %d", count)
	}
	
	entries, _ := store.List(ctx)
	if entries[0].ID != "def789" {
		t.Fatalf("expected def789, got %s", entries[0].ID)
	}
}

func TestFindByID(t *testing.T) {
	ctx := context.Background()
	tmp := filepath.Join(t.TempDir(), "test_index.json")
	store := NewJSONStore(tmp)

	_ = store.Add(ctx, IndexEntry{ID: "1111111", Label: "first", Embedding: []float32{0}})
	_ = store.Add(ctx, IndexEntry{ID: "2222222", Label: "second", Embedding: []float32{0}})

	e, err := store.FindByID(ctx, "111")
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if e == nil || e.Label != "first" {
		t.Fatal("expected to find 'first'")
	}
	
	e, err = store.FindByID(ctx, "999")
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if e != nil {
		t.Fatal("should not find anything for '999'")
	}
}

func TestCosineIdentical(t *testing.T) {
	v := []float32{1, 2, 3}
	score := CosineSimilarity(v, v)
	if score < 0.9999 || score > 1.0001 {
		t.Fatalf("identical vectors should have similarity ~1.0, got %f", score)
	}
}

func TestCosineOrthogonal(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	score := CosineSimilarity(a, b)
	if score > 0.0001 || score < -0.0001 {
		t.Fatalf("orthogonal vectors should have similarity ~0.0, got %f", score)
	}
}

func TestCosineDifferentLengths(t *testing.T) {
	a := []float32{1, 2, 3}
	b := []float32{1, 2}
	score := CosineSimilarity(a, b)
	if score != 0 {
		t.Fatalf("mismatched lengths should return 0, got %f", score)
	}
}

func TestCosineZeroVector(t *testing.T) {
	a := []float32{0, 0, 0}
	b := []float32{1, 2, 3}
	score := CosineSimilarity(a, b)
	if score != 0 {
		t.Fatalf("zero vector should return 0, got %f", score)
	}
}
