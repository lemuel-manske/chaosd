//go:build integration

package events

import (
	"strings"
	"testing"

	"chaosd/cli/application"
	"chaosd/cli/cmd/load"
	"chaosd/cli/cmd/partition"
	"chaosd/cli/cmd/restart"
	"chaosd/cli/internal/event"
	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
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
	eventStore := eventtest.NewTmpEventStore(t)

	loadOutput, err := runLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := strings.TrimSpace(loadOutput)

	_, err = runRestart(t, sessionStore, eventStore, sessionID, "web-1")
	assert.NoError(t, err)

	_, err = runPartition(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-events-1-web-1-1",
		"project-events-1-web-2-1",
	)
	assert.NoError(t, err)

	output, err := runEvents(t, sessionStore, eventStore, sessionID)
	assert.NoError(t, err)

	_, err = runHeal(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-events-1-web-1-1",
		"project-events-1-web-2-1",
	)
	assert.NoError(t, err)

	assert.Contains(t, output, "TIME\tTYPE\tTARGET")
	assert.Contains(t, output, "restart\tweb-1")
	assert.Contains(t, output, "partition\tproject-events-1-web-1-1")
	assert.Contains(t, output, "partition\tproject-events-1-web-2-1")
}

func runEvents(t *testing.T, sessionStore session.SessionStore, eventStore event.EventStore, composeFile string) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	cmd := NewEventsCmd(app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}

func runRestart(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID string,
	serviceName string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	cmd := restart.NewRestartCmd(app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		serviceName,
	)
}

func runLoad(t *testing.T, sessionStore session.SessionStore, eventStore event.EventStore, composeFile string) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	cmd := load.NewLoadCmd(app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}

func runPartition(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID string,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	cmd := partition.NewPartitionCmd(app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		nodeA,
		nodeB,
	)
}

func runHeal(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID string,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockertest.NewRealDockerProvider()),
		application.WithNetworkManager(networktest.NewRealManager()),
	)

	return clitest.ExecuteCommand(
		t,
		partition.NewHealCmd(app),
		sessionID,
		nodeA,
		nodeB,
	)
}
