package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// NewID generates a timestamped, collision-resistant ID with a given prefix.
// Example: NewID("evt") -> "evt_1719283921_a4b9c1"
func NewID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%d_%x", prefix, time.Now().UnixNano(), b)
}

// NewThreadID generates a session/thread ID: "th_..."
func NewThreadID() string {
	return NewID("th")
}

// NewEventID generates an event ID: "evt_..."
func NewEventID() string {
	return NewID("evt")
}

// NewArtifactID generates an artifact ID: "art_..."
func NewArtifactID() string {
	return NewID("art")
}
