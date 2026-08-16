package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ListDirTool struct {
	workingDir string
}

func NewListTool(workingDir string) *ListDirTool {
	return &ListDirTool{workingDir: workingDir}
}

func (l *ListDirTool) Name() string {
	return "list_dir"
}
func (l *ListDirTool) Description() string {
	return "Lists files and folders in a directory with file sizes. Usage: list_dir <path> (or empty for current dir)"
}

func (l *ListDirTool) Execute(path string) (string, error) {
	path = strings.TrimSpace(path)
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
