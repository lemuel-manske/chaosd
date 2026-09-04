package event

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chaosd/cli/internal/session"
	"chaosd/cli/internal/storage"
)

type EventType string

const (
	PartitionAppliedEvent EventType = "partition"
	HealAppliedEvent      EventType = "heal"

	RestartEvent EventType = "restart"
)

type PartitionAppliedEventData struct {
	NodeAName string
	NodeBName string
}

type HealAppliedEventData struct {
	NodeAName string
	NodeBName string
}

type RestartEventData struct {
	ServiceName string
}

type Event struct {
	Type      EventType `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Data      any       `json:"data"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Event struct.
func (e *Event) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type      EventType       `json:"type"`
		CreatedAt time.Time       `json:"created_at"`
		Data      json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	e.Type = aux.Type
	e.CreatedAt = aux.CreatedAt

	switch e.Type {
	case PartitionAppliedEvent:
		var d PartitionAppliedEventData
		if err := json.Unmarshal(aux.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case HealAppliedEvent:
		var d HealAppliedEventData
		if err := json.Unmarshal(aux.Data, &d); err != nil {
			return err
		}
		e.Data = d
	case RestartEvent:
		var d RestartEventData
		if err := json.Unmarshal(aux.Data, &d); err != nil {
			return err
		}
		e.Data = d
	default:
		return fmt.Errorf("unknown event type: %s", e.Type)
	}

	return nil
}

type EventStore interface {
	Append(sessionID session.SessionID, event Event) error
	List(sessionID session.SessionID) ([]Event, error)
}

type InMemoryEventStore struct {
	events map[session.SessionID][]Event
}

func NewInMemoryEventStore() EventStore {
	return &InMemoryEventStore{
		events: make(map[session.SessionID][]Event),
	}
}

func (s *InMemoryEventStore) Append(sessionID session.SessionID, event Event) error {
	s.events[sessionID] = append(s.events[sessionID], event)
	return nil
}

func (s *InMemoryEventStore) List(sessionID session.SessionID) ([]Event, error) {
	return s.events[sessionID], nil
}

type FileEventStore struct {
	writer storage.AtomicWriter
	dir    string
}

func NewFileEventStore(dir string) EventStore {
	writer := storage.NewAtomicFileWriter()

	return &FileEventStore{dir: dir, writer: writer}
}

func (s *FileEventStore) Append(sessionID session.SessionID, event Event) error {
	events, err := s.List(sessionID)

	if err != nil {
		return err
	}

	events = append(events, event)

	path, err := s.createPathToEvents(sessionID)

	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(events, "", "  ")

	if err != nil {
		return fmt.Errorf("encode events: %w", err)
	}

	if err := s.writer.Write(path, data, 0600); err != nil {
		return fmt.Errorf("write events: %w", err)
	}

	return nil
}

func (s *FileEventStore) List(sessionID session.SessionID) ([]Event, error) {
	path, err := s.createPathToEvents(sessionID)

	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, fmt.Errorf("read events %q: %w", sessionID, err)
	}

	var events []Event

	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("decode events %q: %w", sessionID, err)
	}

	return events, nil
}

func (s *FileEventStore) createPathToEvents(sessionID session.SessionID) (string, error) {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return "", fmt.Errorf("create events directory: %w", err)
	}

	path := filepath.Join(s.dir, string(sessionID)+".json")

	return path, nil
}
