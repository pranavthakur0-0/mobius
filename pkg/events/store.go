package events

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"mobius/pkg/utils"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileEventStore struct {
	dir      string
	eventsCh chan Event
	doneCh   chan struct{}
	mu       sync.RWMutex
	closed   bool
}

func NewFileEventStore(dir string, bufferSize int) (*FileEventStore, error) {
	if dir == "" {
		dir = ".mobius/events"
	}
	if bufferSize <= 0 {
		bufferSize = 256
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create events directory: %w", err)
	}

	store := &FileEventStore{
		dir:      dir,
		eventsCh: make(chan Event, bufferSize),
		doneCh:   make(chan struct{}),
	}

	go store.worker()

	return store, nil
}

func (s *FileEventStore) writeSync(event Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	filePath := filepath.Join(s.dir, fmt.Sprintf("%s.jsonl", event.ThreadID))
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open event file: %w", err)
	}
	defer f.Close()
	// Streams JSON directly to file and appends '\n' automatically!
	if err := json.NewEncoder(f).Encode(event); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	return nil
}

func (s *FileEventStore) worker() {
	defer close(s.doneCh)
	for event := range s.eventsCh {
		_ = s.writeSync(event)
	}
}

func (s *FileEventStore) Append(ctx context.Context, event Event) error {
	s.mu.RLock()
	if s.closed {
		s.mu.RUnlock()
		return fmt.Errorf("Event Store is Closed")
	}
	s.mu.RUnlock()
	if event.ID == "" {
		event.ID = utils.NewEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	select {
	case s.eventsCh <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return s.writeSync(event)
	}

}

func (s *FileEventStore) GetEvents(ctx context.Context, threadID string) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.dir, fmt.Sprintf("%s.jsonl", threadID))
	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open event file: %w", err)
	}
	defer f.Close()

	var events []Event
	scanner := bufio.NewScanner(f)

	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("failed to parse event line: %w", err)
		}
		events = append(events, e)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading event file: %w", err)
	}
	return events, nil
}

// Close gracefully flushes all queued events and shuts down the background worker.
func (s *FileEventStore) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.eventsCh) // 1. Signal worker() loop to finish draining
	s.mu.Unlock()
	<-s.doneCh // 2. Wait until worker finishes writing every last event!
	return nil
}
