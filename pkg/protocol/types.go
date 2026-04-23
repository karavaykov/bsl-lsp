package protocol

import "encoding/json"

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity,omitempty"`
	Message  string             `json:"message"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type InitializeParams struct {
	ProcessID int `json:"processId"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync            int                     `json:"textDocumentSync"`
	DocumentSymbolProvider      bool                    `json:"documentSymbolProvider,omitempty"`
	DefinitionProvider          bool                    `json:"definitionProvider,omitempty"`
	HoverProvider               bool                    `json:"hoverProvider,omitempty"`
	CompletionProvider          *CompletionOptions      `json:"completionProvider,omitempty"`
	SemanticTokensProvider      *SemanticTokensOptions  `json:"semanticTokensProvider,omitempty"`
	CodeLensProvider            *CodeLensOptions        `json:"codeLensProvider,omitempty"`
	FoldingRangeProvider        bool                    `json:"foldingRangeProvider,omitempty"`
	DocumentFormattingProvider  bool                    `json:"documentFormattingProvider,omitempty"`
	SignatureHelpProvider       *SignatureHelpOptions   `json:"signatureHelpProvider,omitempty"`
}

type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

type CodeLensOptions struct{}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type TextDocumentPositionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type CompletionParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
	Context  *struct {
		TriggerKind      int    `json:"triggerKind"`
		TriggerCharacter string `json:"triggerCharacter,omitempty"`
	} `json:"context,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          int                `json:"kind,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Documentation string             `json:"documentation,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	TextEdit      *CompletionTextEdit `json:"textEdit,omitempty"`
}

type CompletionTextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

const (
	CompletionKindText          = 1
	CompletionKindMethod        = 2
	CompletionKindFunction      = 3
	CompletionKindConstructor   = 4
	CompletionKindField         = 5
	CompletionKindVariable      = 6
	CompletionKindClass         = 7
	CompletionKindInterface     = 8
	CompletionKindModule        = 9
	CompletionKindProperty      = 10
	CompletionKindUnit          = 11
	CompletionKindValue         = 12
	CompletionKindEnum          = 13
	CompletionKindKeyword       = 14
	CompletionKindSnippet       = 15
	CompletionKindColor         = 16
	CompletionKindFile          = 17
	CompletionKindReference     = 18
	CompletionKindFolder        = 19
	CompletionKindEnumMember    = 20
	CompletionKindConstant      = 21
	CompletionKindStruct        = 22
	CompletionKindEvent         = 23
	CompletionKindOperator      = 24
	CompletionKindTypeParameter = 25
)

type StdoutWriter struct{}

func (StdoutWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type DocumentSymbolParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string            `json:"name"`
	Detail         string            `json:"detail,omitempty"`
	Kind           int               `json:"kind"`
	Tags           []int             `json:"tags,omitempty"`
	Range          Range             `json:"range"`
	SelectionRange Range             `json:"selectionRange"`
	Children       []DocumentSymbol  `json:"children,omitempty"`
}

const (
	SymbolKindFile        = 1
	SymbolKindModule      = 2
	SymbolKindNamespace   = 3
	SymbolKindPackage     = 4
	SymbolKindClass       = 5
	SymbolKindMethod      = 6
	SymbolKindProperty    = 7
	SymbolKindField       = 8
	SymbolKindConstructor = 9
	SymbolKindEnum        = 10
	SymbolKindInterface   = 11
	SymbolKindFunction    = 12
	SymbolKindVariable    = 13
	SymbolKindConstant    = 14
	SymbolKindString      = 15
	SymbolKindNumber      = 16
	SymbolKindBoolean     = 17
	SymbolKindArray       = 18
	SymbolKindObject      = 19
	SymbolKindKey         = 20
	SymbolKindNull        = 21
	SymbolKindEnumMember  = 22
	SymbolKindStruct      = 23
	SymbolKindEvent       = 24
	SymbolKindOperator    = 25
	SymbolKindTypeParameter = 26
)

type SemanticTokensParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type SemanticTokens struct {
	Data []int `json:"data"`
}

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type CodeLensParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type CodeLens struct {
	Range   Range           `json:"range"`
	Command *Command        `json:"command,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type Command struct {
	Title     string          `json:"title"`
	Command   string          `json:"command"`
	Arguments []json.RawMessage `json:"arguments,omitempty"`
}

type FoldingRangeParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type FoldingRange struct {
	StartLine      int    `json:"startLine"`
	StartCharacter int    `json:"startCharacter,omitempty"`
	EndLine        int    `json:"endLine"`
	EndCharacter   int    `json:"endCharacter,omitempty"`
	Kind           string `json:"kind,omitempty"`
}

type DocumentFormattingParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Options FormattingOptions `json:"options"`
}

type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type SignatureHelpParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position Position `json:"position"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}
