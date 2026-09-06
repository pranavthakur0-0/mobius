package guides

import (
	"context"
	"fmt"
	"io/fs"
	"mobius/pkg/llm"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".mobius":      true,
	"bin":          true,
	"dist":         true,
	"build":        true,
	".idea":        true,
	".vscode":      true,
	"__pycache__":   true,
	".pytest_cache":true,
}

// Known manifest or configuration files to look out for during scanning.
var knownManifests = map[string]bool{
	"Makefile":         true,
	"Justfile":         true,
	"go.mod":           true,
	"package.json":     true,
	"Cargo.toml":       true,
	"pyproject.toml":   true,
	"requirements.txt": true,
	"setup.py":         true,
	"build.zig":        true,
	"CMakeLists.txt":   true,
	"pom.xml":          true,
	"build.gradle":     true,
	"Dockerfile":       true,
	"Containerfile":    true,
	"Procfile":         true,
}

// BuildInitPrompt constructs a structured prompt for the LLM based on detected files.
func BuildInitPrompt(manifests, extensions []string) string {
	manifestsStr := "(none detected)"
	if len(manifests) > 0 {
		manifestsStr = strings.Join(manifests, ", ")
	}

	extsStr := "(none detected)"
	if len(extensions) > 0 {
		extsStr = strings.Join(extensions, ", ")
	}

	return fmt.Sprintf(`You are an expert DevOps engineer and harness architect.
Analyze this codebase snapshot:
- Manifest/Config files found: %s
- File extensions present: %s

Identify the primary language/framework and provide the exact shell commands to build, test, and lint this project.
Your output must be formatted in markdown exactly as:
## Verification Commands
Build: <command>
Test: <command>
Lint: <command>

Rules:
1. Output ONLY the markdown block starting with "## Verification Commands". No preamble, no explanation, no backticks around the block.
2. If a command does not apply, omit that line.`,
		manifestsStr,
		extsStr,
	)
}

// ScanWorkspace inspects a directory and returns key manifests and unique file extensions.
func ScanWorkspace(workspaceDir string) (manifests []string, extensions []string, err error) {
	extMap := make(map[string]bool)
	manifestMap := make(map[string]bool)

	err = filepath.WalkDir(workspaceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // Skip unreadable paths
		}

		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()

		if knownManifests[name] {
			manifestMap[name] = true
		}

		ext := strings.ToLower(filepath.Ext(name))
		if ext != "" && len(ext) <= 6 {
			extMap[ext] = true
		}

		return nil
	})

	for m := range manifestMap {
		manifests = append(manifests, m)
	}
	for e := range extMap {
		extensions = append(extensions, e)
	}

	sort.Strings(manifests)
	sort.Strings(extensions)

	return manifests, extensions, err
}

// cleanGeneratedMarkdown trims surrounding code blocks if the LLM wrapped the output.
func cleanGeneratedMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 && strings.HasPrefix(lines[0], "```") {
			lines = lines[1:]
		}
		if len(lines) >= 1 && strings.HasPrefix(lines[len(lines)-1], "```") {
			lines = lines[:len(lines)-1]
		}
		s = strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return s
}

// appendOrUpdateVerificationCommands updates an existing ## Verification Commands section or appends it.
func appendOrUpdateVerificationCommands(existingContent, newCommands string) string {
	const header = "## Verification Commands"
	existingContent = strings.TrimSpace(existingContent)

	if !strings.Contains(existingContent, header) {
		if existingContent == "" {
			return newCommands + "\n"
		}
		return existingContent + "\n\n" + newCommands + "\n"
	}

	// If header already exists, replace the section from the header until the next "## " or EOF
	idx := strings.Index(existingContent, header)
	before := strings.TrimRight(existingContent[:idx], "\n")

	after := ""
	rest := existingContent[idx+len(header):]
	nextHeaderIdx := strings.Index(rest, "\n## ")
	if nextHeaderIdx != -1 {
		after = "\n" + rest[nextHeaderIdx+1:]
	}

	var result string
	if before != "" {
		result = before + "\n\n" + newCommands
	} else {
		result = newCommands
	}

	if after != "" {
		result += "\n" + strings.TrimLeft(after, "\n")
	} else {
		result += "\n"
	}

	return result
}

// InitWorkspace scans the directory, asks the LLM for commands, and writes them to AGENTS.md.
func InitWorkspace(ctx context.Context, p llm.Provider, model string, workspaceDir string) (string, error) {
	manifests, exts, err := ScanWorkspace(workspaceDir)
	if err != nil {
		return "", fmt.Errorf("failed to scan workspace: %w", err)
	}

	if len(manifests) == 0 && len(exts) == 0 {
		return "", fmt.Errorf("no recognizable source files or manifests found in %s", workspaceDir)
	}

	prompt := BuildInitPrompt(manifests, exts)

	req := &llm.ChatRequest{
		Model: model,
		Message: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		return "", fmt.Errorf("LLM failed to generate commands: %w", err)
	}

	generated := cleanGeneratedMarkdown(resp.Content)
	if generated == "" {
		return "", fmt.Errorf("LLM returned empty response")
	}

	// Update or create AGENTS.md
	agentsPath := filepath.Join(workspaceDir, "AGENTS.md")
	var existingContent string
	if data, err := os.ReadFile(agentsPath); err == nil {
		existingContent = string(data)
	}

	updatedContent := appendOrUpdateVerificationCommands(existingContent, generated)

	if err := os.WriteFile(agentsPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write to AGENTS.md: %w", err)
	}

	return generated, nil
}
