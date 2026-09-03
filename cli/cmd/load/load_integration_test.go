//go:build integration

package load

import (
	"strings"
	"testing"

	"chaosd/cli/application"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"chaosd/cli/internal/session"

	"github.com/stretchr/testify/assert"
)

func TestLoadCmd_RunningComposeProject_CreatesSession(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-load-1`, `name: project-load-1
services:
  web:
    image: nginx
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	output, err := runLoad(t, sessionStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := strings.TrimSpace(output)

	s, err := sessionStore.Get(session.SessionID(sessionID))
	assert.NoError(t, err)

	assert.Equal(t, `project-load-1`, s.Project)
	assert.Equal(t, app.ComposeFile, s.ComposeFile)
}

func runLoad(t *testing.T, sessionStore session.SessionStore, composeFile string) (string, error) {
	t.Helper()

	eventStore := eventtest.NewTmpEventStore(t)
	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		eventStore,
		dockerProvider,
		networkManager,
	)

	cmd := NewLoadCmd(*app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}
