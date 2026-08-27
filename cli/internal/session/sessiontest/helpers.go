package sessiontest

import (
	"testing"

	"chaosd/cli/internal/session"
)

func CreateRealStore(t *testing.T) session.Store {
	t.Helper()

	store, err := session.NewDefaultStore()

	if err != nil {
		t.Fatalf("failed to create default session store: %v", err)
	}

	return store
}

func CreateStubStore(t *testing.T) session.Store {
	t.Helper()

	return session.NewStore(t.TempDir())
}
