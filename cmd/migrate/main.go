package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/rajkumarpawar07/searchmagnet/internal/store"
)

func main() {
	_ = godotenv.Load()

	// Connect to ChromaDB
	ctx := context.Background()
	chromaStore, err := store.NewChromaStore(ctx)
	if err != nil {
		log.Fatalf("failed to connect to ChromaDB: %v\nMake sure it is running via docker-compose up -d", err)
	}

	// Load JSON index directly using the legacy helper
	fmt.Println("Loading existing embeddings.json...")
	idx, err := store.LegacyLoad("embeddings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No embeddings.json found. Nothing to migrate.")
			return
		}
		log.Fatalf("failed to load JSON index: %v", err)
	}

	entries := store.LegacyEntries(idx)
	if len(entries) == 0 {
		fmt.Println("embeddings.json is empty. Nothing to migrate.")
		return
	}

	fmt.Printf("Found %d entries. Starting migration to ChromaDB...\n", len(entries))

	var success, failures int
	for i, entry := range entries {
		fmt.Printf("[%d/%d] Migrating %s (%s)... ", i+1, len(entries), entry.Label, entry.ContentType)
		
		err := chromaStore.Add(ctx, entry)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failures++
		} else {
			fmt.Println("OK")
			success++
		}
	}

	fmt.Printf("\nMigration complete!\nSuccessfully migrated: %d\nFailures: %d\n", success, failures)
	if failures == 0 {
		fmt.Println("\nYou can now safely set STORE_BACKEND=chroma in your .env file and restart the server.")
		fmt.Println("Once verified, you may delete embeddings.json.")
	}
}
