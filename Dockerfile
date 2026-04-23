FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bsl-lsp ./cmd/bsl-lsp

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bsl-lsp /usr/local/bin/bsl-lsp
ENTRYPOINT ["bsl-lsp"]
