package cli

import (
	"fmt"
	"os"
	"strings"

	"mobius/pkg/agent"
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
	defer term.Restore(fd, oldState)

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

		// Slash menu trigger on empty line
		if b == '/' && len(input) == 0 {
			term.Restore(fd, oldState)
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

func StartREPL(sm *session.Manager, registry *tools.Registry, cfg *llm.Config) {
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

			switch {
			case input == "/exit":
				fmt.Println("Goodbye!")
				return
			case input == "/listchats":
				selectedID := ShowSessionMenu(sm)
				if selectedID != "" {
					s, err := sm.SwitchSession(selectedID)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					} else {
						fmt.Printf("Switched to session: '%s'\n", s.Name)
					}
				}
			case input == "/newchat":
				handleNewChat(sm, registry, cfg)
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
func handleNewChat(sm *session.Manager, registry *tools.Registry, cfg *llm.Config) {
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

	newAgent, err := agent.NewAgent(provider, registry, cfg.ActiveModel)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	newSess := sm.CreateSession("New Chat", newAgent)
	fmt.Printf("Started new session '%s' (ID: %s)\n", newSess.Name, newSess.ID)
}
