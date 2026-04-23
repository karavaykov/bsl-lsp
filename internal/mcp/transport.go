package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Transport struct {
	server  *Server
	sseClients map[int]chan []byte
	sseMu      sync.Mutex
	sseNextID  int
}

func NewTransport(srv *Server) *Transport {
	return &Transport{
		server:     srv,
		sseClients: make(map[int]chan []byte),
	}
}

func (t *Transport) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/":
		t.handlePost(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/sse":
		t.handleSSE(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (t *Transport) handlePost(w http.ResponseWriter, r *http.Request) {
	var req jsonRPCMessage
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, -32700, "Parse error: "+err.Error())
		return
	}

	if req.Method == "" {
		writeJSONError(w, http.StatusBadRequest, -32600, "Invalid Request: empty method")
		return
	}

	resp := t.server.Handle(req)

	if resp != nil {
		writeJSON(w, http.StatusOK, resp)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func (t *Transport) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan []byte, 16)
	t.sseMu.Lock()
	id := t.sseNextID
	t.sseNextID++
	t.sseClients[id] = ch
	t.sseMu.Unlock()

	defer func() {
		t.sseMu.Lock()
		delete(t.sseClients, id)
		t.sseMu.Unlock()
	}()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (t *Transport) broadcast(event []byte) {
	t.sseMu.Lock()
	defer t.sseMu.Unlock()
	for _, ch := range t.sseClients {
		select {
		case ch <- event:
		default:
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, code int, msg string) {
	writeJSON(w, status, jsonRPCMessage{
		JSONRPC: "2.0",
		Error: &jsonRPCError{
			Code:    code,
			Message: msg,
		},
	})
}
