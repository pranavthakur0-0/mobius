package tools

import (
	"fmt"
	"sort"
)

// Registry manages the set of available tools for the agent harness.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry initializes an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by its unique name.
func (r *Registry) Get(name string) (Tool, error) {
	if name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool '%s' not found in registry", name)
	}
	return tool, nil
}

// List returns all registered tools sorted alphabetically by name.
func (r *Registry) List() []Tool {
	result := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		result = append(result, tool)
	}
	// Sort deterministically by tool name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result
}

// NewDefaultRegistry creates a registry pre-loaded with all standard tools.
func NewDefaultRegistry(workDir string) *Registry {
	r := NewRegistry()
	r.Register(NewViewFileTool(workDir))
	r.Register(NewWriteTool(workDir))
	r.Register(NewListTool(workDir))
	r.Register(NewBashTool(workDir))
	r.Register(NewEditFileTool(workDir))
	r.Register(NewGrepTool(workDir))
	return r
}
