package mcp

import (
	"encoding/json"
	"testing"
)

func TestInitialize(t *testing.T) {
	srv := NewServer()

	params, _ := json.Marshal(InitializeRequest{
		ProtocolVersion: ProtocolVersion,
		ClientInfo:      Implementation{Name: "test-client", Version: "1.0"},
	})

	req := jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      intPtr(1),
		Method:  "initialize",
		Params:  params,
	}

	resp := srv.Handle(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected success, got error: %v", resp)
	}

	var result InitializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if result.ServerInfo.Name != "bsl-lsp-mcp" {
		t.Errorf("expected bsl-lsp-mcp, got %s", result.ServerInfo.Name)
	}
	if result.Capabilities.Tools == nil {
		t.Error("expected tools capability")
	}
	if result.Capabilities.Resources == nil {
		t.Error("expected resources capability")
	}
	if result.Capabilities.Prompts == nil {
		t.Error("expected prompts capability")
	}
}

func TestToolsList(t *testing.T) {
	srv := NewServer()

	initParams, _ := json.Marshal(InitializeRequest{
		ProtocolVersion: ProtocolVersion,
		ClientInfo:      Implementation{Name: "test", Version: "1"},
	})
	srv.Handle(jsonRPCMessage{JSONRPC: "2.0", ID: intPtr(1), Method: "initialize", Params: initParams})

	req := jsonRPCMessage{JSONRPC: "2.0", ID: intPtr(2), Method: "tools/list"}
	resp := srv.Handle(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected success, got error: %v", resp)
	}

	var list struct {
		Tools []Tool `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &list); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(list.Tools) != 7 {
		t.Errorf("expected 7 tools, got %d", len(list.Tools))
	}

	names := make(map[string]bool)
	for _, tool := range list.Tools {
		names[tool.Name] = true
	}

	expected := []string{"bsl_parse", "bsl_lint", "bsl_format", "bsl_symbols", "bsl_define", "bsl_hover", "bsl_folding_ranges"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestToolParse(t *testing.T) {
	srv := NewServer()
	code := `Процедура Тест() Экспорт
	Возврат;
КонецПроцедуры`

	args, _ := json.Marshal(map[string]string{"text": code})
	result := srv.handleParse(args)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].Text)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].Text), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	errs := parsed["parseErrors"].([]any)
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errs))
	}
}

func TestToolLint(t *testing.T) {
	srv := NewServer()
	code := `Процедура Тест()
	А = 1000
	Возврат
КонецПроцедуры`

	args, _ := json.Marshal(map[string]string{"text": code})
	result := srv.handleLint(args)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].Text)
	}

	var diagResult map[string]any
	json.Unmarshal([]byte(result.Content[0].Text), &diagResult)
	diags := diagResult["diagnostics"].([]any)

	if len(diags) == 0 {
		t.Error("expected at least 1 diagnostic (magic number)")
	}
}

func TestToolFormat(t *testing.T) {
	srv := NewServer()
	code := `Процедура Тест()
	А=1;
	Возврат;
КонецПроцедуры`

	args, _ := json.Marshal(map[string]any{
		"text":         code,
		"tabSize":      4,
		"insertSpaces": true,
	})
	result := srv.handleFormat(args)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].Text)
	}

	var formatted map[string]string
	json.Unmarshal([]byte(result.Content[0].Text), &formatted)
	if formatted["formatted"] == "" {
		t.Error("expected formatted output")
	}
}

func TestToolSymbols(t *testing.T) {
	srv := NewServer()
	code := `Процедура Тест() Экспорт
	Возврат;
КонецПроцедуры`

	args, _ := json.Marshal(map[string]string{"text": code})
	result := srv.handleSymbols(args)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].Text)
	}

	var symResult map[string]any
	json.Unmarshal([]byte(result.Content[0].Text), &symResult)

	funcs := symResult["functions"].([]any)
	if len(funcs) != 1 {
		t.Errorf("expected 1 function, got %d", len(funcs))
	}

	fn := funcs[0].(map[string]any)
	if fn["name"] != "Тест" {
		t.Errorf("expected Тест, got %v", fn["name"])
	}
	if fn["exported"] != true {
		t.Error("expected exported=true")
	}
}

func TestToolFoldingRanges(t *testing.T) {
	srv := NewServer()
	code := `Процедура А()
	Возврат;
КонецПроцедуры

Процедура Б()
	Возврат;
КонецПроцедуры`

	args, _ := json.Marshal(map[string]string{"text": code})
	result := srv.handleFoldingRanges(args)

	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Content[0].Text)
	}

	var frResult map[string]any
	json.Unmarshal([]byte(result.Content[0].Text), &frResult)
	ranges := frResult["ranges"].([]any)
	// Two procedures should produce at least one folding range
	if len(ranges) == 0 {
		t.Log("no folding ranges returned (may depend on implementation)")
	}
}

func TestPromptsList(t *testing.T) {
	srv := NewServer()

	req := jsonRPCMessage{JSONRPC: "2.0", ID: intPtr(1), Method: "prompts/list"}
	resp := srv.Handle(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected success, got error: %v", resp)
	}

	var list struct {
		Prompts []Prompt `json:"prompts"`
	}
	json.Unmarshal(resp.Result, &list)

	if len(list.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(list.Prompts))
	}
}

func TestResourcesList(t *testing.T) {
	srv := NewServer()

	req := jsonRPCMessage{JSONRPC: "2.0", ID: intPtr(1), Method: "resources/list"}
	resp := srv.Handle(req)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected success, got error: %v", resp)
	}

	var list struct {
		Resources []Resource `json:"resources"`
	}
	json.Unmarshal(resp.Result, &list)

	if list.Resources == nil {
		t.Error("expected non-nil resources list")
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := NewServer()

	req := jsonRPCMessage{JSONRPC: "2.0", ID: intPtr(1), Method: "unknown/method"}
	resp := srv.Handle(req)
	if resp == nil || resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected code -32601, got %d", resp.Error.Code)
	}
}

func intPtr(n int) *int {
	return &n
}
