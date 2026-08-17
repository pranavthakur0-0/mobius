package tracer

import (
	"sync"
)



var logMu sync.Mutex


// LogEntry records what happened in one step.
type LogEntry struct {
	ThreadID string `json:"thread_id"`
	Step     int    `json:"step"`
	Thought  string `json:"thought,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
	Output   string `json:"output,omitempty"`
}


