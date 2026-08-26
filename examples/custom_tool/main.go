package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"mobius/pkg/agent"
	"mobius/pkg/agentctx"
	"mobius/pkg/events"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
)

// CalculatorTool implements the tools.Tool interface for custom math operations.
type CalculatorTool struct{}

type CalcParams struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
}

func (c *CalculatorTool) Name() string {
	return "calculator"
}

func (c *CalculatorTool) Description() string {
	return "Performs arithmetic operations: add, subtract, multiply, divide."
}

func (c *CalculatorTool) Schema() tools.ToolSchema {
	return tools.ToolSchema{
		Type: "object",
		Properties: map[string]tools.Property{
			"operation": {
				Type:        "string",
				Description: "The math operation to perform (add, subtract, multiply, divide)",
			},
			"a": {
				Type:        "number",
				Description: "First number",
			},
			"b": {
				Type:        "number",
				Description: "Second number",
			},
		},
		Required: []string{"operation", "a", "b"},
	}
}

func (c *CalculatorTool) Execute(ctx context.Context, input string) (string, error) {
	var params CalcParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid json arguments: %w", err)
	}

	var res float64
	switch params.Operation {
	case "add":
		res = params.A + params.B
	case "subtract":
		res = params.A - params.B
	case "multiply":
		res = params.A * params.B
	case "divide":
		if params.B == 0 {
			return "", fmt.Errorf("division by zero")
		}
		res = params.A / params.B
	default:
		return "", fmt.Errorf("unsupported operation: %s", params.Operation)
	}

	return fmt.Sprintf("Result: %g", res), nil
}

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseURL := "https://api.deepseek.com/v1"
	model := "deepseek-chat"

	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		baseURL = "https://api.openai.com/v1"
		model = "gpt-4o-mini"
	}
	if apiKey == "" {
		log.Fatal("Please set DEEPSEEK_API_KEY or OPENAI_API_KEY in your environment.")
	}

	ctx := context.Background()

	provider := llm.NewOpenAIProvider(apiKey, baseURL)

	// Register our custom tool alongside standard tools
	registry := tools.NewDefaultRegistry(".")
	registry.Register(&CalculatorTool{})

	eventStore, err := events.NewFileEventStore(".mobius/events", 256)
	if err != nil {
		log.Fatalf("Failed to initialize event store: %v", err)
	}
	defer func() { _ = eventStore.Close() }()

	ag, err := agent.NewAgent(provider, registry, model, 0.14, 0.28, eventStore)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	conv := agentctx.NewConversationContext(agentctx.BuildSystemPrompt(""))

	goal := "What is (1248 * 42) / 3? Use the calculator tool."
	fmt.Printf("Running custom tool goal: %s\n\n", goal)

	response, err := ag.Run(ctx, conv, goal)
	if err != nil {
		log.Fatalf("Agent run failed: %v", err)
	}

	fmt.Printf("\nFinal Response:\n%s\n", response)
}
