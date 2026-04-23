.PHONY: build test vet run clean

build:
	go build -o bsl-lsp ./cmd/bsl-lsp

test:
	go test -v ./...

vet:
	go vet ./...

run:
	go run ./cmd/bsl-lsp

clean:
	rm -f bsl-lsp
