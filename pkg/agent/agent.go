package agent

import (
	"fmt"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
	"mobius/pkg/tracer"
	"time"
)



type Agent struct {
    threadID string // Session / Thread ID
    provider llm.Provider
    registry *tools.Registry
    model    string
    maxSteps int
    maxCost  float64
    timeout  time.Duration
    toolDefs []llm.ToolDefinition
}








func NewAgent(provider llm.Provider, registry *tools.Registry, model string) (*Agent, error) {


	// 1. Validate required core components
	if provider == nil {
		return nil, fmt.Errorf("llm provider is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("tools registry is required")
	}
	if model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	agentCfg, err := LoadAgentConfig("pkg/agent/agent.toml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent config : %w",  err)
	}


	// 2. Strict validation for agent budget limits
	if agentCfg.MaxSteps <= 0 {
		return nil, fmt.Errorf("max_steps must be greater than 0, got %d", agentCfg.MaxSteps)
	}
	if agentCfg.MaxCost <= 0 {
		return nil, fmt.Errorf("max_cost must be greater than 0, got %.2f", agentCfg.MaxCost)
	}
	if agentCfg.TimeoutSeconds <= 0 {
		return nil, fmt.Errorf("timeout_seconds must be greater than 0, got %d", agentCfg.TimeoutSeconds)
	}


	
	// 3. Build tool definitions once
		var toolDefs []llm.ToolDefinition
		for _, t := range registry.List() {
			toolDefs = append(toolDefs, llm.ToolDefinition{
				Type: "function",
				Function: llm.FunctionDef{
					Name:        t.Name(),
					Description: t.Description(),
					Parameters:  t.Schema(),
				},
			})
		}
		agent := &Agent{
			threadID: tracer.NewThreadID(),
			provider: provider,
			registry: registry,
			model:    model,
			maxSteps: agentCfg.MaxSteps,
			maxCost:  agentCfg.MaxCost,
			timeout:  time.Duration(agentCfg.TimeoutSeconds) * time.Second,
			toolDefs: toolDefs, // Store once
		}
	return agent, nil
}



