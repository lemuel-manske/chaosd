package adapters

import (
	"fmt"
	"os"
	"path/filepath"

	"chaosd/cli/internal/event"
	"chaosd/cli/internal/session"
)

const (
	homeDir = ".chaosd"

	sessionsDir = "sessions"
	eventsDir   = "events"
)

func NewDefaultSessionStore() (session.SessionStore, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	return session.NewFileSessionStore(filepath.Join(home, homeDir, sessionsDir)), nil
}

func NewDefaultEventStore() (event.EventStore, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("get user home dir: %w", err)
	}

	return event.NewFileEventStore(filepath.Join(home, homeDir, eventsDir)), nil
}
