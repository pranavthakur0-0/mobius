package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// grepArgs defines the JSON arguments for the grep_search tool.
type grepArgs struct {
	Pattern         string `json:"pattern"`
	Path            string `json:"path"`
	CaseInsensitive bool   `json:"case_insensitive"`
}

// GrepTool searches for regex or text patterns across files in a directory tree.
type GrepTool struct {
	workingDir string
}

// NewGrepTool creates a new GrepTool bound to the specified working directory.
func NewGrepTool(workingDir string) *GrepTool {
	return &GrepTool{workingDir: workingDir}
}

// Name returns the tool identifier "grep_search".
func (g *GrepTool) Name() string {
	return "grep_search"
}

// Description returns a helpful summary of the tool for the LLM.
func (g *GrepTool) Description() string {
	return "Searches for a text pattern or regex across files in a directory."
}

// Schema returns the JSON schema parameter specification for grep_search.
func (g *GrepTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"pattern": {
				Type:        "string",
				Description: "The search term or regular expression pattern.",
			},
			"path": {
				Type:        "string",
				Description: "Optional relative directory to search within. Defaults to current directory.",
			},
			"case_insensitive": {
				Type:        "boolean",
				Description: "Optional flag to perform case-insensitive search.",
			},
		},
		Required: []string{"pattern"},
	}
}

// Execute walks the directory tree matching lines against the pattern and returns line-numbered results.
func (g *GrepTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var args grepArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("invalid json arguments for grep_search: %w", err)
	}

	pattern := strings.TrimSpace(args.Pattern)
	if pattern == "" {
		return "", fmt.Errorf("search pattern cannot be empty")
	}

	searchPath := strings.TrimSpace(args.Path)
	if searchPath == "" {
		searchPath = "."
	}

	targetDir := searchPath
	if !filepath.IsAbs(searchPath) && g.workingDir != "" {
		targetDir = filepath.Join(g.workingDir, searchPath)
	}

	if args.CaseInsensitive {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex '%s': %w", pattern, err)
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
			return nil
		}

		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

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
					results = append(results, fmt.Sprintf("... [capped at %d matches]", maxResults))
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
