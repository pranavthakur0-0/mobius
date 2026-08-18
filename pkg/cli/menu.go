package cli

import (
	"fmt"
	"mobius/pkg/llm"
	"mobius/pkg/session"
	"sort"

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
					huh.NewOption("Models", "/models"),
					huh.NewOption("Switch / List Chats", "/listchats"),
					huh.NewOption("Exit Mobius", "/exit"),
				).
				Value(&selected),
		),
	).WithKeyMap(newMenuKeyMap())
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
	).WithKeyMap(newMenuKeyMap())
	err := form.Run()
	if err != nil {
		return "" // User cancelled with Esc or Ctrl+C
	}
	return selectedID
}


func newMenuKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit.SetKeys("esc", "ctrl+c")
	return km
}

func ShowModelMenu(cfg *llm.Config, currentModel string) (string, error) {
	allModels, err := cfg.ListAllModels()
	if err != nil {
		return "", fmt.Errorf("error listing models: %w", err)
	}
	keys := make([]string, 0, len(allModels))
	for k := range allModels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var options []huh.Option[string]
	for _, modelName := range keys {
		info := allModels[modelName]
		prefix := "  "
		suffix := ""
		if info.Model == currentModel {
			prefix = "\033[32m*\033[0m "
			suffix = " \033[32m(---active---)\033[0m"
		}
		label := fmt.Sprintf("%s%-30s (%s)%s", prefix, info.Model, info.ProviderName, suffix)
		options = append(options, huh.NewOption(label, info.Model))
	}

	var selectedModel string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a Model").
				Description("Use ↑/↓ to navigate, Enter to switch, Esc to cancel").
				Options(options...).
				Value(&selectedModel),
		),
	).WithKeyMap(newMenuKeyMap()) // <-- Enable Esc to cancel
	err = form.Run()
	if err != nil {
		return "", nil // Esc or Ctrl+C returns empty string
	}
	return selectedModel, nil
}