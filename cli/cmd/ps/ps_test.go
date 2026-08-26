package ps

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestPsCmdWithNoArgsThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewPsCmd(dockertest.EmptyDockerProvider()))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestPsCmdWithMultipleArgsThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewPsCmd(dockertest.EmptyDockerProvider()), "file1", "file2")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestPsCmdWithNonExistentFileThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewPsCmd(dockertest.EmptyDockerProvider()), "file1")

	assert.Error(t, err)
	assert.Contains(t, output, "file file1 does not exist")
}

func TestLoadCmdWithInvalidYamlThenFail(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := clitest.ExecuteCommand(t, NewPsCmd(dockertest.EmptyDockerProvider()), file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestPsCmdWithRunningContainerThenPrintTopology(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
`)

	cmd := NewPsCmd(dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project-ps-1",
				},
				State: "running",
			},
		},
	))

	output, err := clitest.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)

	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
}
