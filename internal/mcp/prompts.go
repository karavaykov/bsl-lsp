package mcp

import "encoding/json"

func (s *Server) registerPrompts() {
	s.addPrompt(Prompt{
		Name:        "review_bsl_code",
		Description: "Review BSL code for errors and style issues",
		Arguments: []PromptArgument{
			{
				Name:        "code",
				Description: "BSL source code to review",
				Required:    true,
			},
		},
	})

	s.addPrompt(Prompt{
		Name:        "explain_bsl_module",
		Description: "Explain the structure of a BSL module",
		Arguments: []PromptArgument{
			{
				Name:        "code",
				Description: "BSL source code to explain",
				Required:    true,
			},
		},
	})
}

func (s *Server) buildReviewPrompt(raw json.RawMessage) PromptGetResult {
	var args struct {
		Code string `json:"code"`
	}
	json.Unmarshal(raw, &args)

	lintResult := s.handleLint(mustMarshal(map[string]string{"text": args.Code}))

	return PromptGetResult{
		Messages: []PromptMessage{
			{
				Role: "system",
				Content: ContentItem{
					Type: "text",
					Text: "You are a BSL (1C:Enterprise) code reviewer. Analyze the provided code and suggest improvements.",
				},
			},
			{
				Role: "user",
				Content: ContentItem{
					Type: "text",
					Text: "Please review this BSL code:\n\n```bsl\n" + args.Code + "\n```\n\nLinter diagnostics:\n" + lintResult.Content[0].Text,
				},
			},
		},
	}
}

func (s *Server) buildExplainPrompt(raw json.RawMessage) PromptGetResult {
	var args struct {
		Code string `json:"code"`
	}
	json.Unmarshal(raw, &args)

	symResult := s.handleSymbols(mustMarshal(map[string]string{"text": args.Code}))

	return PromptGetResult{
		Messages: []PromptMessage{
			{
				Role: "system",
				Content: ContentItem{
					Type: "text",
					Text: "You are a BSL (1C:Enterprise) expert. Explain the structure of the provided module.",
				},
			},
			{
				Role: "user",
				Content: ContentItem{
					Type: "text",
					Text: "Explain the structure of this BSL module:\n\n```bsl\n" + args.Code + "\n```\n\nSymbols found:\n" + symResult.Content[0].Text,
				},
			},
		},
	}
}

func mustMarshal(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
