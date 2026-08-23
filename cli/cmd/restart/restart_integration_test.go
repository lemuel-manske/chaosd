package restart

import (
	"chaosd/cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestartCmdWithNonExistentServiceThenFail(t *testing.T) {
	app := test.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runRestart(t, app.ComposeFile, "nonexistent-service")

	require.Error(t, err)

	assert.Contains(t, output, "service nonexistent-service not found in project project-restart-1")
}

func TestRestartCmdWithExistingServiceThenSucceed(t *testing.T) {
	app := test.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runRestart(t, app.ComposeFile, "web")

	require.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
}

func TestRestartCmdWithMultipleReplicasThenSucceed(t *testing.T) {
	app := test.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	output, err := runRestart(t, app.ComposeFile, "web")

	require.NoError(t, err)

	assert.Contains(t, output, "project-restart-1-web-1")
	assert.Contains(t, output, "project-restart-1-web-2")
	assert.Contains(t, output, "project-restart-1-web-3")
}

func TestRestartCmdEnsureServiceIsIndeedRestarted(t *testing.T) {
	app := test.StartComposeApp(t, `project-restart-1`, `
name: project-restart-1
services:
  web:
    image: nginx:alpine
`)

	containerBeforeRestart := test.ContainerByServiceName(t, "project-restart-1", "web")
	beforeInspect := test.InspectContainer(t, containerBeforeRestart.GetContainerID())

	output, err := runRestart(t, app.ComposeFile, "web")

	require.NoError(t, err)
	assert.Contains(t, output, "project-restart-1-web-1")

	containerAfterRestart := test.ContainerByServiceName(t, "project-restart-1", "web")
	afterInspect := test.InspectContainer(t, containerAfterRestart.GetContainerID())

	assert.Equal(t, containerBeforeRestart.GetContainerID(), containerAfterRestart.GetContainerID())

	assert.NotEqual(t, beforeInspect.Container.State.StartedAt, afterInspect.Container.State.StartedAt)
}

func runRestart(t *testing.T, composeFile string, serviceName string) (string, error) {
	t.Helper()

	cmd := NewRestartCmd(test.RealDockerProvider())

	return test.ExecuteCommand(t, cmd, composeFile, serviceName)
}
