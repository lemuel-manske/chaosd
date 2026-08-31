package application

import (
	"time"

	"chaosd/cli/internal/session"
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
	Type      EventType
	CreatedAt time.Time
	Data      any
}

type EventStore interface {
	Append(sessionID session.SessionID, event Event) error
	List(sessionID session.SessionID) ([]Event, error)
}

type InMemoryEventStore struct {
	events map[session.SessionID][]Event
}

func NewInMemoryEventStore() *InMemoryEventStore {
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

type FileEventStore struct{}
