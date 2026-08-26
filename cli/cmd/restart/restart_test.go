package restart

import (
	"fmt"
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestRestartCmdWithNoArgsThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 0")
}

func TestRestartCmdWithOneArgThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()), "file1")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 1")
}

func TestRestarTCmdWithThreeArgsThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()), "file1", "service1", "extra")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 3")
}

func TestRestartCmdWithNonExistentFileThenFail(t *testing.T) {
	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()), "file1", "service1")

	assert.Error(t, err)
	assert.Contains(t, output, "file file1 does not exist")
}

func TestRestartCmdWithInvalidYamlThenFail(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()), file, "app")

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestRestartCmdWithNonExistentServiceAndNoContainersRunningThenFail(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	output, err := clitest.ExecuteCommand(t, NewRestartCmd(dockertest.EmptyDockerProvider()), file, "app")

	assert.Error(t, err)
	assert.Contains(t, output, "service app not found in project project-restart-1")
}

func TestRestartCmdWithRunningContainerThenFailToRestart(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	dockerProvider := &dockertest.DockerProviderMock{
		Client: &dockertest.DockerClientMock{
			Containers: []container.Summary{
				{
					ID:    "1",
					Names: []string{"chaosd-app-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-restart-1",
					},
					State: "running",
				},
			},

			RestartErr: map[string]error{
				"1": fmt.Errorf("restart failed"),
			},
		},
	}

	cmd := NewRestartCmd(dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, file, "web")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-app-1 failed to restart")
}

func TestRestartCmdWithMultipleReplicasThenFailToRestartOne(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
    deploy:
      replicas: 3
`)

	dockerProvider := &dockertest.DockerProviderMock{
		Client: &dockertest.DockerClientMock{
			Containers: []container.Summary{
				{
					ID:    "1",
					Names: []string{"chaosd-app-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-restart-1",
					},
					State: "running",
				},
				{
					ID:    "2",
					Names: []string{"chaosd-app-2"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-restart-1",
					},
					State: "running",
				},
				{
					ID:    "3",
					Names: []string{"chaosd-app-3"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-restart-1",
					},
					State: "running",
				},
			},

			RestartErr: map[string]error{
				"2": fmt.Errorf("restart failed"),
			},
		},
	}

	cmd := NewRestartCmd(dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, file, "web")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-app-1 restarted")
	assert.Contains(t, output, "chaosd-app-2 failed to restart")
	assert.Contains(t, output, "chaosd-app-3 restarted")
}

func TestRestartCmdWithRunningContainerThenPrintContainerName(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	cmd := NewRestartCmd(dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project-restart-1",
				},
				State: "running",
			},
		},
	))

	output, err := clitest.ExecuteCommand(t, cmd, file, "web")

	assert.NoError(t, err)

	assert.Contains(t, output, "chaosd-app-1 restarted")
}
