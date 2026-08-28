//go:build !integration

package restart

import (
	"fmt"
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestRestartCmd_NoArguments_ReturnsError(t *testing.T) {
	output, err := executeRestart(t, sessiontest.CreateStubStore(t))

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 2 arg(s), received 0")
}

func TestRestartCmd_OneArgument_ReturnsError(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
	)

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 2 arg(s), received 1")
}

func TestRestartCmd_ThreeArguments_ReturnsError(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"service1",
		"extra",
	)

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 2 arg(s), received 3")
}

func TestRestartCmd_NonexistentSession_ReturnsError(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"service1",
	)

	assert.Error(t, err)

	assert.Contains(t, output, `read session "session1"`)
}

func TestRestartCmd_InvalidYAML_ReturnsError(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	output, err := executeRestart(t, store, string(session.ID), "app")

	assert.Error(t, err)

	assert.Contains(t, output, "failed to parse file")
}

func TestRestartCmd_NonexistentService_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	output, err := executeRestart(t, store, string(session.ID), "app")

	assert.Error(t, err)

	assert.Contains(t, output, "service app not found in project project-restart-1")
}

func TestRestartCmd_ContainerRestartFails_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

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

	cmd := NewRestartCmd(store, dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, string(session.ID), "web")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-app-1 failed to restart")
}

func TestRestartCmd_OneReplicaRestartFails_ReportsEachResult(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
    deploy:
      replicas: 3
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

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

	cmd := NewRestartCmd(store, dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, string(session.ID), "web")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-app-1 restarted")
	assert.Contains(t, output, "chaosd-app-2 failed to restart")
	assert.Contains(t, output, "chaosd-app-3 restarted")
}

func TestRestartCmd_RunningContainer_RestartsAndPrintsContainerName(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	cmd := NewRestartCmd(
		store,
		dockertest.FakeDockerProvider(
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
		),
	)

	output, err := clitest.ExecuteCommand(t, cmd, string(session.ID), "web")

	assert.NoError(t, err)

	assert.Contains(t, output, "chaosd-app-1 restarted")
}

func executeRestart(
	t *testing.T,
	store session.Store,
	args ...string,
) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewRestartCmd(store, dockertest.EmptyDockerProvider()),
		args...,
	)
}
