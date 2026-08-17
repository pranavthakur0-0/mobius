package cli

import (
	"bufio"
	"fmt"
	"mobius/pkg/agent"
	"mobius/pkg/llm"
	"mobius/pkg/session"
	"mobius/pkg/tools"
	"os"
	"strings"
)

func StartREPL(sm *session.Manager, registry *tools.Registry, cfg *llm.Config) {
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

		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Exiting Mobius.")
			break
		}

		if strings.ToLower(input) == "/listchats" {
			list := sm.ListSession()
			fmt.Println("List of sessions : ")
			for _, item := range list {
				activeMarker := " "
				if item.IsActive {
					activeMarker = "*"
				}
				fmt.Printf("  [%s] %s \n", activeMarker, item.Name)
				
			}
			continue
		}

		if strings.ToLower(input) == "/newchat" {
			if unstarted := sm.GetUnstarted(); unstarted != nil {
				_, _ = sm.SwitchSession(unstarted.ID)
				fmt.Printf("Reusing empty session '%s' (ID: %s).\n", unstarted.Name, unstarted.ID)
				continue
			}

			provider, err := cfg.GetProviderForModel("")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			newAgent, err := agent.NewAgent(provider, registry, cfg.ActiveModel)
			if err != nil {
				fmt.Printf("Error creating agent: %v\n", err)
				continue
			}

			newSess := sm.CreateSession("chat", newAgent)
			fmt.Printf("Started new session '%s' (ID: %s)\n", newSess.Name, newSess.ID)
			continue
		}

		activeSession, err := sm.GetActive()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		_ = RunGoal(activeSession, input)
	}
}
