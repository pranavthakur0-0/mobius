package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// listArgs defines the JSON arguments for the list_dir tool.
type listArgs struct {
	Path string `json:"path"`
}

// ListDirTool lists the directory contents and file sizes.
type ListDirTool struct {
	workingDir string
}

// NewListTool creates a new ListDirTool bound to the specified working directory.
func NewListTool(workingDir string) *ListDirTool {
	return &ListDirTool{workingDir: workingDir}
}

// Name returns the tool identifier "list_dir".
func (l *ListDirTool) Name() string {
	return "list_dir"
}

// Description returns a helpful summary of the tool for the LLM.
func (l *ListDirTool) Description() string {
	return "Lists files and directories in a given path."
}

// Schema returns the JSON schema parameter specification for list_dir.
func (l *ListDirTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Optional directory path to list. Defaults to current directory if omitted.",
			},
		},
	}
}

// Execute reads the directory and formats entry names and byte sizes.
func (l *ListDirTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var args listArgs
	if strings.TrimSpace(jsonArgs) != "" {
		_ = json.Unmarshal([]byte(jsonArgs), &args)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		path = "."
	}

	targetPath := path
	if !filepath.IsAbs(path) && l.workingDir != "" {
		targetPath = filepath.Join(l.workingDir, path)
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to list directory '%s': %w", targetPath, err)
	}

	if len(entries) == 0 {
		return fmt.Sprintf("Directory '%s' is empty", targetPath), nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Contents of %s:\n", targetPath))

	for _, entry := range entries {
		if entry.IsDir() {
			builder.WriteString(fmt.Sprintf("  [DIR]   %s/\n", entry.Name()))
		} else {
			info, err := entry.Info()
			if err != nil {
				builder.WriteString(fmt.Sprintf("  [FILE]  %s\n", entry.Name()))
			} else {
				builder.WriteString(fmt.Sprintf("  [FILE]  %-25s (%d bytes)\n", entry.Name(), info.Size()))
			}
		}
	}

	return builder.String(), nil
}
