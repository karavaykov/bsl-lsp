package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/karavaikov/bsl-lsp/internal/analysis"
	"github.com/karavaikov/bsl-lsp/internal/analysis/linters"
	"github.com/karavaikov/bsl-lsp/internal/parser"
)

func (s *Server) registerTools() {
	s.registerTool(Tool{
		Name:        "bsl_parse",
		Description: "Parse BSL source code and return AST and parse errors",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code to parse",
				},
			},
			Required: []string{"text"},
		},
	}, s.handleParse)

	s.registerTool(Tool{
		Name:        "bsl_lint",
		Description: "Run static analysis on BSL code and return linter diagnostics",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code to analyze",
				},
			},
			Required: []string{"text"},
		},
	}, s.handleLint)

	s.registerTool(Tool{
		Name:        "bsl_format",
		Description: "Format BSL source code",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code to format",
				},
				"tabSize": {
					Type:        "number",
					Description: "Number of spaces per indent level (default: 4)",
					Default:     4,
				},
				"insertSpaces": {
					Type:        "boolean",
					Description: "Use spaces instead of tabs (default: true)",
					Default:     true,
				},
			},
			Required: []string{"text"},
		},
	}, s.handleFormat)

	s.registerTool(Tool{
		Name:        "bsl_symbols",
		Description: "Extract symbols (procedures, functions, variables, parameters) from BSL module",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code to analyze",
				},
			},
			Required: []string{"text"},
		},
	}, s.handleSymbols)

	s.registerTool(Tool{
		Name:        "bsl_define",
		Description: "Find definition of an identifier at a given position",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code",
				},
				"line": {
					Type:        "number",
					Description: "1-based line number of the identifier",
				},
				"col": {
					Type:        "number",
					Description: "1-based column number of the identifier",
				},
			},
			Required: []string{"text", "line", "col"},
		},
	}, s.handleDefine)

	s.registerTool(Tool{
		Name:        "bsl_hover",
		Description: "Get hover information for an identifier at a given position",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code",
				},
				"line": {
					Type:        "number",
					Description: "1-based line number of the identifier",
				},
				"col": {
					Type:        "number",
					Description: "1-based column number of the identifier",
				},
			},
			Required: []string{"text", "line", "col"},
		},
	}, s.handleHover)

	s.registerTool(Tool{
		Name:        "bsl_folding_ranges",
		Description: "Get folding ranges for a BSL module",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropertySchema{
				"text": {
					Type:        "string",
					Description: "BSL source code",
				},
			},
			Required: []string{"text"},
		},
	}, s.handleFoldingRanges)
}

type parseArgs struct {
	Text string `json:"text"`
}

func (s *Server) handleParse(raw json.RawMessage) ToolCallResult {
	var args parseArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	parseErrs := p.Errors()

	type parseErrorJSON struct {
		Line    int    `json:"line"`
		Col     int    `json:"col"`
		Length  int    `json:"length"`
		Message string `json:"message"`
	}

	errs := make([]parseErrorJSON, len(parseErrs))
	for i, e := range parseErrs {
		errs[i] = parseErrorJSON{Line: e.Line, Col: e.Col, Length: e.Length, Message: e.Message}
	}

	result := map[string]any{
		"parseErrors": errs,
		"ast":         astToJSON(mod),
	}

	return textResult(encodeJSON(result))
}

func (s *Server) handleLint(raw json.RawMessage) ToolCallResult {
	var args parseArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		return errorResult("Parse errors: " + p.Errors()[0].Message)
	}

	st := analysis.BuildSymbolTable(mod)
	diags := linters.RunAll(mod, st)

	type diagJSON struct {
		Line     int    `json:"line"`
		Col      int    `json:"col"`
		Length   int    `json:"length"`
		Message  string `json:"message"`
		Code     string `json:"code"`
		Severity int    `json:"severity"`
	}

	items := make([]diagJSON, len(diags))
	for i, d := range diags {
		items[i] = diagJSON{Line: d.Line, Col: d.Col, Length: d.Length, Message: d.Message, Code: d.Code, Severity: d.Severity}
	}

	return textResult(encodeJSON(map[string]any{"diagnostics": items}))
}

