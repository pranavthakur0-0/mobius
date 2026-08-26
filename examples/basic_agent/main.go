package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"mobius/pkg/agent"
	"mobius/pkg/agentctx"
	"mobius/pkg/events"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
)

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

	// 1. Initialize LLM Provider
	provider := llm.NewOpenAIProvider(apiKey, baseURL)

	// 2. Initialize Default Tool Registry
	registry := tools.NewDefaultRegistry(".")

	// 3. Initialize Event Store
	eventStore, err := events.NewFileEventStore(".mobius/events", 256)
	if err != nil {
		log.Fatalf("Failed to initialize event store: %v", err)
	}
	defer eventStore.Close()

	// 4. Create Agent
	ag, err := agent.NewAgent(provider, registry, model, 0.14, 0.28, eventStore)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 5. Initialize Conversation Context
	conv := agentctx.NewConversationContext(agentctx.BuildSystemPrompt(""))

	// 6. Run a Goal
	goal := "List the current directory contents and describe the files you see."
	fmt.Printf("Running goal: %s\n\n", goal)

	response, err := ag.Run(ctx, conv, goal)
	if err != nil {
		log.Fatalf("Agent run failed: %v", err)
	}

	fmt.Printf("\nFinal Response:\n%s\n", response)
}
