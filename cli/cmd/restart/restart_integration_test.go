//go:build integration

package restart

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestartCmd_RunningComposeProjectWithNonexistentService_ReturnsError(t *testing.T) {
	app := dockertest.StartCompose(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	output, err := clitest.RunRestart(t, sessionStore, eventStore, sessionID, "nonexistent-service")
	require.Error(t, err)

	assert.Contains(t, output, "service nonexistent-service not found in project project-restart-1")
}

func TestRestartCmd_ExistingService_PrintsRestartedContainer(t *testing.T) {
	app := dockertest.StartCompose(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	output, err := clitest.RunRestart(t, sessionStore, eventStore, sessionID, "web")
	require.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
}

func TestRestartCmd_MultipleReplicas_PrintsAllRestartedContainers(t *testing.T) {
	app := dockertest.StartCompose(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	output, err := clitest.RunRestart(t, sessionStore, eventStore, sessionID, "web")
	require.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
	assert.Contains(t, output, "project-restart-1-web-2")
	assert.Contains(t, output, "project-restart-1-web-3")
}

func TestRestartCmd_ExistingService_RestartsContainer(t *testing.T) {
	app := dockertest.StartCompose(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	containerBeforeRestart :=
		dockertest.ContainerByServiceName(t, "project-restart-1", "web")

	beforeInspect :=
		dockertest.InspectContainer(t, containerBeforeRestart.GetContainerID())

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	loadOutput, err := clitest.RunLoad(t, sessionStore, eventStore, app.ComposeFile)
	assert.NoError(t, err)

	sessionID := session.ParseSessionID(loadOutput)

	output, err := clitest.RunRestart(t, sessionStore, eventStore, sessionID, "web")
	require.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")

	containerAfterRestart :=
		dockertest.ContainerByServiceName(t, "project-restart-1", "web")

	afterInspect :=
		dockertest.InspectContainer(t, containerAfterRestart.GetContainerID())

	assert.Equal(
		t, containerBeforeRestart.GetContainerID(), containerAfterRestart.GetContainerID(),
	)

	assert.NotEqual(
		t, beforeInspect.Container.State.StartedAt, afterInspect.Container.State.StartedAt,
	)
}
