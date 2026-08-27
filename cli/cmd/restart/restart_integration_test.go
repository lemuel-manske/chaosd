//go:build integration

package restart

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestartCmd_RunningComposeProjectWithNonexistentService_ReturnsError(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runRestart(t, app.ComposeFile, "nonexistent-service")

	require.Error(t, err)

	assert.Contains(t, output, "service nonexistent-service not found in project project-restart-1")
}

func TestRestartCmd_ExistingService_PrintsRestartedContainer(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runRestart(t, app.ComposeFile, "web")

	assert.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
}

func TestRestartCmd_MultipleReplicas_PrintsAllRestartedContainers(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	output, err := runRestart(t, app.ComposeFile, "web")

	assert.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
	assert.Contains(t, output, "project-restart-1-web-2")
	assert.Contains(t, output, "project-restart-1-web-3")
}

func TestRestartCmd_ExistingService_RestartsContainer(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	containerBeforeRestart :=
		dockertest.ContainerByServiceName(t, "project-restart-1", "web")

	beforeInspect :=
		dockertest.InspectContainer(t, containerBeforeRestart.GetContainerID())

	output, err := runRestart(t, app.ComposeFile, "web")

	assert.NoError(t, err)
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

func runRestart(t *testing.T, composeFile string, serviceName string) (string, error) {
	t.Helper()

	store := sessiontest.CreateStubStore(t)

	s, err := store.Create("project", composeFile)

	assert.NoError(t, err)

	cmd := NewRestartCmd(store, dockertest.RealDockerProvider())

	return clitest.ExecuteCommand(t, cmd, s.ID, serviceName)
}
