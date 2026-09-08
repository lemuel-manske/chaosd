//go:build integration

package delay

import (
	"strings"
	"testing"
	"time"

	"chaosd/cli/application"
	"chaosd/cli/cmd/load"
	"chaosd/cli/internal/event"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestDelayCmd_RunningComposeProject_DelaysTraffic(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-delay-1`, `name: project-delay-1
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

	delayDuration := 500 * time.Millisecond

	before, err := dockertest.MeasureRequestDuration(t, "project-delay-1", "web-1", "http://web-2")
	assert.NoError(t, err)

	_, err = runDelay(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-delay-1-web-1-1",
		"project-delay-1-web-2-1",
		delayDuration,
	)
	assert.NoError(t, err)

	after, err := dockertest.MeasureRequestDuration(t, "project-delay-1", "web-1", "http://web-2")
	assert.NoError(t, err)

	assert.GreaterOrEqual(
		t,
		after-before,
		450*time.Millisecond,
	)
}

func TestDelayCmd_RunningComposeProject_DelaysTraffic_AndHeals(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-delay-2`, `name: project-delay-2
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

	delayDuration := 500 * time.Millisecond

	before, err := dockertest.MeasureRequestDuration(t, "project-delay-2", "web-1", "http://web-2")
	assert.NoError(t, err)

	_, err = runDelay(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-delay-2-web-1-1",
		"project-delay-2-web-2-1",
		delayDuration,
	)
	assert.NoError(t, err)

	after, err := dockertest.MeasureRequestDuration(t, "project-delay-2", "web-1", "http://web-2")
	assert.NoError(t, err)

	assert.GreaterOrEqual(
		t,
		after-before,
		450*time.Millisecond,
	)

	_, err = runHeal(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-delay-2-web-1-1",
		"project-delay-2-web-2-1",
	)
	assert.NoError(t, err)

	afterHeal, err := dockertest.MeasureRequestDuration(t, "project-delay-2", "web-1", "http://web-2")
	assert.NoError(t, err)

	assert.LessOrEqual(
		t,
		afterHeal-before,
		100*time.Millisecond,
	)
}

func runLoad(
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

	cmd := load.NewLoadCmd(app)

	return clitest.ExecuteCommand(t, cmd, composeFile)
}

func runDelay(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID string,
	nodeAName string,
	nodeBName string,
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

	cmd := NewDelayCmd(app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		nodeAName,
		nodeBName,
		delayDuration.String(),
	)
}

func runHeal(
	t *testing.T,
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	sessionID string,
	nodeAName string,
	nodeBName string,
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

	cmd := NewHealCmd(app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		sessionID,
		nodeAName,
		nodeBName,
	)
}
