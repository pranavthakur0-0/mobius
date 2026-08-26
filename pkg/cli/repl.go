package cli

import (
	"fmt"
	"os"
	"strings"

	"mobius/pkg/agent"
	"mobius/pkg/events"
	"mobius/pkg/llm"
	"mobius/pkg/session"
	"mobius/pkg/tools"

	"github.com/charmbracelet/x/term"
)

func readPrompt(prompt string) (string, bool) {
	fmt.Print(prompt)
	fd := os.Stdin.Fd()
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", false
	}
	defer func() { _ = term.Restore(fd, oldState) }()

	var input []rune
	buf := make([]byte, 1)

	for {
		_, err = os.Stdin.Read(buf)
		if err != nil {
			return "", false
		}

		b := buf[0]

		// Enter key
		if b == '\r' || b == '\n' {
			fmt.Print("\r\n")
			return string(input), true
		}

		// Ctrl+C (ASCII 3)
		if b == 3 {
			fmt.Print("\r\n")
			return "", false
		}

		// Escape key (ASCII 27)
		if b == 27 {
			// If it's a standalone Esc, clear current input
			for len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
			continue
		}

		// Slash menu trigger on empty line
		if b == '/' && len(input) == 0 {
			_ = term.Restore(fd, oldState)
			fmt.Print("\r\n")

			selected := ShowActionMenu()
			if selected != "" {
				return selected, true
			}

			oldState, _ = term.MakeRaw(fd)
			fmt.Print(prompt)
			continue
		}

		// Backspace (127 / 8)
		if b == 127 || b == 8 {
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}
			continue
		}

		// Printable characters
		if b >= 32 && b <= 126 {
			r := rune(b)
			input = append(input, r)
			fmt.Print(string(r))
		}
	}
}

func StartREPL(sm *session.Manager, registry *tools.Registry, cfg *llm.Config, agent *agent.Agent, eventStore events.EventStore) {
	PrintBanner(cfg.ActiveModel)

	for {
		input, ok := readPrompt("mobius > ")
		if !ok {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		switch input {
		case "/exit":
			fmt.Println("Goodbye!")
			return
		case "/listchats":
			selectedID := ShowSessionMenu(sm)
			if selectedID != "" {
				s, err := sm.SwitchSession(selectedID)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				} else {
					fmt.Printf("Switched to session: '%s'\n", s.Name)
				}
			}
		case "/newchat":
			handleNewChat(sm, registry, cfg, eventStore)
		case "/models":
			activeSession, err := sm.GetActive()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			currentModel := activeSession.Agent.GetModel()
			model, err := ShowModelMenu(cfg, currentModel)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if model == "" || model == currentModel {
				continue // User cancelled with Esc or re-selected current model
			}
			provider, err := cfg.GetProviderForModel(model)
			if err != nil {
				fmt.Printf("Error getting provider for %s: %v\n", model, err)
				continue
			}
			pCost, cCost := cfg.GetPrices(model)
			activeSession.Agent.SetModel(model, provider, pCost, cCost)
			fmt.Printf("Switched model to '%s'\n", model)
		default:
			// Normal AI Prompt
			activeSession, err := sm.GetActive()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			_ = RunGoal(activeSession, input)
		}
	}
}

// Helper to keep StartREPL tidy
func handleNewChat(sm *session.Manager, registry *tools.Registry, cfg *llm.Config, eventStore events.EventStore) {
	if unstarted := sm.GetUnstarted(); unstarted != nil {
		_, _ = sm.SwitchSession(unstarted.ID)
		fmt.Printf("Reusing empty session '%s' (ID: %s).\n", unstarted.Name, unstarted.ID)
		return
	}

	provider, err := cfg.GetProviderForModel("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	pCost, cCost := cfg.GetPrices(cfg.ActiveModel)
	newAgent, err := agent.NewAgent(provider, registry, cfg.ActiveModel, pCost, cCost, eventStore)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	newSess := sm.CreateSession("New Chat", newAgent)
	fmt.Printf("Started new session '%s' (ID: %s)\n", newSess.Name, newSess.ID)
}