func (s *Server) handleFormat(raw json.RawMessage) ToolCallResult {
	var args struct {
		Text         string `json:"text"`
		TabSize      int    `json:"tabSize"`
		InsertSpaces bool   `json:"insertSpaces"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}
	if args.TabSize <= 0 {
		args.TabSize = 4
	}

	formatted := analysis.FormatDocument(args.Text, args.TabSize, args.InsertSpaces)
	return textResult(encodeJSON(map[string]any{"formatted": formatted}))
}

func (s *Server) handleSymbols(raw json.RawMessage) ToolCallResult {
	var args parseArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		return errorResult("Parse errors: " + p.Errors()[0].Message)
	}

	st := analysis.BuildSymbolTable(mod)

	type symbolJSON struct {
		Name   string `json:"name"`
		Kind   string `json:"kind"`
		Line   int    `json:"line"`
		Col    int    `json:"col"`
		Export bool   `json:"export"`
	}

	symbols := make([]symbolJSON, 0, len(st.Symbols))
	for _, sym := range st.Symbols {
		symbols = append(symbols, symbolJSON{
			Name:   sym.Name,
			Kind:   sym.Kind.String(),
			Line:   sym.Line,
			Col:    sym.Col,
			Export: sym.Export,
		})
	}

	type funcJSON struct {
		Name     string `json:"name"`
		Exported bool   `json:"exported"`
		Line     int    `json:"line"`
		Col      int    `json:"col"`
	}

	funcs := make([]funcJSON, 0)
	for _, sym := range st.Symbols {
		if sym.Kind == analysis.SymbolProcedure || sym.Kind == analysis.SymbolFunction {
			funcs = append(funcs, funcJSON{
				Name:     sym.Name,
				Exported: sym.Export,
				Line:     sym.Line,
				Col:      sym.Col,
			})
		}
	}

	return textResult(encodeJSON(map[string]any{
		"symbols":   symbols,
		"functions": funcs,
	}))
}

func (s *Server) handleDefine(raw json.RawMessage) ToolCallResult {
	var args struct {
		Text string `json:"text"`
		Line int    `json:"line"`
		Col  int    `json:"col"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		return errorResult("Parse errors: " + p.Errors()[0].Message)
	}

	st := analysis.BuildSymbolTable(mod)
	ident := analysis.FindIdentAtPos(mod, args.Line, args.Col)
	if ident == nil {
		return textResult(encodeJSON(map[string]any{"found": false}))
	}

	sym := analysis.FindDefinition(st, ident.Name)
	if sym == nil {
		return textResult(encodeJSON(map[string]any{"found": false}))
	}

	return textResult(encodeJSON(map[string]any{
		"found": true,
		"name":  sym.Name,
		"kind":  sym.Kind.String(),
		"line":  sym.Line,
		"col":   sym.Col,
		"export": sym.Export,
	}))
}

func (s *Server) handleHover(raw json.RawMessage) ToolCallResult {
	var args struct {
		Text string `json:"text"`
		Line int    `json:"line"`
		Col  int    `json:"col"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		return errorResult("Parse errors: " + p.Errors()[0].Message)
	}

	st := analysis.BuildSymbolTable(mod)
	ident := analysis.FindIdentAtPos(mod, args.Line, args.Col)
	if ident == nil {
		return textResult(encodeJSON(map[string]any{"found": false}))
	}

	sym := analysis.FindDefinition(st, ident.Name)
	if sym == nil {
		return textResult(encodeJSON(map[string]any{"found": false}))
	}

	exportTag := "Локальная"
	if sym.Export {
		exportTag = "Экспорт"
	}

	hover := fmt.Sprintf("%s — %s (%s)", sym.Name, sym.Kind.String(), exportTag)

	return textResult(encodeJSON(map[string]any{
		"found": true,
		"name":  sym.Name,
		"hover": hover,
	}))
}

func (s *Server) handleFoldingRanges(raw json.RawMessage) ToolCallResult {
	var args parseArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return errorResult("Invalid arguments: " + err.Error())
	}

	p := parser.NewParser(args.Text)
	mod := p.ParseModule()
	if len(p.Errors()) > 0 {
		return errorResult("Parse errors: " + p.Errors()[0].Message)
	}

	st := analysis.BuildSymbolTable(mod)
	ranges := analysis.CollectFoldingRanges(mod, st)

	type foldJSON struct {
		StartLine int    `json:"startLine"`
		EndLine   int    `json:"endLine"`
		Kind      string `json:"kind"`
	}

	items := make([]foldJSON, len(ranges))
	for i, r := range ranges {
		items[i] = foldJSON{StartLine: r.StartLine, EndLine: r.EndLine, Kind: r.Kind}
	}

	return textResult(encodeJSON(map[string]any{"ranges": items}))
}

func encodeJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func textResult(text string) ToolCallResult {
	return ToolCallResult{
		Content: []ContentItem{
			{Type: "text", Text: text},
		},
	}
}

func errorResult(msg string) ToolCallResult {
	return ToolCallResult{
		Content: []ContentItem{
			{Type: "text", Text: msg},
		},
		IsError: true,
	}
}

func astToJSON(mod *parser.Module) map[string]any {
	return map[string]any{
		"type":       "Module",
		"directives": len(mod.Directives),
		"statements": len(mod.Statements),
	}
}
