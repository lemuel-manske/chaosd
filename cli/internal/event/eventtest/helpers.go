package eventtest

import (
	"testing"

	"chaosd/cli/internal/event"
)

func NewTmpEventStore(t *testing.T) event.EventStore {
	return event.NewFileEventStore(t.TempDir())
}
