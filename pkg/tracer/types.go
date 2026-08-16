package tracer

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
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




// NewThreadID generates a unique session ID.
func NewThreadID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("th_%d_%x", time.Now().UnixNano(), b)
}

// AppendLog writes one JSON line to .mobius/traces/history.jsonl.
func AppendLog(entry LogEntry) error {

	logMu.Lock()
	defer logMu.Unlock()

	_ = os.MkdirAll(".mobius/traces", 0755)

	f, err := os.OpenFile(fmt.Sprintf(".mobius/traces/%s.jsonl", entry.ThreadID), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(entry)
}
