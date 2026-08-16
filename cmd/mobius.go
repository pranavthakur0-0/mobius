package main

import (
	// "bufio"
	"context"
	"fmt"
	"os"

	// "os"

	// "strings"

	// "mobius/pkg/llm"
	"mobius/pkg/agent"
	"mobius/pkg/llm"
	"mobius/pkg/tools"

	"github.com/joho/godotenv"
)



func main() {
	_ = godotenv.Load()
	fmt.Println("-------------------------------------------")
	fmt.Println("       🌀 MOBIUS AGENT HARNESS 🌀          ")
	fmt.Println("-------------------------------------------")
	fmt.Println("Commands:")
	fmt.Println("  tools                       - List registered tools")
	fmt.Println("  run <command>               - Run a shell command")
	fmt.Println("  view <file>                 - View a file with line numbers")
	fmt.Println("  write <file> <content>      - Write content to a file")
	fmt.Println("  list [path]                 - List files in a directory")
	fmt.Println("  grep <pattern> [dir]        - Search text/regex across files")
	fmt.Println("  edit <file>                 - Interactive exact text replacement")
	fmt.Println("  exit / quit                 - Exit harness")

	// 1. Initialize Registry and Register All Standard Tools
	registry := tools.NewDefaultRegistry(".")



	cfg, err := llm.LoadConfig("config.toml")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}


	provider, err := cfg.GetProviderForModel("")
	if err != nil {
		fmt.Printf("Failed to get provider: %s\n", err)
		return 
	}



	agent, err := agent.NewAgent(provider, registry, cfg.ActiveModel)
	if err != nil {
		fmt.Printf("Failed to get provider: %s\n", err)
		return 
	}

	
	goal := "Inspect config.toml using view_file and summarize the configured providers."
	result, err := agent.Run(context.Background(), goal)
	if err != nil {
		fmt.Printf("❌ Agent execution failed: %v\n", err)
		return
	}
	fmt.Println("===========================================")
	fmt.Println("🎉 AGENT COMPLETED GOAL:")
	fmt.Println(result)
	fmt.Println("===========================================")

}
