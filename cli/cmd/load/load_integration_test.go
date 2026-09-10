//go:build integration

package load

import (
	"strings"
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/session/sessiontest"

	"chaosd/cli/internal/session"

	"github.com/stretchr/testify/assert"
)

func TestLoadCmd_RunningComposeProject_CreatesSession(t *testing.T) {
	app := dockertest.StartCompose(t, `project-load-1`, `name: project-load-1
services:
  web:
    image: nginx
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	output, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := strings.TrimSpace(output)

	s, err := sessionStore.Get(session.SessionID(sessionID))
	assert.NoError(t, err)

	assert.Equal(t, `project-load-1`, s.Project)
	assert.Equal(t, app.ComposeFile, s.ComposeFile)
}
