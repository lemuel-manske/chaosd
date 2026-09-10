package event_test

import (
	"testing"
	"time"

	"chaosd/cli/internal/event"
	"chaosd/cli/internal/session"

	"chaosd/cli/internal/event/eventtest"

	"github.com/stretchr/testify/assert"
)

func TestAppendEvent(t *testing.T) {
	store := eventtest.NewTmpEventStore(t)

	sessionID := session.SessionID("test-session")

	err := store.Append(sessionID, event.Event{
		Type:      event.RestartEvent,
		CreatedAt: time.Now(),
		Data:      event.RestartEventData{ServiceName: "web"},
	})

	assert.NoError(t, err)
}

func TestAppendEvent_AndList(t *testing.T) {
	store := eventtest.NewTmpEventStore(t)

	sessionID := session.SessionID("test-session")

	err := store.Append(sessionID, event.Event{
		Type:      event.RestartEvent,
		CreatedAt: time.Now(),
		Data:      event.RestartEventData{ServiceName: "web"},
	})

	assert.NoError(t, err)

	events, err := store.List(sessionID)

	assert.NoError(t, err)
	assert.Len(t, events, 1)
}
