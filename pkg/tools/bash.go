package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// bashArgs defines the JSON arguments for the run_command tool.
type bashArgs struct {
	Command string `json:"command"`
}

// BashTool executes shell commands with context-aware process isolation.
type BashTool struct {
	WorkingDir string
}

// NewBashTool creates a new BashTool bound to the specified working directory.
func NewBashTool(workDir string) *BashTool {
	return &BashTool{WorkingDir: workDir}
}

// Name returns the tool identifier "run_command".
func (b *BashTool) Name() string {
	return "run_command"
}

// Description returns a helpful summary of the tool for the LLM.
func (b *BashTool) Description() string {
	return "Executes a shell/bash command and returns the combined stdout and stderr."
}

// Schema returns the JSON schema parameter specification for run_command.
func (b *BashTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]Property{
			"command": {
				Type:        "string",
				Description: "The shell command line to execute.",
			},
		},
		Required: []string{"command"},
	}
}

// Execute runs the command in a subprocess, respecting context cancellation and capping output length.
func (b *BashTool) Execute(ctx context.Context, jsonArgs string) (string, error) {
	var args bashArgs
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return "", fmt.Errorf("invalid json arguments for run_command: %w", err)
	}

	command := strings.TrimSpace(args.Command)
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	// Uses CommandContext so if ctx cancels or times out, the subprocess is killed immediately
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	if b.WorkingDir != "" {
		cmd.Dir = b.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	outStr := string(output)

	// Context-safe output capping (50KB max)
	maxLen := 50000
	if len(outStr) > maxLen {
		outStr = outStr[:maxLen] + fmt.Sprintf("\n... [Output truncated: capped at %d bytes]", maxLen)
	}

	if err != nil {
		return outStr, fmt.Errorf("command execution failed: %w", err)
	}

	return outStr, nil
}
