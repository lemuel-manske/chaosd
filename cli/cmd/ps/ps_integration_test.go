package ps

import (
	"chaosd/cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPsWithRealCompose(t *testing.T) {
	app := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	test.AssertLineCountContains(t, output, 1, "web", "running")
}

func TestPsWithMultipleProjects(t *testing.T) {
	app1 := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
`)

	test.StartComposeApp(t, `project2`, `
name: project2
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app1.ComposeFile)

	require.NoError(t, err)

	test.AssertLineCountContains(t, output, 1, "web", "running")

	assert.NotContains(t, output, "project2")
}

func TestPsWithMultipleReplicas(t *testing.T) {
	app := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
    deploy:
      replicas: 3
`)

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	test.AssertLineCountContains(t, output, 3, "web", "running")
}

func TestPsWithStoppedContainer(t *testing.T) {
	app := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
`)

	test.StopContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	test.AssertLineCountContains(t, output, 1, "web", "exited")
}

func TestPsWithMissingContainer(t *testing.T) {
	app := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
`)

	test.RemoveContainerByServiceName(t, app.ProjectName, "web")

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)

	test.AssertLineCountContains(t, output, 1, "web", "missing")
}

func runPs(t *testing.T, composeFile string) (string, error) {
	t.Helper()

	cmd := NewPsCmd(test.RealDockerProvider())

	return test.ExecuteCommand(t, cmd, composeFile)
}
