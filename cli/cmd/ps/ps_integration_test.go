//go:build integration

package ps

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestPsCmd_RunningComposeProject_PrintsRunningService(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app.ComposeFile)

	assert.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "running")
}

func TestPsCmd_MultipleComposeProjects_PrintsOnlySessionProject(t *testing.T) {
	app1 := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	dockertest.StartComposeApp(t, `project2`, `
name: project2
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app1.ComposeFile)

	assert.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "running")

	assert.NotContains(t, output, "project2")
}

func TestPsCmd_MultipleReplicas_PrintsAllReplicas(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	output, err := runPs(t, app.ComposeFile)

	assert.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 3, "web", "running")
}

func TestPsCmd_StoppedContainer_PrintsExitedStatus(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	dockertest.StopContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	assert.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "exited")
}

func TestPsCmd_MissingContainer_PrintsMissingStatus(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	dockertest.RemoveContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	assert.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "missing")
}

func runPs(t *testing.T, composeFile string) (string, error) {
	t.Helper()

	store := sessiontest.CreateStubStore(t)

	s, err := store.Create("project", composeFile)

	assert.NoError(t, err)

	cmd := NewPsCmd(store, dockertest.RealDockerProvider())

	return clitest.ExecuteCommand(t, cmd, string(s.ID))
}
