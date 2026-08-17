package cli

import (
	"fmt"
	"mobius/pkg/session"

	"github.com/charmbracelet/huh"
)

func ShowActionMenu() string {
	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Mobius Actions").
				Description("Use ↑/↓ arrows to navigate, Enter to select").
				Options(
					huh.NewOption("Start New Chat", "/newchat"),
					huh.NewOption("Switch / List Chats", "/listchats"),
					huh.NewOption("Exit Mobius", "/exit"),
				).
				Value(&selected),
		),
	)
	err := form.Run()
	if err != nil {
		return ""
	}
	return selected
}



// ShowSessionMenu opens an interactive select menu listing all sessions
func ShowSessionMenu(sm *session.Manager) string {
	list := sm.ListSession()
	if len(list) == 0 {
		return ""
	}
	var options []huh.Option[string]
	for _, item := range list {
		label := item.Name
		if item.IsActive {
			label = fmt.Sprintf("* %s (active)", item.Name)
		} else {
			label = fmt.Sprintf("  %s (ID: %s)", item.Name, item.ID)
		}
		// label is shown to user, item.ID is returned as value
		options = append(options, huh.NewOption(label, item.ID))
	}
	var selectedID string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Session").
				Description("Use ↑/↓ to navigate, Enter to switch, Esc to cancel").
				Options(options...).
				Value(&selectedID),
		),
	)
	err := form.Run()
	if err != nil {
		return "" // User cancelled with Esc or Ctrl+C
	}
	return selectedID
}
