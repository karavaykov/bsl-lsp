package lsp

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/parser"
	"github.com/karavaikov/bsl-lsp/internal/workspace"
	"github.com/karavaikov/bsl-lsp/pkg/protocol"
)

type documentState struct {
	doc        *workspace.Document
	symbols    *analysis.SymbolTable
}

type Handler struct {
	mu          sync.Mutex
	initialized bool
	workspace   *workspace.Manager
	documents   map[string]*documentState
	logf        LogFunc
}

func NewHandler(logf LogFunc) *Handler {
	return &Handler{
		workspace: workspace.NewManager(),
		documents: make(map[string]*documentState),
		logf:      logf,
	}
}

func (h *Handler) Handle(req jsonRPCMessage) *jsonRPCMessage {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch req.Method {
	case "initialize":
		return h.handleInitialize(req)
	case "initialized":
		h.initialized = true
		return nil
	case "textDocument/didOpen":
		h.handleDidOpen(req)
		return nil
	case "textDocument/didChange":
		h.handleDidChange(req)
		return nil
	case "textDocument/didClose":
		h.handleDidClose(req)
		return nil
	case "textDocument/didSave":
		return nil
	case "textDocument/documentSymbol":
		return h.handleDocumentSymbol(req)
	case "textDocument/definition":
		return h.handleDefinition(req)
	case "textDocument/hover":
		return h.handleHover(req)
	case "textDocument/completion":
		return h.handleCompletion(req)
	case "shutdown":
		return &jsonRPCMessage{
			JSONRPC: "2.0",
			ID:      req.ID,
		}
	case "exit":
		return nil
	default:
		if req.ID != nil {
			return &jsonRPCMessage{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &jsonRPCError{
					Code:    -32601,
					Message: fmt.Sprintf("method not found: %s", req.Method),
				},
			}
		}
		return nil
	}
}

func (h *Handler) handleInitialize(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &jsonRPCMessage{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32700,
				Message: fmt.Sprintf("parse error: %v", err),
			},
		}
	}

	result := protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync:   1,
			DocumentSymbolProvider: true,
			DefinitionProvider:     true,
			HoverProvider:          true,
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{".", " "},
			},
		},
	}

	resultData, _ := json.Marshal(result)

	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resultData,
	}
}

type didOpenParams struct {
	TextDocument protocol.TextDocumentItem `json:"textDocument"`
}

func (h *Handler) handleDidOpen(req jsonRPCMessage) {
	var params didOpenParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.logf("failed to parse didOpen: %v", err)
		return
	}

	h.workspace.Open(
		params.TextDocument.URI,
		params.TextDocument.Text,
		params.TextDocument.Version,
	)

	h.publishDiagnostics(params.TextDocument.URI)
}

type didChangeParams struct {
	TextDocument   protocol.VersionedTextDocumentIdentifier `json:"textDocument"`
	ContentChanges []protocol.TextDocumentContentChangeEvent `json:"contentChanges"`
}

func (h *Handler) handleDidChange(req jsonRPCMessage) {
	var params didChangeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.logf("failed to parse didChange: %v", err)
		return
	}

	if len(params.ContentChanges) > 0 {
		h.workspace.Update(
			params.TextDocument.URI,
			params.ContentChanges[len(params.ContentChanges)-1].Text,
			params.TextDocument.Version,
		)
	}

	h.publishDiagnostics(params.TextDocument.URI)
}

type didCloseParams struct {
	TextDocument protocol.VersionedTextDocumentIdentifier `json:"textDocument"`
}

func (h *Handler) handleDidClose(req jsonRPCMessage) {
	var params didCloseParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		h.logf("failed to parse didClose: %v", err)
		return
	}
	h.workspace.Close(params.TextDocument.URI)
	delete(h.documents, params.TextDocument.URI)
}

