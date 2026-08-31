package sessiontest

import (
	"testing"

	"chaosd/cli/internal/session"
)

func NewTmpSessionStore(t *testing.T) session.SessionStore {
	return session.NewStore(t.TempDir())
}
