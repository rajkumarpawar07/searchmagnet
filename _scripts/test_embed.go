package main
import (
	"context"
	"fmt"
	"log"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)
func main() {
	godotenv.Load()
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	model := "gemini-embedding-2-preview"
	contents := []*genai.Content{{Role: genai.RoleUser, Parts: []*genai.Part{{Text: "hello"}}}}
	res, err := client.Models.EmbedContent(ctx, model, contents, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("len: %d\n", len(res.Embeddings[0].Values))
}
