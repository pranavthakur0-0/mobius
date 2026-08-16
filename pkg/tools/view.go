package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ViewFileTool struct {
	WorkingDir string
}

func NewViewFileTool(workDir string) *ViewFileTool {
	return &ViewFileTool{
		WorkingDir: workDir,
	}
}

func (v *ViewFileTool) Name() string {
	return "view_file"
}

func (v *ViewFileTool) Description() string {
	return "Reads the contents of a file with line numbers. Argument: <filename>"
}

func (v *ViewFileTool) Execute(path string) (string, error) {
	path = strings.TrimSpace(path)
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
	lineNum := 1

	for scanner.Scan() {
		builder.WriteString(fmt.Sprintf("%d: %s\n", lineNum, scanner.Text()))
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file '%s': %w", targetPath, err)
	}

	return builder.String(), nil
}
