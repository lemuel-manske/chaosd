//go:build integration

package delay

import (
	"testing"
	"time"

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

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	delayDuration := 500 * time.Millisecond

	before, err := dockertest.MeasureRequestDuration(t, "project-delay-1", "web-1", "http://web-2")
	assert.NoError(t, err)

	_, err = clitest.RunDelay(
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

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	delayDuration := 500 * time.Millisecond

	before, err := dockertest.MeasureRequestDuration(t, "project-delay-2", "web-1", "http://web-2")
	assert.NoError(t, err)

	delayOutput, err := clitest.RunDelay(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-delay-2-web-1-1",
		"project-delay-2-web-2-1",
		delayDuration,
	)
	assert.NoError(t, err)

	faultID := session.ParseFaultID(delayOutput)

	after, err := dockertest.MeasureRequestDuration(t, "project-delay-2", "web-1", "http://web-2")
	assert.NoError(t, err)

	assert.GreaterOrEqual(
		t,
		after-before,
		450*time.Millisecond,
	)

	_, err = clitest.RunHeal(
		t,
		sessionStore,
		eventStore,
		sessionID,
		faultID,
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
