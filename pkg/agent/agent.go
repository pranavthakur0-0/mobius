package agent

import (
	"context"
	"fmt"
	contextPkg "mobius/pkg/context"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
	"os"
	"time"
	"github.com/pelletier/go-toml/v2"
)


type AgentConfig struct {
	MaxSteps       int     `toml:"max_steps"`
	MaxCost        float64 `toml:"max_cost"`
	TimeoutSeconds int     `toml:"timeout_seconds"`
}

type Agent struct {
	provider llm.Provider
	registry *tools.Registry
	model    string
	maxSteps int
	maxCost  float64
	timeout  time.Duration
	toolDefs []llm.ToolDefinition 
}


func loadAgentConfig(path string) (*AgentConfig, error) {
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

	agentCfg, err := loadAgentConfig("pkg/agent/agent.toml")
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





func (a *Agent) Run(ctx context.Context, userInstruction string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	systemPrompt := contextPkg.BuildSystemPrompt("")
	c := contextPkg.NewConverationContext(systemPrompt)

	c.AddUserMessage(userInstruction)

	fmt.Printf("🎯 Goal: %s\n\n", userInstruction)


	for step := 1; step <= a.maxSteps; step++ {
		if err := ctx.Err(); err != nil {
			return "", fmt.Errorf("agent interrupted: %w", err)
		}

		fmt.Printf("🔄 [Step %d/%d] Thinking...\n", step, a.maxSteps)

	

		req := &llm.ChatRequest{
			Model: a.model,
			Message: c.Messages(),
			Tools: a.toolDefs,
		}

		resp, err := a.provider.Generate(ctx, req)
		if err != nil {
			return "", fmt.Errorf("step %d LLM call failed: %w", step, err)
		}

		c.AddAssistantMessage(resp.Content, resp.ToolCalls)
		if resp.Content != "" {
			fmt.Printf("🤖 Thought:\n%s\n\n", resp.Content)
		}

		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		for _, tc := range resp.ToolCalls {
			fmt.Printf("🛠️ Executing Tool: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
			tool, err := a.registry.Get(tc.Function.Name)
			var output string
			if err != nil {
				output = fmt.Sprintf("Error: tool '%s' not found", tc.Function.Name)
			} else {
				out, execErr := tool.Execute(ctx, tc.Function.Arguments)
				if execErr != nil {
					output = fmt.Sprintf("Tool error: %s\nOutput: %s", execErr, out)
				} else {
					output = out
				}
			}
			// Add tool observation to history
			c.AddToolResult(tc.ID, output)
		}
		
	}
	
	return "", fmt.Errorf("agent reached maximum step budget (%d steps)", a.maxSteps)

}