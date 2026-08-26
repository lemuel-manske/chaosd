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

func TestRestartCmdWithNoArgsThenFail(t *testing.T) {
	output, err := executeRestart(t, sessiontest.StubSessionStore(t))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 0")
}

func TestRestartCmdWithOneArgThenFail(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.StubSessionStore(t),
		"session1",
	)

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 1")
}

func TestRestartCmdWithThreeArgsThenFail(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.StubSessionStore(t),
		"session1",
		"service1",
		"extra",
	)

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 2 arg(s), received 3")
}

func TestRestartCmdWithNonExistentSessionThenFail(t *testing.T) {
	output, err := executeRestart(
		t,
		sessiontest.StubSessionStore(t),
		"session1",
		"service1",
	)

	assert.Error(t, err)
	assert.Contains(t, output, `read session "session1"`)
}

func TestRestartCmdWithInvalidYamlThenFail(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := executeRestartSession(t, file, "app")

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestRestartCmdWithNonExistentServiceAndNoContainersRunningThenFail(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	output, err := executeRestartSession(t, file, "app")

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

	store, sessionID := restartSession(t, file)

	cmd := NewRestartCmd(store, dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "web")

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

	store, sessionID := restartSession(t, file)

	cmd := NewRestartCmd(store, dockerProvider)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "web")

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

	store, sessionID := restartSession(t, file)

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

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "web")

	assert.NoError(t, err)

	assert.Contains(t, output, "chaosd-app-1 restarted")
}

func executeRestart(
	t *testing.T,
	store *session.Store,
	args ...string,
) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewRestartCmd(store, dockertest.EmptyDockerProvider()),
		args...,
	)
}

func executeRestartSession(
	t *testing.T,
	composeFile string,
	serviceName string,
) (string, error) {
	t.Helper()

	store, sessionID := restartSession(t, composeFile)

	return executeRestart(t, store, sessionID, serviceName)
}

func restartSession(t *testing.T, composeFile string) (*session.Store, string) {
	t.Helper()

	store := sessiontest.StubSessionStore(t)

	s, err := store.Create("project", composeFile)

	assert.NoError(t, err)

	return store, s.ID
}
