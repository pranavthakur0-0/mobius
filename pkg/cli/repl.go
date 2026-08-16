package cli

import (
	"bufio"
	"fmt"
	"mobius/pkg/agent"
	"mobius/pkg/llm"
	"mobius/pkg/tools"
	"os"
	"strings"
)

func StartREPL(a *agent.Agent, registry *tools.Registry, cfg *llm.Config) {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Err() != nil {
		return 
	}

	for {
		fmt.Print("mobius > ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == ""{
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Exiting Mobius.")
			break
		}

		_ = RunGoal(a, input)
	}
}