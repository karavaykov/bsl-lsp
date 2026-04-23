package main

import (
	"log"

	"github.com/karavaikov/bsl-lsp/internal/lsp"
)

func main() {
	log.SetFlags(log.Lshortfile)
	log.Println("bsl-lsp starting...")

	if err := lsp.Run(); err != nil {
		log.Fatalf("bsl-lsp error: %v", err)
	}
}
