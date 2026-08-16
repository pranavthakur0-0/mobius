package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type EditTool struct {
	workingDir string
}

func NewEditFileTool(workingDir string) *EditTool {
	return &EditTool{workingDir: workingDir}
}

func NewEditTool(workingDir string) *EditTool {
	return NewEditFileTool(workingDir)
}

func (e *EditTool) Name() string {
	return "edit_file"
}

func (e *EditTool) Description() string {
	return "Performs exact string replacement in a file. Requires format:\n<filepath>\n<<<OLD>>>\n<exact old text to find>\n<<<NEW>>>\n<replacement text>"
}

func (e *EditTool) Execute(args string) (string, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return "", fmt.Errorf("edit_file requires: <filepath>\n<<<OLD>>>\n<old_text>\n<<<NEW>>>\n<new_text>")
	}

	// 1. Extract the filepath (first line)
	filePath, rest, found := strings.Cut(args, "\n")
	if !found {
		return "", fmt.Errorf("invalid format: expected filepath on the first line")
	}
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// 2. Extract <<<NEW>>> delimiter
	beforeNew, newText, foundNew := strings.Cut(rest, "<<<NEW>>>\n")
	if !foundNew {
		beforeNew, newText, foundNew = strings.Cut(rest, "<<<NEW>>>")
		if !foundNew {
			return "", fmt.Errorf("missing '<<<NEW>>>' delimiter")
		}
	}

	// 3. Extract <<<OLD>>> delimiter
	_, oldText, foundOld := strings.Cut(beforeNew, "<<<OLD>>>\n")
	if !foundOld {
		_, oldText, foundOld = strings.Cut(beforeNew, "<<<OLD>>>")
		if !foundOld {
			return "", fmt.Errorf("missing '<<<OLD>>>' delimiter")
		}
	}

	// 4. Resolve file path against workingDir
	targetPath := filePath
	if !filepath.IsAbs(filePath) && e.workingDir != "" {
		targetPath = filepath.Join(e.workingDir, filePath)
	}

	// 5. Read the existing file into memory
	contentBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", targetPath, err)
	}
	content := string(contentBytes)

	// 6. Safety check: Ensure oldText exists and is unique
	count := strings.Count(content, oldText)
	if count == 0 {
		return "", fmt.Errorf("target old text was not found in '%s'", targetPath)
	}
	if count > 1 {
		return "", fmt.Errorf("target old text was found %d times in '%s' - please provide more surrounding context to be unique", count, targetPath)
	}

	// 7. Perform replacement and save back to disk
	newContent := strings.Replace(content, oldText, newText, 1)
	err = os.WriteFile(targetPath, []byte(newContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save changes to '%s': %w", targetPath, err)
	}

	return fmt.Sprintf("Successfully edited '%s'", targetPath), nil
}
