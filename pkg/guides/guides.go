package guides

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Guide struct {
	Source   string
	Path     string
	Content  string
	Priority int
}

type GuideSet struct {
	Guides []Guide
}

func NewGuideSet() *GuideSet {
	return &GuideSet{
		Guides: make([]Guide, 0),
	}
}

func (gs *GuideSet) Add(g Guide) {
	gs.Guides = append(gs.Guides, g)
}

func (gs *GuideSet) IsEmpty() bool {
	return len(gs.Guides) == 0
}

func (gs *GuideSet) SortByPriority() {
	sort.SliceStable(gs.Guides, func(i, j int) bool {
		return gs.Guides[i].Priority > gs.Guides[j].Priority
	})
}

type Candidate struct {
	Path     string `toml:"path"`
	Priority int
}

type GuidesConfig struct {
	Candidates []Candidate `toml:"candidates"`
}

func LoadFromWorkSpace(configPath string, workspaceRoot string) (*GuideSet, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read guides config '%s': %w", configPath, err)
	}

	var cfg GuidesConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parser guides config '%s' : %w", configPath, err)
	}

	gs := NewGuideSet()
	for _, c := range cfg.Candidates {
		targetPath := filepath.Join(workspaceRoot, c.Path)

	 	data, err := os.ReadFile(targetPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("failed to read the md '%s' : %w", targetPath, err)
		}

		var guide = Guide{
			Source: c.Path,
			Path: targetPath,
			Content: string(data),
			Priority: c.Priority,
		}
		gs.Add(guide)
	}
	gs.SortByPriority()

	return gs, nil
}



func (gs *GuideSet) RenderSystemPrompt() string {
	if gs.IsEmpty() {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\n# WORKSPACE GUIDES & RULES\n")
	sb.WriteString("The following project-specific instructions and constraints take precedence:\n\n")
	for _, g := range gs.Guides {
		sb.WriteString("## Guide: ")
		sb.WriteString(g.Source)
		sb.WriteString("\n```markdown\n")
		sb.WriteString(strings.TrimSpace(g.Content))
		sb.WriteString("\n```\n\n")
	}
	return sb.String()
}