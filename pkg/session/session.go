package session

import (
	"mobius/pkg/agent"
	"mobius/pkg/agentctx"
	"mobius/pkg/guides"
	"time"
)

// Session represents an independent agent with its own persistent memory history.
type Session struct {
	ID        string
	Name      string
	Agent     *agent.Agent
	Context   *agentctx.ConversationContext
	CreatedAt time.Time
	UpdatedAt time.Time
	Started   bool
}

// NewSession initializes an agent session with its own memory context.
func NewSession(id, name string, a *agent.Agent, systemPrompt ...string) *Session {
	prompt := ""
	if len(systemPrompt) > 0 {
		prompt = systemPrompt[0]
	}
	if prompt == "" {
		guidesPrompt := ""
		if gs, err := guides.LoadFromWorkSpace("config/guides.toml", "."); err == nil {
			guidesPrompt = gs.RenderSystemPrompt()
		}
		prompt = agentctx.BuildSystemPrompt("", guidesPrompt)
	}

	convContext := agentctx.NewConversationContext(prompt)
	now := time.Now()
	return &Session{
		ID:        id,
		Name:      name,
		Agent:     a,
		Context:   convContext,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
