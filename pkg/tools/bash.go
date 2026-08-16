package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

type BashTool struct {
	WorkingDir string
}

func NewBashTool(workDir string) *BashTool {
	return &BashTool{
		WorkingDir: workDir,
	}
}

// Name satisfies the Tool interface
func (b *BashTool) Name() string {
	return "run_command"
}

func (b *BashTool) Description() string {
	return "Executes a shell/bash command and returns the combined stdout and stderr."
}

func (b *BashTool) Execute(command string) (string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	cmd := exec.Command("bash", "-c", command)
	if b.WorkingDir != "" {
		cmd.Dir = b.WorkingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w (output: %s)", err, string(output))
	}
	return string(output), nil
}
