//go:build !integration

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

func TestPsCmd_NoArguments_ReturnsError(t *testing.T) {
	output, err := executePs(t, sessiontest.CreateStubStore(t))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestPsCmd_MultipleArguments_ReturnsError(t *testing.T) {
	output, err := executePs(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"session2",
	)

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestPsCmd_NonexistentSession_ReturnsError(t *testing.T) {
	output, err := executePs(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
	)

	assert.Error(t, err)

	assert.Contains(t, output, `read session "session1"`)
}

func TestPsCmd_RunningContainer_PrintsTopology(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

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

	output, err := clitest.ExecuteCommand(t, cmd, session.ID)

	assert.NoError(t, err)

	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
}

func executePs(t *testing.T, store session.Store, args ...string) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewPsCmd(store, dockertest.EmptyDockerProvider()),
		args...,
	)
}
