package sensors

import (
	"bufio"
	"strings"
)

// ExtractCommands scans a guide (like AGENTS.md) for build and test commands.
func ExtractCommands(content string) (buildCmd, testCmd string) {
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
	}
	return buildCmd, testCmd
}



func NewRegistryFromGuide(workspaceDir string, guideContent string) *Registry {
	reg := NewRegistry()
	buildCmd, testCmd := ExtractCommands(guideContent)
	if buildCmd != "" {
		_ = reg.Register(NewCommandSensor("build", buildCmd, workspaceDir))
	}
	if testCmd != "" {
		_ = reg.Register(NewCommandSensor("test", testCmd, workspaceDir))
	}
	return reg
}
