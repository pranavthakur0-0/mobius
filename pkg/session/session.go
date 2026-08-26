package session

import (
	"mobius/pkg/agent"
	"mobius/pkg/agentctx"
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
func NewSession(id, name string, a *agent.Agent) *Session {
	systemPrompt := agentctx.BuildSystemPrompt("")
	convContext := agentctx.NewConversationContext(systemPrompt)
	now := time.Now()
	return &Session{
		ID:        id,
		Name:      "New Chat",
		Agent:     a,
		Context:   convContext,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
