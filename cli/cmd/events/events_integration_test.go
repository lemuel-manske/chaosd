//go:build integration

package events

import (
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestEventsCmd_RunningComposeProject_ListsEvents(t *testing.T) {
	app := dockertest.StartCompose(t, `project-events-1`, `name: project-events-1
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

	_, err = clitest.RunRestart(t, sessionStore, eventStore, sessionID, "web-1")
	assert.NoError(t, err)

	partitionOutput, err := clitest.RunPartition(
		t,
		sessionStore,
		eventStore,
		sessionID,
		"project-events-1-web-1-1",
		"project-events-1-web-2-1",
	)
	assert.NoError(t, err)

	faultID := session.ParseFaultID(partitionOutput)

	output, err := clitest.RunEvents(t, sessionStore, eventStore, sessionID)
	assert.NoError(t, err)

	_, err = clitest.RunHeal(
		t,
		sessionStore,
		eventStore,
		sessionID,
		faultID,
	)
	assert.NoError(t, err)

	assert.Contains(t, output, "TIME\tTYPE\tTARGET")
	assert.Contains(t, output, "restart\tweb-1")
	assert.Contains(t, output, "partition\tproject-events-1-web-1-1")
	assert.Contains(t, output, "partition\tproject-events-1-web-2-1")
}
