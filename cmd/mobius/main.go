package main

import (
	"context"
	"fmt"
	"mobius/pkg/agent"
	"mobius/pkg/agentctx"
	"mobius/pkg/cli"
	"mobius/pkg/events"
	"mobius/pkg/guides"
	"mobius/pkg/llm"
	"mobius/pkg/sensors"
	"mobius/pkg/session"
	"mobius/pkg/tools"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {

	guidesPrompt := ""
	if gs, err := guides.LoadFromWorkSpace("config/guides.toml", "."); err == nil {
		guidesPrompt = gs.RenderSystemPrompt()
	}

	systemPrompt := agentctx.BuildSystemPrompt("", guidesPrompt)

	_ = godotenv.Load()
	registry := tools.NewDefaultRegistry(".")

	// 1. Create the Shared Event Store
	eventStore, err := events.NewFileEventStore(".mobius/events", 256)
	if err != nil {
		fmt.Printf("Error: Failed to init event store: %v\n", err)
		os.Exit(1)
	}
	defer eventStore.Close() // Gracefully flushes all events when app exits!

	cfg, err := llm.LoadConfig("config/model_config.toml")
	if err != nil {
		// Fallback to root model_config.toml if present
		cfg, err = llm.LoadConfig("model_config.toml")
		if err != nil {
			fmt.Printf("Error: Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	provider, err := cfg.GetProviderForModel("")
	if err != nil {
		fmt.Printf("Error: Failed to get provider: %v\n", err)
		os.Exit(1)
	}

	// 2. Create Default Agent with the shared eventStore
	pCost, cCost := cfg.GetPrices(cfg.ActiveModel)
	defaultAgent, err := agent.NewAgent(provider, registry, cfg.ActiveModel, pCost, cCost, eventStore)
	if err != nil {
		fmt.Printf("Error: Failed to create agent: %v\n", err)
		os.Exit(1)
	}

	// 2b. Initialize sensors from AGENTS.md if present
	if agentsData, err := os.ReadFile("AGENTS.md"); err == nil {
		sensorReg := sensors.NewRegistryFromGuide(".", string(agentsData))
		defaultAgent.SetSensors(sensorReg)
	}

	defaultSess := session.NewSession(defaultAgent.ThreadID(), "default", defaultAgent, systemPrompt)
	manager := session.NewManager(defaultSess)

	// 3. One-shot CLI run if arguments passed
	if len(os.Args) > 1 {
		if os.Args[1] == "init" {
			targetDir := "."
			if len(os.Args) > 2 {
				targetDir = os.Args[2]
			}
			fmt.Printf("Scanning workspace (%s) and generating verification commands with %s...\n", targetDir, cfg.ActiveModel)
			gen, err := guides.InitWorkspace(context.Background(), provider, cfg.ActiveModel, targetDir)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\nAGENTS.md updated successfully:\n\n%s\n", gen)
			return
		}

		userInstruction := strings.Join(os.Args[1:], " ")
		_ = cli.RunGoal(defaultSess, userInstruction)
		return
	}

	// 4. Start Interactive REPL
	cli.StartREPL(manager, registry, cfg, defaultAgent, eventStore)
}
