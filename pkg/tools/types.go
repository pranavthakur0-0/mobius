package tools

import (
	"context"
)

// Property defines a single parameter field in the JSON Schema.
type Property struct {
	Type        string `json:"type"`        // Data type: "string", "integer", "boolean", "array", "object"
	Description string `json:"description"` // Explains to the LLM what this argument does and how to format it
}

// ToolSchema describes the JSON structure of arguments the tool accepts.
type ToolSchema struct {
	Type       string              `json:"type"`               // Almost always "object" for tool parameter specs
	Properties map[string]Property `json:"properties"`         // Map of parameter name to property definition
	Required   []string            `json:"required,omitempty"` // List of required parameter names
}

// Tool is the standard interface every agent tool must implement.
type Tool interface {
	// Name returns the unique identifier for the tool (e.g. "view_file", "run_command").
	Name() string

	// Description explains to the LLM what the tool does and when to use it.
	Description() string

	// Schema returns the JSON schema describing the expected input parameters.
	Schema() ToolSchema

	// Execute runs the tool with the given context and JSON-encoded arguments string.
	Execute(ctx context.Context, jsonArgs string) (string, error)
}
