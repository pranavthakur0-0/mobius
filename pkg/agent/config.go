package agent

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type AgentConfig struct {
	MaxSteps       int     `toml:"max_steps"`
	MaxCost        float64 `toml:"max_cost"`
	TimeoutSeconds int     `toml:"timeout_seconds"`
}



func LoadAgentConfig(path string) (*AgentConfig, error) {
	if path == "" {
		path = "pkg/agent/agent.toml"
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agent config file '%s': %w", path, err)
	}
	var fileData struct {
		Agent AgentConfig `toml:"agent"`
	}
	if err := toml.Unmarshal(content, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse agent config '%s': %w", path, err)
	}
	return &fileData.Agent, nil
}

