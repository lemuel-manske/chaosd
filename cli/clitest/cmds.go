package clitest

import (
	"testing"
	"time"

	"chaosd/cli/application"
	"chaosd/cli/internal/event"
	"chaosd/cli/internal/session"

	delay_api "chaosd/cli/cmd/delay/api"
	events_api "chaosd/cli/cmd/events/api"
	heal_api "chaosd/cli/cmd/heal/api"
	load_api "chaosd/cli/cmd/load/api"
	partition_api "chaosd/cli/cmd/partition/api"
	ps_api "chaosd/cli/cmd/ps/api"
	restart_api "chaosd/cli/cmd/restart/api"

	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/test"

	"github.com/stretchr/testify/assert"
)

func RunPartition(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID session.SessionID,
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

	cmd := partition_api.NewPartitionCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		nodeA,
		nodeB,
	)
}

func RunHeal(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID session.SessionID,
	faultID session.FaultID,
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

	cmd := heal_api.NewHealCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		string(faultID),
	)
}

func RunEvents(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID session.SessionID,
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

	cmd := events_api.NewEventsCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
	)
}

func RunRestart(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID session.SessionID,
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

	cmd := restart_api.NewRestartCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		serviceName,
	)
}

func RunLoad(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	composeFile string,
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

	cmd := load_api.NewLoadCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		composeFile,
	)
}

func RunDelay(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID session.SessionID,
	nodeA string,
	nodeB string,
	delayDuration time.Duration,
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

	cmd := delay_api.NewDelayCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		nodeA,
		nodeB,
		delayDuration.String(),
	)
}

func RunPs(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	composeFile string,
) (string, error) {
	t.Helper()

	createdSession, err := sessionStore.Create("project", composeFile)

	assert.NoError(t, err)

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		application.WithSessionStore(sessionStore),
		application.WithEventStore(eventStore),
		application.WithDockerProvider(dockerProvider),
		application.WithNetworkManager(networkManager),
	)

	cmd := ps_api.NewPsCmd(app)

	return test.ExecuteCommand(
		t,
		cmd,
		string(createdSession.ID),
	)
}
