package main

import (
	"context"
	"fmt"
	"os"

	"github.com/tikhomirovv/book-translator/internal/infrastructure/extract"
)

func main() {
	r := extract.NewRegistry()
	p, err := r.Extract(context.Background(), ".books/1.pdf")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("paragraphs:", len(p))
	for i := 0; i < len(p) && i < 5; i++ {
		text := p[i].Text
		n := len(text)
		if n > 400 {
			text = text[:400] + "..."
		}
		fmt.Printf("p[%d] len=%d\n%s\n---\n", i, n, text)
	}
}
