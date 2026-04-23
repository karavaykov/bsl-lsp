.PHONY: build test vet run clean run-mcp

build:
	go build -o bsl-lsp ./cmd/bsl-lsp
	go build -o bsl-lsp-mcp ./cmd/bsl-lsp-mcp

test:
	go test -v ./...

vet:
	go vet ./...

run:
	go run ./cmd/bsl-lsp

run-mcp:
	go run ./cmd/bsl-lsp-mcp

clean:
	rm -f bsl-lsp bsl-lsp-mcp