func (h *Handler) publishDiagnostics(uri string) {
	doc, ok := h.workspace.Get(uri)
	if !ok {
		return
	}

	text := doc.GetText()
	p := parser.NewParser(text)
	mod := p.ParseModule()

	var diagnostics []protocol.Diagnostic

	for _, err := range p.Errors() {
		line := err.Line - 1
		col := err.Col - 1
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + 1},
			},
			Severity: protocol.SeverityError,
			Message:  err.Message,
		})
	}

	symbols := analysis.BuildSymbolTable(mod)
	h.documents[uri] = &documentState{
		doc:     doc,
		symbols: symbols,
	}

	diagParams := struct {
		URI         string                `json:"uri"`
		Diagnostics []protocol.Diagnostic `json:"diagnostics"`
	}{
		URI:         uri,
		Diagnostics: diagnostics,
	}

	diagData, _ := json.Marshal(diagParams)

	notification := jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params:  diagData,
	}

	if err := writeMessage(protocol.StdoutWriter{}, notification); err != nil {
		h.logf("failed to publish diagnostics: %v", err)
	}
}

func (h *Handler) handleDocumentSymbol(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.DocumentSymbolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &jsonRPCMessage{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &jsonRPCError{
				Code:    -32700,
				Message: fmt.Sprintf("parse error: %v", err),
			},
		}
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return &jsonRPCMessage{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  []byte("[]"),
		}
	}

	symbols := state.symbols
	var result []protocol.DocumentSymbol

	kindMap := map[analysis.SymbolKind]int{
		analysis.SymbolVariable:  protocol.SymbolKindVariable,
		analysis.SymbolProcedure: protocol.SymbolKindMethod,
		analysis.SymbolFunction:  protocol.SymbolKindFunction,
		analysis.SymbolParameter: protocol.SymbolKindVariable,
	}

	for _, sym := range symbols.Symbols {
		// Only include top-level symbols (global scope) and
		// skip parameters (they are shown as children of their parent)
		if sym.Scope != symbols.Global {
			continue
		}

		lspKind, ok := kindMap[sym.Kind]
		if !ok {
			lspKind = protocol.SymbolKindVariable
		}

		ds := protocol.DocumentSymbol{
			Name: sym.Name,
			Kind: lspKind,
			Range: protocol.Range{
				Start: protocol.Position{Line: sym.Line, Character: sym.Col},
				End:   protocol.Position{Line: sym.Line, Character: sym.Col + 1},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{Line: sym.Line, Character: sym.Col},
				End:   protocol.Position{Line: sym.Line, Character: sym.Col + 1},
			},
		}

		// Add body-level children (locals, params) for procedures/functions
		if sym.Kind == analysis.SymbolProcedure || sym.Kind == analysis.SymbolFunction {
			children := h.collectChildren(sym, symbols)
			if len(children) > 0 {
				ds.Children = children
			}
		}

		result = append(result, ds)
	}

	if result == nil {
		result = []protocol.DocumentSymbol{}
	}

	resultData, _ := json.Marshal(result)

	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  resultData,
	}
}

