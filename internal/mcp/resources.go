package mcp

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
)

type ResourceStore struct {
	mu   sync.RWMutex
	data map[string]*cachedResource
}

type cachedResource struct {
	content ResourceContent
}

func NewResourceStore() *ResourceStore {
	return &ResourceStore{
		data: make(map[string]*cachedResource),
	}
}

func (s *ResourceStore) Store(code string, content ResourceContent) string {
	hash := hashCode(code)
	s.mu.Lock()
	s.data[hash] = &cachedResource{content: content}
	s.mu.Unlock()
	return hash
}

func (s *ResourceStore) Get(uri string) (ResourceContent, bool) {
	hash := uriToHash(uri)
	s.mu.RLock()
	cached, ok := s.data[hash]
	s.mu.RUnlock()
	if !ok {
		return ResourceContent{}, false
	}
	return cached.content, true
}

func (s *ResourceStore) List() []Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	resources := make([]Resource, 0, len(s.data))
	for hash, cached := range s.data {
		resources = append(resources, Resource{
			URI:         "bsl://" + hash + "/" + cached.content.MimeType,
			Name:        "BSL Module (" + hash[:8] + ")",
			Description: "Cached analysis result",
			MimeType:    cached.content.MimeType,
		})
	}
	return resources
}

func (s *Server) registerResources() {
}

func (s *Server) cacheAnalysisResult(code string, result map[string]any, suffix string) string {
	content := ResourceContent{
		MimeType: "application/json",
		Text:     encodeJSON(result),
	}
	return s.resourceStore.Store(code, content)
}

func hashCode(code string) string {
	prefix := code
	if len(prefix) > 100 {
		prefix = prefix[:100]
	}
	h := sha256.Sum256([]byte(prefix))
	return fmt.Sprintf("%x", h[:16])
}

func uriToHash(uri string) string {
	if strings.HasPrefix(uri, "bsl://") {
		rest := uri[6:]
		if idx := strings.IndexByte(rest, '/'); idx > 0 {
			return rest[:idx]
		}
	}
	return uri
}
