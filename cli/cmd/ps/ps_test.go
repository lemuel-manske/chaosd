package ps

import (
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestPsCmdWithNoArgsThenFail(t *testing.T) {
	output, err := executePs(t, sessiontest.StubSessionStore(t))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestPsCmdWithMultipleArgsThenFail(t *testing.T) {
	output, err := executePs(
		t,
		sessiontest.StubSessionStore(t),
		"session1",
		"session2",
	)

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestPsCmdWithNonExistentSessionThenFail(t *testing.T) {
	output, err := executePs(
		t,
		sessiontest.StubSessionStore(t),
		"session1",
	)

	assert.Error(t, err)
	assert.Contains(t, output, `read session "session1"`)
}

func TestLoadCmdWithInvalidYamlThenFail(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := executePsSession(t, file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestPsCmdWithRunningContainerThenPrintTopology(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
`)

	store, sessionID := psSession(t, file)

	cmd := NewPsCmd(
		store,
		dockertest.FakeDockerProvider(
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
		),
	)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID)

	assert.NoError(t, err)

	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
}

func executePs(t *testing.T, store *session.Store, args ...string) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewPsCmd(store, dockertest.EmptyDockerProvider()),
		args...,
	)
}

func executePsSession(t *testing.T, composeFile string) (string, error) {
	t.Helper()

	store, sessionID := psSession(t, composeFile)

	return executePs(t, store, sessionID)
}

func psSession(t *testing.T, composeFile string) (*session.Store, string) {
	t.Helper()

	store := sessiontest.StubSessionStore(t)

	s, err := store.Create("project", composeFile)

	assert.NoError(t, err)

	return store, s.ID
}
