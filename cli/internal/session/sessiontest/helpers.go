package sessiontest

import (
	"testing"

	"chaosd/cli/internal/session"
)

func NewTmpSessionStore(t *testing.T) session.Store {
	return session.NewStore(t.TempDir())
}
