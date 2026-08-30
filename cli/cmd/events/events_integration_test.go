//go:build integration

package events

import (
	"strings"
	"testing"

	"chaosd/cli/application"
	"chaosd/cli/cmd/load"
	"chaosd/cli/cmd/partition"
	"chaosd/cli/cmd/restart"
	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestEventsCmd_RunningComposeProject_ListsEvents(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-events-1`, `name: project-events-1
services:
  web-1:
    image: nginx
  web-2:
    image: nginx
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	loadOutput, err := runLoad(t, sessionStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := strings.TrimSpace(loadOutput)

	_, err = runRestart(t, sessionStore, sessionID, "web-1")
	assert.NoError(t, err)

	_, err = runPartition(t, sessionStore, sessionID, "web-1", "web-2")
	assert.NoError(t, err)

	output, err := runEvents(t, sessionStore, sessionID)
	assert.NoError(t, err)

	assert.Contains(t, output, "TIME\tTYPE\tTARGET")
	assert.Contains(t, output, "restart\tweb-1")
	assert.Contains(t, output, "partition\tweb-1")
	assert.Contains(t, output, "partition\tweb-2")
}

func runEvents(t *testing.T, sessionStore session.Store, composeFile string) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := NewEventsCmd(*app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}

func runRestart(
	t *testing.T,
	sessionStore session.Store,
	sessionID string,
	serviceName string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := restart.NewRestartCmd(*app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		serviceName,
	)
}

func runLoad(t *testing.T, sessionStore session.Store, composeFile string) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := load.NewLoadCmd(*app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}

func runPartition(
	t *testing.T,
	sessionStore session.Store,
	sessionID string,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := partition.NewPartitionCmd(*app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		nodeA,
		nodeB,
	)
}
