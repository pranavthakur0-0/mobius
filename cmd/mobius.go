package main

import (
	// "bufio"
	"context"
	"fmt"
	"os"

	// "strings"

	"mobius/pkg/llm"
	"github.com/joho/godotenv"
	// "mobius/pkg/tools"
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
	fmt.Println("  exit / quit                 - Exit harness\n")

	// 1. Initialize Registry and Register All Standard Tools
	// registry := tools.NewRegistry()
	// registry.Register(tools.NewBashTool("."))
	// registry.Register(tools.NewViewFileTool("."))
	// registry.Register(tools.NewWriteTool("."))
	// registry.Register(tools.NewListTool("."))
	// registry.Register(tools.NewEditFileTool("."))
	// registry.Register(tools.NewGrepTool("."))

	// 2. Load the provider

	cfg, err := llm.LoadConfig("config.toml")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	allModels, err := cfg.ListAllModels()
	if err != nil {
		fmt.Printf("Failed to load the active %s", err)
	}
	fmt.Println(allModels)

	provider, err := cfg.GetProviderForModel("")
	if err != nil {
		fmt.Printf("Failed to get provider: %s\n", err)
		return 
	}

	req := &llm.ChatRequest{

    Model: cfg.ActiveModel,
    Message: []llm.Message{
        {Role: llm.RoleUser, Content: "Hello World"},
    },
}

	response, err := provider.Generate(context.Background(), req)
	if err != nil {
		fmt.Printf("Failed to get to llm : %s\n", err)
		return 
	}

	fmt.Println(response)
	

	// provider := llm

	// activeModel := cfg.ActiveModel

	// scanner := bufio.NewScanner(os.Stdin)

	// for {
	// 	fmt.Print("mobius > ")
	// 	if !scanner.Scan() {
	// 		break
	// 	}
	// 	input := strings.TrimSpace(scanner.Text())
	// 	if input == "" {
	// 		continue
	// 	}

	// 	if strings.ToLower(input) == "exit" || strings.ToLower(input) == "quit" {
	// 		fmt.Println("Goodbye!!")
	// 		break
	// 	}

	// 	if strings.ToLower(input) == "tools" {
	// 		fmt.Println("\n🛠️  Available Tools in Harness:")
	// 		for _, tool := range registry.List() {
	// 			fmt.Printf("  %-15s : %s\n", tool.Name(), tool.Description())
	// 		}
	// 		fmt.Println()
	// 		continue
	// 	}

	// 	parts := strings.Fields(input)
	// 	firstWord := strings.ToLower(parts[0])

	// 	switch firstWord {
	// 	case "run", "!run":
	// 		if len(parts) < 2 {
	// 			fmt.Println("❌ Error: Please specify a command (e.g. run ls -la)")
	// 			continue
	// 		}
	// 		cmdToRun := strings.Join(parts[1:], " ")
	// 		tool, err := registry.Get("run_command")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(cmdToRun)
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	case "read", "!read", "view", "!view", "cat":
	// 		if len(parts) < 2 {
	// 			fmt.Println("❌ Error: Please specify a file path (e.g. view main.go)")
	// 			continue
	// 		}
	// 		filePath := parts[1]
	// 		tool, err := registry.Get("view_file")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(filePath)
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	case "write", "!write":
	// 		if len(parts) < 2 {
	// 			fmt.Println("❌ Error: Usage: write <filepath> [content]")
	// 			continue
	// 		}
	// 		filePath := parts[1]
	// 		content := ""
	// 		if len(parts) > 2 {
	// 			content = strings.Join(parts[2:], " ")
	// 		}
	// 		tool, err := registry.Get("write_file")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(fmt.Sprintf("%s\n%s", filePath, content))
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	case "list", "!list", "ls", "dir":
	// 		targetDir := "."
	// 		if len(parts) >= 2 {
	// 			targetDir = parts[1]
	// 		}
	// 		tool, err := registry.Get("list_dir")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(targetDir)
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	case "grep", "!grep", "search":
	// 		if len(parts) < 2 {
	// 			fmt.Println("❌ Error: Usage: grep <pattern> [directory]")
	// 			continue
	// 		}
	// 		args := strings.Join(parts[1:], " ")
	// 		tool, err := registry.Get("grep_search")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(args)
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	case "edit", "!edit":
	// 		if len(parts) < 2 {
	// 			fmt.Println("❌ Error: Usage: edit <filepath>")
	// 			continue
	// 		}
	// 		filePath := parts[1]

	// 		fmt.Print("Enter exact text to replace (old text): ")
	// 		if !scanner.Scan() {
	// 			continue
	// 		}
	// 		oldText := scanner.Text()

	// 		fmt.Print("Enter replacement text (new text): ")
	// 		if !scanner.Scan() {
	// 			continue
	// 		}
	// 		newText := scanner.Text()

	// 		editArg := fmt.Sprintf("%s\n<<<OLD>>>\n%s\n<<<NEW>>>\n%s", filePath, oldText, newText)

	// 		tool, err := registry.Get("edit_file")
	// 		if err != nil {
	// 			fmt.Printf("❌ %v\n", err)
	// 			continue
	// 		}
	// 		output, err := tool.Execute(editArg)
	// 		if err != nil {
	// 			fmt.Printf("❌ Error: %v\n", err)
	// 		}
	// 		fmt.Println(output)

	// 	default:
	// 		fmt.Printf("❓ Unknown command: %s. Type 'tools' to see available tools.\n", firstWord)
	// 	}
	// }
}
