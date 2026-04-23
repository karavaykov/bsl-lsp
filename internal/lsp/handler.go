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
	project     *analysis.ProjectAnalysis
	logf        LogFunc
}

func NewHandler(logf LogFunc) *Handler {
	return &Handler{
		workspace: workspace.NewManager(),
		documents: make(map[string]*documentState),
		project:   analysis.NewProjectAnalysis(),
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
	case "textDocument/semanticTokens/full":
		return h.handleSemanticTokens(req)
	case "textDocument/codeLens":
		return h.handleCodeLens(req)
	case "textDocument/foldingRange":
		return h.handleFoldingRange(req)
	case "textDocument/formatting":
		return h.handleFormatting(req)
	case "textDocument/signatureHelp":
		return h.handleSignatureHelp(req)
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

	semTokLegend := protocol.SemanticTokensLegend{
		TokenTypes:     []string{"keyword", "variable", "function", "method", "parameter", "property", "string", "number", "comment", "operator"},
		TokenModifiers: []string{},
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
			SemanticTokensProvider: &protocol.SemanticTokensOptions{
				Legend: semTokLegend,
				Full:   true,
			},
			CodeLensProvider:     &protocol.CodeLensOptions{},
			FoldingRangeProvider: true,
			DocumentFormattingProvider: true,
			SignatureHelpProvider: &protocol.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
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
	h.project.RemoveModule(params.TextDocument.URI)
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
	h.project.UpdateModule(uri, symbols)

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
		foundURI, foundSym := h.project.LookupSymbol(ident.Name)
		if foundSym == nil {
			return h.jsonResult(req.ID, []byte("null"))
		}
		loc := protocol.Location{
			URI: foundURI,
			Range: protocol.Range{
				Start: protocol.Position{Line: foundSym.Line, Character: foundSym.Col},
				End:   protocol.Position{Line: foundSym.Line, Character: foundSym.Col + len(foundSym.Name)},
			},
		}
		data, _ := json.Marshal(loc)
		return h.jsonResult(req.ID, data)
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

	var sym *analysis.Symbol

	sym = state.symbols.Lookup(ident.Name)
	if sym == nil {
		_, foundSym := h.project.LookupSymbol(ident.Name)
		sym = foundSym
	}
	if sym == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	kindStr := map[analysis.SymbolKind]string{
		analysis.SymbolVariable:  "Переменная",
		analysis.SymbolProcedure: "Процедура",
		analysis.SymbolFunction:  "Функция",
		analysis.SymbolParameter: "Параметр",
	}[sym.Kind]

	exportTag := ""
	if sym.Export {
		exportTag = " (Экспорт)"
	}
	content := fmt.Sprintf("**%s** `%s`%s", kindStr, sym.Name, exportTag)

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

		for _, ms := range h.project.Modules {
			if ms.URI == params.TextDocument.URI {
				continue
			}
			for name, sym := range ms.Exports {
				if seen[name] {
					continue
				}
				seen[name] = true

				ckind := protocol.CompletionKindMethod
				switch sym.Kind {
				case analysis.SymbolFunction:
					ckind = protocol.CompletionKindFunction
				}
				items = append(items, protocol.CompletionItem{
					Label:  name,
					Kind:   ckind,
					Detail: sym.Kind.String() + " (Экспорт, " + ms.URI + ")",
				})
			}
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

func (h *Handler) handleSemanticTokens(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.SemanticTokensParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte(`{"data":[]}`))
	}

	doc, ok := h.workspace.Get(params.TextDocument.URI)
	if !ok {
		return h.jsonResult(req.ID, []byte(`{"data":[]}`))
	}

	text := doc.GetText()
	p := parser.NewParser(text)
	mod := p.ParseModule()

	tokens := analysis.CollectSemanticTokens(mod, state.symbols)
	data, _ := json.Marshal(protocol.SemanticTokens{Data: tokens})
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleCodeLens(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.CodeLensParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte("[]"))
	}

	var result []protocol.CodeLens
	for _, sym := range state.symbols.Symbols {
		if sym.Scope != state.symbols.Global {
			continue
		}
		if sym.Kind != analysis.SymbolProcedure && sym.Kind != analysis.SymbolFunction {
			continue
		}

		exportLabel := ""
		if sym.Export {
			exportLabel = "Экспорт"
		} else {
			exportLabel = "Локальная"
		}

		result = append(result, protocol.CodeLens{
			Range: protocol.Range{
				Start: protocol.Position{Line: sym.Line, Character: sym.Col},
				End:   protocol.Position{Line: sym.Line, Character: sym.Col + len(sym.Name)},
			},
			Command: &protocol.Command{
				Title:   exportLabel,
				Command: "",
			},
		})
	}

	if result == nil {
		result = []protocol.CodeLens{}
	}

	data, _ := json.Marshal(result)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleFoldingRange(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.FoldingRangeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	state, ok := h.documents[params.TextDocument.URI]
	if !ok {
		return h.jsonResult(req.ID, []byte("[]"))
	}

	doc, ok := h.workspace.Get(params.TextDocument.URI)
	if !ok {
		return h.jsonResult(req.ID, []byte("[]"))
	}

	text := doc.GetText()
	p := parser.NewParser(text)
	mod := p.ParseModule()

	ranges := analysis.CollectFoldingRanges(mod, state.symbols)

	var result []protocol.FoldingRange
	for _, r := range ranges {
		result = append(result, protocol.FoldingRange{
			StartLine: r.StartLine,
			EndLine:   r.EndLine,
			Kind:      r.Kind,
		})
	}

	if result == nil {
		result = []protocol.FoldingRange{}
	}

	data, _ := json.Marshal(result)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleFormatting(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.DocumentFormattingParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return h.jsonError(req.ID, -32700, fmt.Sprintf("parse error: %v", err))
	}

	doc, ok := h.workspace.Get(params.TextDocument.URI)
	if !ok {
		return h.jsonResult(req.ID, []byte("[]"))
	}

	text := doc.GetText()
	formatted := analysis.FormatDocument(text, params.Options.TabSize, params.Options.InsertSpaces)

	edits := []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: len(text), Character: 0},
		},
		NewText: formatted,
	}}

	data, _ := json.Marshal(edits)
	return h.jsonResult(req.ID, data)
}

func (h *Handler) handleSignatureHelp(req jsonRPCMessage) *jsonRPCMessage {
	var params protocol.SignatureHelpParams
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

	call := analysis.FindCallAtPos(mod, line+1, col+1)
	if call == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	sym := state.symbols.Lookup(call.Name)
	if sym == nil {
		_, foundSym := h.project.LookupSymbol(call.Name)
		sym = foundSym
	}
	if sym == nil {
		return h.jsonResult(req.ID, []byte("null"))
	}

	sig := protocol.SignatureInformation{
		Label: sym.Kind.String() + " " + call.Name + "(...)",
	}

	if sym.BodyScope != nil {
		for _, child := range state.symbols.Symbols {
			if child.Scope == sym.BodyScope && child.Kind == analysis.SymbolParameter {
				sig.Parameters = append(sig.Parameters, protocol.ParameterInformation{
					Label: child.Name,
				})
			}
		}
	}

	help := protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{sig},
		ActiveSignature: 0,
		ActiveParameter: call.ActiveParam,
	}

	data, _ := json.Marshal(help)
	return h.jsonResult(req.ID, data)
}
