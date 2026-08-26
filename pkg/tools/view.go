package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// viewFileArgs defines the JSON arguments for the view_file tool.
type viewFileArgs struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// ViewFileTool reads and displays file contents with line numbers and line-range slicing.
type ViewFileTool struct {
	WorkingDir string
}

// NewViewFileTool creates a new ViewFileTool bound to the specified working directory.
func NewViewFileTool(workDir string) *ViewFileTool {
	return &ViewFileTool{
		WorkingDir: workDir,
	}
}

// Name returns the tool identifier "view_file".
func (v *ViewFileTool) Name() string {
	return "view_file"
}

// Description returns a helpful summary of the tool for the LLM.
func (v *ViewFileTool) Description() string {
	return "Reads the contents of a file with line numbers and optional line range slicing."
}

// Schema returns the JSON schema parameter specification for view_file.
func (v *ViewFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"path": {
				Type:        "string",
				Description: "Relative path to the file to inspect.",
			},
			"start_line": {
				Type:        "integer",
				Description: "Optional 1-indexed starting line number.",
			},
			"end_line": {
				Type:        "integer",
				Description: "Optional 1-indexed ending line number.",
			},
		},
		Required: []string{"path"},
	}
}

// Execute reads the file from disk and formats its contents with 1-based line numbers.
func (v *ViewFileTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var args viewFileArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("invalid json arguments for view_file: %w", err)
	}

	path := strings.TrimSpace(args.Path)
	if path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	targetPath := path
	if !filepath.IsAbs(path) && v.WorkingDir != "" {
		targetPath = filepath.Join(v.WorkingDir, path)
	}

	file, err := os.Open(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %w", targetPath, err)
	}
	defer file.Close()

	var builder strings.Builder
	scanner := bufio.NewScanner(file)

	startLine := args.StartLine
	if startLine <= 0 {
		startLine = 1
	}

	endLine := args.EndLine
	currentLine := 1
	linesWritten := 0
	maxLimit := 800

	for scanner.Scan() {
		if currentLine >= startLine && (endLine <= 0 || currentLine <= endLine) {
			line := scanner.Text()
			builder.WriteString(fmt.Sprintf("%d: %s\n", currentLine, line))
			linesWritten++
			if linesWritten >= maxLimit {
				builder.WriteString(fmt.Sprintf("\n... [Output capped at %d lines]\n", maxLimit))
				break
			}
		}

		if endLine > 0 && currentLine >= endLine {
			break
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file '%s': %w", targetPath, err)
	}

	return builder.String(), nil
}
