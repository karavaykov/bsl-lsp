package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/karavaikov/bsl-lsp/internal/mcp"
)

func main() {
	host := flag.String("host", "localhost", "Host to listen on")
	port := flag.Int("port", 9090, "Port to listen on")
	flag.Parse()

	srv := mcp.NewServer()
	transport := mcp.NewTransport(srv)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("bsl-lsp-mcp starting on %s", addr)
	log.Printf("  POST /  — JSON-RPC 2.0 endpoint")
	log.Printf("  GET  /sse — SSE stream")

	if err := http.ListenAndServe(addr, transport); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