func (h *Handler) handleDefinition(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.TextDocumentPositionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte("null"))
	}

	doc, ok := h.workspace.Get(params.TextDocument.URI)
	if !ok {
		return h.jsonResult(req.ID, []byte("null"))
	}

	text := doc.GetText()
	p := parser.NewParser(text)
	mod := p.ParseModule()

	line := params.Position.Line
	col := params.Position.Character

	ident := analysis.FindIdentAtPos(mod, line+1, col+1)
	if ident == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	sym := state.symbols.Lookup(ident.Name)
	if sym == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	loc := protocol.Location{
		URI: params.TextDocument.URI,
		Range: protocol.Range{
			Start: protocol.Position{Line: sym.Line, Character: sym.Col},
			End:   protocol.Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
		},
	}

	data, _ := json.Marshal(loc)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleHover(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.TextDocumentPositionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte("null"))
	}

	doc, ok := h.workspace.Get(params.TextDocument.URI)
	if !ok {
		return h.jsonResult(req.ID, []byte("null"))
	}

	text := doc.GetText()
	p := parser.NewParser(text)
	mod := p.ParseModule()

	line := params.Position.Line
	col := params.Position.Character

	ident := analysis.FindIdentAtPos(mod, line+1, col+1)
	if ident == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	sym := state.symbols.Lookup(ident.Name)
	if sym == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	kindStr := map[analysis.SymbolKind]string{
		analysis.SymbolVariable:  "Переменная",
		analysis.SymbolProcedure: "Процедура",
		analysis.SymbolFunction:  "Функция",
		analysis.SymbolParameter: "Параметр",
	}[sym.Kind]

	content := fmt.Sprintf("**%s** `%s`", kindStr, sym.Name)

	hover := protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
	}

	data, _ := json.Marshal(hover)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleCompletion(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.CompletionParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte("null"))
	}

	var items []protocol.CompletionItem

	dotTrigger := params.Context != nil && params.Context.TriggerCharacter == "."

	if dotTrigger {
		items = append(items, protocol.CompletionItem{
			Label:  "Количество",
			Kind:   protocol.CompletionKindProperty,
			Detail: "Количество()",
		}, protocol.CompletionItem{
			Label:  "Найти",
			Kind:   protocol.CompletionKindMethod,
			Detail: "Найти(Значение)",
		}, protocol.CompletionItem{
			Label:  "Удалить",
			Kind:   protocol.CompletionKindMethod,
			Detail: "Удалить(Значение)",
		}, protocol.CompletionItem{
			Label:  "Добавить",
			Kind:   protocol.CompletionKindMethod,
			Detail: "Добавить(Значение)",
		}, protocol.CompletionItem{
			Label:  "Вставить",
			Kind:   protocol.CompletionKindMethod,
			Detail: "Вставить(Индекс, Значение)",
		}, protocol.CompletionItem{
			Label:  "Очистить",
			Kind:   protocol.CompletionKindMethod,
			Detail: "Очистить()",
		})
	} else {
		for _, kw := range analysis.BSLKeywords {
			items = append(items, protocol.CompletionItem{
				Label:      kw.Name,
				Kind:       protocol.CompletionKindKeyword,
				Detail:     kw.Detail,
				InsertText: kw.InsertText,
			})
		}

		for _, gm := range analysis.BSLGlobalMethods {
			items = append(items, protocol.CompletionItem{
				Label:  gm.Name,
				Kind:   protocol.CompletionKindFunction,
				Detail: gm.Detail,
			})
		}

		seen := make(map[string]bool)
		for _, sym := range state.symbols.Symbols {
			if seen[sym.Name] {
				continue
			}
			seen[sym.Name] = true

			ckind := protocol.CompletionKindVariable
			switch sym.Kind {
			case analysis.SymbolProcedure:
				ckind = protocol.CompletionKindMethod
			case analysis.SymbolFunction:
				ckind = protocol.CompletionKindFunction
			case analysis.SymbolParameter:
				ckind = protocol.CompletionKindVariable
			}

			items = append(items, protocol.CompletionItem{
				Label:  sym.Name,
				Kind:   ckind,
				Detail: sym.Kind.String(),
			})
		}
	}

	if items == nil {
		items = []protocol.CompletionItem{}
	}

	result := protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}

	data, _ := json.Marshal(result)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) jsonError(id *int, code int, msg string) *jsonRPCMessage {
	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error: &jsonRPCError{
			Code:    code,
			Message: msg,
		},
	}
}

func (h *Handler) jsonResult(id *int, data []byte) *jsonRPCMessage {
	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      id,
		Result:  data,
	}
}

func (h *Handler) collectChildren(parent *analysis.Symbol, table *analysis.SymbolTable) []protocol.DocumentSymbol {
	if parent.BodyScope == nil {
		return nil
	}

	var children []protocol.DocumentSymbol

	kindMap := map[analysis.SymbolKind]int{
		analysis.SymbolVariable:  protocol.SymbolKindVariable,
		analysis.SymbolParameter: protocol.SymbolKindVariable,
	}

	for _, sym := range table.Symbols {
		if sym.Scope == parent.BodyScope {
			lspKind, ok := kindMap[sym.Kind]
			if !ok {
				lspKind = protocol.SymbolKindVariable
			}

			children = append(children, protocol.DocumentSymbol{
				Name: sym.Name,
				Kind: lspKind,
				Range: protocol.Range{
					Start: protocol.Position{Line: sym.Line, Character: sym.Col},
					End:   protocol.Position{Line: sym.Line, Character: sym.Col + 1},
				},
				SelectionRange: protocol.Range{
					Start: protocol.Position{Line: sym.Line, Character: sym.Col},
					End:   protocol.Position{Line: sym.Line, Character: sym.Col + 1},
				},
			})
		}
	}

	return children
}
