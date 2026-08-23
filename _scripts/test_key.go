package main
import (
	"fmt"
	"os"
	"github.com/joho/godotenv"
)
func main() {
	godotenv.Load()
	fmt.Println(os.Getenv("GEMINI_API_KEY"))
}
