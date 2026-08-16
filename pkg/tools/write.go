package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WriteTool struct {
	WorkingDir string
}

func NewWriteTool(workingDir string) *WriteTool {
	return &WriteTool{WorkingDir: workingDir}
}

// Deprecated alias for backwards compatibility
func NewwriteTool(workingDir string) *WriteTool {
	return NewWriteTool(workingDir)
}

func (w *WriteTool) Name() string {
	return "write_file"
}

func (w *WriteTool) Description() string {
	return "Write or overwrite content to a file. Format: <filename>\\n<content>"
}

func (w *WriteTool) Execute(args string) (string, error) {
	if args == "" {
		return "", fmt.Errorf("write_file requires: <path>\\n<content>")
	}

	path, content, found := strings.Cut(args, "\n")
	path = strings.TrimSpace(path)
	if !found {
		content = ""
	}
	if path == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	targetPath := path
	if !filepath.IsAbs(path) && w.WorkingDir != "" {
		targetPath = filepath.Join(w.WorkingDir, path)
	}

	err := os.WriteFile(targetPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write to '%s': %w", targetPath, err)
	}

	return fmt.Sprintf("Successfully wrote to %s", targetPath), nil
}
