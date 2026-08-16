package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// writeArgs defines the JSON arguments for the write_file tool.
type writeArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// WriteTool writes or completely overwrites file contents on disk.
type WriteTool struct {
	WorkingDir string
}

// NewWriteTool creates a new WriteTool bound to the specified working directory.
func NewWriteTool(workingDir string) *WriteTool {
	return &WriteTool{WorkingDir: workingDir}
}

// Name returns the tool identifier "write_file".
func (w *WriteTool) Name() string {
	return "write_file"
}

// Description returns a helpful summary of the tool for the LLM.
func (w *WriteTool) Description() string {
	return "Writes or overwrites full content to a file. Automatically creates parent directories if needed."
}

// Schema returns the JSON schema parameter specification for write_file.
func (w *WriteTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "The relative path of the file to write to.",
			},
			"content": {
				Type:        "string",
				Description: "The full text content to write into the file.",
			},
		},
		Required: []string{"path", "content"},
	}
}

// Execute writes the given content to the target file, creating parent directories if necessary.
func (w *WriteTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var req writeArgs
	if err := json.Unmarshal([]byte(jsonArgs), &req); err != nil {
		return "", fmt.Errorf("invalid json arguments for write_file: %w", err)
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	targetPath := path
	if !filepath.IsAbs(path) && w.WorkingDir != "" {
		targetPath = filepath.Join(w.WorkingDir, path)
	}

	// Auto-create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directories: %w", err)
	}

	if err := os.WriteFile(targetPath, []byte(req.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write to '%s': %w", targetPath, err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(req.Content), targetPath), nil
}
