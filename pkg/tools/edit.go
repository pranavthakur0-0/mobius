package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// editArgs defines the JSON arguments for the edit_file tool.
type editArgs struct {
	Path               string `json:"path"`
	TargetContent      string `json:"target_content"`
	ReplacementContent string `json:"replacement_content"`
}

// EditTool performs targeted, exact string replacement in files.
type EditTool struct {
	workingDir string
}

// NewEditFileTool creates a new EditTool bound to the specified working directory.
func NewEditFileTool(workingDir string) *EditTool {
	return &EditTool{workingDir: workingDir}
}

// Name returns the tool identifier "edit_file".
func (e *EditTool) Name() string {
	return "edit_file"
}

// Description returns a helpful summary of the tool for the LLM.
func (e *EditTool) Description() string {
	return "Performs exact string replacement in a file. target_content must match exactly one occurrence."
}

// Schema returns the JSON schema parameter specification for edit_file.
func (e *EditTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the file to modify.",
			},
			"target_content": {
				Type:        "string",
				Description: "The exact character sequence in the file to be replaced.",
			},
			"replacement_content": {
				Type:        "string",
				Description: "The new content to replace the target content with.",
			},
		},
		Required: []string{"path", "target_content", "replacement_content"},
	}
}

// Execute validates that target_content is uniquely present in the file, then replaces it.
func (e *EditTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var args editArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("invalid json arguments for edit_file: %w", err)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	targetPath := path
	if !filepath.IsAbs(path) && e.workingDir != "" {
		targetPath = filepath.Join(e.workingDir, path)
	}

	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", targetPath, err)
	}
	content := string(contentBytes)

	// Ensure target_content exists and is unique
	count := strings.Count(content, args.TargetContent)
	if count == 0 {
		return "", fmt.Errorf("target_content was not found in '%s'", targetPath)
	}
	if count > 1 {
		return "", fmt.Errorf("target_content was found %d times in '%s' — please provide more surrounding context to be unique", count, targetPath)
	}

	newContent := strings.Replace(content, args.TargetContent, args.ReplacementContent, 1)
	if err := os.WriteFile(targetPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to save changes to '%s': %w", targetPath, err)
	}

	return fmt.Sprintf("Successfully edited '%s'", targetPath), nil
}
