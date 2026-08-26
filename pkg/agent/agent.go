package agent

import (
	"context"
	"fmt"
	"mobius/pkg/agentctx"
	"mobius/pkg/artifact"
	"mobius/pkg/budget"
	"mobius/pkg/events"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
	"mobius/pkg/utils"
	"strings"
	"time"
)



type Agent struct {
	threadID      string
	provider      llm.Provider
	registry      *tools.Registry
	model         string
	maxSteps      int
	maxCost       float64
	timeout       time.Duration
	toolDefs      []llm.ToolDefinition
	tracker       budget.CostTracker
	events        events.EventStore // EventStore
	artifactStore *artifact.Store
	compactor     *agentctx.Compactor
}





func NewAgent(provider llm.Provider, registry *tools.Registry, model string, pricePrompt float64, priceComp float64,  eventStore events.EventStore) (*Agent, error) {


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

	artStore, err := artifact.NewStore(".mobius/artifacts")
		if err != nil {
			return nil, fmt.Errorf("failed to init artifact store: %w", err)
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
		 // Pass the real prices instead of 0.15 and 0.60
        budgetTracker := budget.NewTracker(agentCfg.MaxCost, pricePrompt, priceComp)

		agent := &Agent{
			threadID:      utils.NewThreadID(),
			provider:      provider,
			registry:      registry,
			model:         model,
			maxSteps:      agentCfg.MaxSteps,
			maxCost:       agentCfg.MaxCost,
			timeout:       time.Duration(agentCfg.TimeoutSeconds) * time.Second,
			toolDefs:      toolDefs, // Store once
			tracker:       budgetTracker,
			events:        eventStore,
			artifactStore: artStore,
			compactor:     agentctx.NewCompactor(agentctx.DefaultCompactorConfig(), provider, model),
		}
	return agent, nil
}


func (a *Agent) Events() events.EventStore {
	return a.events
}


func (a *Agent) ThreadID() string {
	return a.threadID
}



// GenerateTitle asks the LLM for a concise 3-4 word title for the session
func (a *Agent) GenerateTitle(ctx context.Context, prompt string) string {
	req := &llm.ChatRequest{
		Model: a.model,
		Message: []llm.Message{
			{
				Role:    llm.RoleUser,
				Content: agentctx.TitlePrompt(prompt),
			},
		},
	}

	resp, err := a.provider.Generate(ctx, req)
	if err != nil || strings.TrimSpace(resp.Content) == "" {
		// Fallback: take first 4 words of the prompt
		words := strings.Fields(prompt)
		if len(words) > 4 {
			words = words[:4]
		}
		return strings.Join(words, " ")
	}

	return strings.TrimSpace(resp.Content)
}



func (a *Agent) GetModel() string {
	return a.model
}


func (a *Agent) SetModel(model string, provider llm.Provider, pricePrompt, priceComp float64) {
    a.model = model
    a.provider = provider
    a.tracker.UpdatePrices(pricePrompt, priceComp)
    if a.compactor != nil {
        a.compactor.UpdateModel(model, provider)
    }
}
