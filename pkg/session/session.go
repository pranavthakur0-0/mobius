package agent

import (
	"context"
	"time"
)

// Session represents a single conversational agent session with its own independent memory.
type Session struct {
	ID        string
	Title     string
	Context   *context.Context
	CreatedAt time.Time
	UpdatedAt time.Time
}


// NewSession creates a new session with its own conversation context.
func NewSession(id, title string, ctx *context.Context) *Session {
	now := time.Now()
	return &Session{
		ID:        id,
		Title:     title,
		Context:   ctx,
		CreatedAt: now,
		UpdatedAt: now,
	}
}