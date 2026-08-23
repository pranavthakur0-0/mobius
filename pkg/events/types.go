package events

import (
	"context"
	"mobius/pkg/llm"
	"time"
)

type EventType string

const (
	EventUserMessage      EventType = "user_message"
	EventAssistantMessage EventType = "assistant_message"
	EventToolCall         EventType = "tool_call"
	EventToolResult       EventType = "tool_result"
	EventSummary          EventType = "summary"
)

// Event represents an immutable fact in the agent's timeline.
type Event struct {
	ID        string    `json:"id"`
	ThreadID  string    `json:"thread_id"`
	Seq       int       `json:"seq"`
	Step      int       `json:"step"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`

	// Content payloads
	Content   string         `json:"content,omitempty"`
	ToolCalls []llm.ToolCall `json:"tool_calls,omitempty"`

	// Tool execution details
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	ToolArgs   string `json:"tool_args,omitempty"`
	ToolOutput string `json:"tool_output,omitempty"`
	ToolError  string `json:"tool_error,omitempty"`

	// Artifact & Token optimization fields
	ContentRef string     `json:"content_ref,omitempty"`
	Preview    string     `json:"preview,omitempty"`
	Tokens     int        `json:"tokens,omitempty"`
	Usage      *llm.Usage `json:"usage,omitempty"`
}

// EventStore defines the storage and retrieval contract for session events.
type EventStore interface {
	Append(ctx context.Context, event Event) error
	GetEvents(ctx context.Context, threadID string) ([]Event, error)
	Close() error
}
