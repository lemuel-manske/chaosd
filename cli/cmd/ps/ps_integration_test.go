package ps

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPsWithRealCompose(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "running")
}

func TestPsWithMultipleProjects(t *testing.T) {
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

	require.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "running")

	assert.NotContains(t, output, "project2")
}

func TestPsWithMultipleReplicas(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 3, "web", "running")
}

func TestPsWithStoppedContainer(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	dockertest.StopContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "exited")
}

func TestPsWithMissingContainer(t *testing.T) {
	app := dockertest.StartComposeApp(t, `project-ps-1`, `
name: project-ps-1
services:
  web:
    image: nginx:alpine
`)

	dockertest.RemoveContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	clitest.AssertLineCountContains(t, output, 1, "web", "missing")
}

func runPs(t *testing.T, composeFile string) (string, error) {
	t.Helper()

	store := sessiontest.StubSessionStore(t)

	s, err := store.Create("project", composeFile)

	require.NoError(t, err)

	cmd := NewPsCmd(store, dockertest.RealDockerProvider())

	return clitest.ExecuteCommand(t, cmd, s.ID)
}
