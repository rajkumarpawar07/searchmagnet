package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadEmptyIndex(t *testing.T) {
	idx, err := Load("nonexistent_file_for_test.json")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(idx.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(idx.Entries))
	}
}

func TestSaveAndReload(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test_index.json")

	idx := &Index{
		Entries: []IndexEntry{
			{
				ID:          "test-1",
				Label:       "test label",
				FilePath:    "/fake/path.png",
				ContentType: "image",
				MIMEType:    "image/png",
				Embedding:   []float32{0.1, 0.2, 0.3},
				IndexedAt:   time.Now(),
			},
		},
	}

	if err := Save(idx, tmp); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(tmp)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Label != "test label" {
		t.Fatalf("expected label 'test label', got %q", loaded.Entries[0].Label)
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "corrupt.json")
	os.WriteFile(tmp, []byte("{bad json"), 0644)

	_, err := Load(tmp)
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}

func TestHasFilePath(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{FilePath: "/a/b.png"},
			{FilePath: "/c/d.mp4"},
		},
	}
	if !idx.HasFilePath("/a/b.png") {
		t.Fatal("expected to find /a/b.png")
	}
	if idx.HasFilePath("/x/y.pdf") {
		t.Fatal("should not find /x/y.pdf")
	}
}

func TestDeleteByIDPrefix(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{ID: "abc123"},
			{ID: "abc456"},
			{ID: "def789"},
		},
	}
	n := idx.DeleteByIDPrefix("abc")
	if n != 2 {
		t.Fatalf("expected 2 deleted, got %d", n)
	}
	if len(idx.Entries) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(idx.Entries))
	}
	if idx.Entries[0].ID != "def789" {
		t.Fatalf("expected def789, got %s", idx.Entries[0].ID)
	}
}

func TestFindByID(t *testing.T) {
	idx := &Index{
		Entries: []IndexEntry{
			{ID: "1111111", Label: "first"},
			{ID: "2222222", Label: "second"},
		},
	}
	e := idx.FindByID("111")
	if e == nil || e.Label != "first" {
		t.Fatal("expected to find 'first'")
	}
	if idx.FindByID("999") != nil {
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
