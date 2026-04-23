package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Server struct {
	mu      sync.Mutex
	initialized bool
	clientInfo  Implementation
	tools     []Tool
	resources []Resource
	prompts   []Prompt
	toolHandlers map[string]func(json.RawMessage) ToolCallResult
	resourceStore *ResourceStore
}

func NewServer() *Server {
	s := &Server{
		toolHandlers: make(map[string]func(json.RawMessage) ToolCallResult),
		resourceStore: NewResourceStore(),
	}
	s.registerTools()
	s.registerResources()
	s.registerPrompts()
	return s
}

func (s *Server) Handle(req jsonRPCMessage) *jsonRPCMessage {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	case "resources/list":
		return s.handleResourcesList(req)
	case "resources/read":
		return s.handleResourcesRead(req)
	case "prompts/list":
		return s.handlePromptsList(req)
	case "prompts/get":
		return s.handlePromptsGet(req)
	default:
		if req.ID != nil {
			return s.jsonError(*req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
		}
		return nil
	}
}

func (s *Server) handleInitialize(req jsonRPCMessage) *jsonRPCMessage {
	var params InitializeRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.jsonError(*req.ID, -32602, "Invalid initialize params: "+err.Error())
	}

	s.mu.Lock()
	s.initialized = true
	s.clientInfo = params.ClientInfo
	s.mu.Unlock()

	return s.jsonResult(*req.ID, InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCapabilities{
			Tools:     &ToolsCapability{ListChanged: true},
			Resources: &ResourcesCapability{ListChanged: true},
			Prompts:   &PromptsCapability{ListChanged: true},
		},
		ServerInfo: Implementation{
			Name:    "bsl-lsp-mcp",
			Version: "1.0.0",
		},
	})
}

func (s *Server) handleToolsList(req jsonRPCMessage) *jsonRPCMessage {
	s.mu.Lock()
	tools := make([]Tool, len(s.tools))
	copy(tools, s.tools)
	s.mu.Unlock()
	return s.jsonResult(*req.ID, map[string]any{"tools": tools})
}

func (s *Server) handleToolsCall(req jsonRPCMessage) *jsonRPCMessage {
	var params ToolCallRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.jsonError(*req.ID, -32602, "Invalid tool call params: "+err.Error())
	}

	s.mu.Lock()
	handler, ok := s.toolHandlers[params.Name]
	s.mu.Unlock()
	if !ok {
		return s.jsonError(*req.ID, -32602, fmt.Sprintf("Tool not found: %s", params.Name))
	}

	result := handler(params.Arguments)
	return s.jsonResult(*req.ID, result)
}

func (s *Server) handleResourcesList(req jsonRPCMessage) *jsonRPCMessage {
	s.mu.Lock()
	resources := make([]Resource, len(s.resources))
	copy(resources, s.resources)
	s.mu.Unlock()
	return s.jsonResult(*req.ID, map[string]any{"resources": resources})
}

func (s *Server) handleResourcesRead(req jsonRPCMessage) *jsonRPCMessage {
	var params ResourceReadRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.jsonError(*req.ID, -32602, "Invalid resource read params: "+err.Error())
	}

	content, ok := s.resourceStore.Get(params.URI)
	if !ok {
		return s.jsonError(*req.ID, -32602, fmt.Sprintf("Resource not found: %s", params.URI))
	}

	return s.jsonResult(*req.ID, ResourceReadResult{
		Contents: []ResourceContent{content},
	})
}

func (s *Server) handlePromptsList(req jsonRPCMessage) *jsonRPCMessage {
	s.mu.Lock()
	prompts := make([]Prompt, len(s.prompts))
	copy(prompts, s.prompts)
	s.mu.Unlock()
	return s.jsonResult(*req.ID, map[string]any{"prompts": prompts})
}

func (s *Server) handlePromptsGet(req jsonRPCMessage) *jsonRPCMessage {
	var params PromptGetRequest
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.jsonError(*req.ID, -32602, "Invalid prompt get params: "+err.Error())
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var result PromptGetResult
	switch params.Name {
	case "review_bsl_code":
		result = s.buildReviewPrompt(params.Arguments)
	case "explain_bsl_module":
		result = s.buildExplainPrompt(params.Arguments)
	default:
		return s.jsonError(*req.ID, -32602, fmt.Sprintf("Prompt not found: %s", params.Name))
	}

	return s.jsonResult(*req.ID, result)
}

func (s *Server) registerTool(t Tool, handler func(json.RawMessage) ToolCallResult) {
	s.tools = append(s.tools, t)
	s.toolHandlers[t.Name] = handler
}

func (s *Server) addResource(r Resource) {
	s.resources = append(s.resources, r)
}

func (s *Server) addPrompt(p Prompt) {
	s.prompts = append(s.prompts, p)
}

func (s *Server) jsonResult(id int, result any) *jsonRPCMessage {
	data, _ := json.Marshal(result)
	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &id,
		Result:  data,
	}
}

func (s *Server) jsonError(id int, code int, msg string) *jsonRPCMessage {
	return &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      &id,
		Error: &jsonRPCError{
			Code:    code,
			Message: msg,
		},
	}
}
