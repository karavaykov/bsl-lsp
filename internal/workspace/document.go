package workspace

import "sync"

type Document struct {
	mu      sync.RWMutex
	URI     string
	Version int
	Text    string
}

func NewDocument(uri, text string, version int) *Document {
	return &Document{
		URI:     uri,
		Version: version,
		Text:    text,
	}
}

func (d *Document) SetText(text string, version int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Text = text
	d.Version = version
}

func (d *Document) GetText() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Text
}

func (d *Document) GetVersion() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Version
}

type Manager struct {
	mu        sync.RWMutex
	documents map[string]*Document
}

func NewManager() *Manager {
	return &Manager{
		documents: make(map[string]*Document),
	}
}

func (m *Manager) Open(uri, text string, version int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents[uri] = NewDocument(uri, text, version)
}

func (m *Manager) Update(uri, text string, version int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if doc, ok := m.documents[uri]; ok {
		doc.SetText(text, version)
	}
}

func (m *Manager) Close(uri string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.documents, uri)
}

func (m *Manager) Get(uri string) (*Document, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	doc, ok := m.documents[uri]
	return doc, ok
}
