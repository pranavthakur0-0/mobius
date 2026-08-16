package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GrepTool struct {
	workingDir string
}

func NewGrepTool(workingDir string) *GrepTool {
	return &GrepTool{workingDir: workingDir}
}

func (g *GrepTool) Name() string {
	return "grep_search"
}

func (g *GrepTool) Description() string {
	return "Searches for a text pattern or regex across files. Usage: grep_search <pattern> [directory]"
}

func (g *GrepTool) Execute(args string) (string, error) {
	args = strings.TrimSpace(args)
	if args == "" {
		return "", fmt.Errorf("grep_search requires: <pattern> [directory]")
	}

	parts := strings.Fields(args)
	pattern := parts[0]
	searchPath := "."
	if len(parts) > 1 {
		searchPath = parts[1]
	}

	targetDir := searchPath
	if !filepath.IsAbs(searchPath) && g.workingDir != "" {
		targetDir = filepath.Join(g.workingDir, searchPath)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regular expression '%s': %w", pattern, err)
	}

	var results []string
	maxResults := 50

	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		".agents":      true,
		".gemini":      true,
	}

	err = filepath.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip unreadable paths
		}

		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip known non-text/binary extensions
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf", ".zip", ".tar", ".gz", ".exe", ".bin":
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		relPath, _ := filepath.Rel(g.workingDir, path)
		if relPath == "" {
			relPath = path
		}

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", relPath, lineNum, strings.TrimSpace(line)))
				if len(results) >= maxResults {
					results = append(results, fmt.Sprintf("... (capped at %d results)", maxResults))
					return filepath.SkipAll
				}
			}
			lineNum++
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("grep search failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No matches found for '%s' in '%s'", pattern, targetDir), nil
	}

	return strings.Join(results, "\n"), nil
}
