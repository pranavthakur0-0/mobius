package llm

import (
	"context"
)


type Role string


const (
	RoleSystem		Role = "system"
	RoleUser		Role = "user"
	RoleAssistant	Role = "assistant"
	RoleTool 		Role = "tool"
)


type ToolCall struct {
	ID		string `json:"id"`
	Name	string `json:"name"`
	Arguments	string `json:"arguments"`
}

type Message struct {
	Role 	Role 	`json:"role"`
	Content	string	`json:"content,omitempty"`
	ToolCalls	[]ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}


type ChatRequest struct {
	Model	string 	`json:"model"`
	Message []Message	`json:"messages"`
}


type ChatResponse struct {
	Content	string `json:"content"`
}

type Provider interface {
	Name()	string
	Generate(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}



