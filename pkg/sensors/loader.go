package sensors

import (
	"bufio"
	"strings"
)

// ExtractCommands scans a guide (like AGENTS.md) for build and test commands.
func ExtractCommands(content string) (buildCmd, testCmd string) {
	b, t, _ := ExtractAllCommands(content)
	return b, t
}

// ExtractAllCommands scans a guide for build, test, and lint commands.
func ExtractAllCommands(content string) (buildCmd, testCmd, lintCmd string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lower := strings.ToLower(line)

		// Look for lines like "Build: go build ./..." or "- Build: make build"
		if strings.HasPrefix(lower, "build:") || strings.HasPrefix(lower, "- build:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				buildCmd = strings.TrimSpace(parts[1])
			}
		}

		// Look for lines like "Test: go test ./..." or "- Test: pytest"
		if strings.HasPrefix(lower, "test:") || strings.HasPrefix(lower, "- test:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				testCmd = strings.TrimSpace(parts[1])
			}
		}

		// Look for lines like "Lint: golangci-lint run" or "- Lint: flake8"
		if strings.HasPrefix(lower, "lint:") || strings.HasPrefix(lower, "- lint:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				lintCmd = strings.TrimSpace(parts[1])
			}
		}
	}
	return buildCmd, testCmd, lintCmd
}

func NewRegistryFromGuide(workspaceDir string, guideContent string) *Registry {
	reg := NewRegistry()
	buildCmd, testCmd, lintCmd := ExtractAllCommands(guideContent)
	if buildCmd != "" {
		_ = reg.Register(NewCommandSensor("build", buildCmd, workspaceDir))
	}
	if testCmd != "" {
		_ = reg.Register(NewCommandSensor("test", testCmd, workspaceDir))
	}
	if lintCmd != "" {
		_ = reg.Register(NewCommandSensor("lint", lintCmd, workspaceDir))
	}
	return reg
}
