//go:build !integration

package load

import (
	"strings"
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/google/uuid"
	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestLoadCmd_NoArguments_ReturnsError(t *testing.T) {
	cmd := NewLoadCmd(sessiontest.CreateStubStore(t), dockertest.EmptyDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd)

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestLoadCmd_MultipleArguments_ReturnsError(t *testing.T) {
	cmd := NewLoadCmd(sessiontest.CreateStubStore(t), dockertest.EmptyDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd, "file1", "file2")

	assert.Error(t, err)

	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestLoadCmd_NonexistentFile_ReturnsError(t *testing.T) {
	cmd := NewLoadCmd(sessiontest.CreateStubStore(t), dockertest.EmptyDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd, "file1")

	assert.Error(t, err)

	assert.Contains(t, output, "file file1 does not exist")
}

func TestLoadCmd_InvalidYAML_ReturnsError(t *testing.T) {
	file := clitest.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	cmd := NewLoadCmd(sessiontest.CreateStubStore(t), dockertest.EmptyDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd, file)

	assert.Error(t, err)

	assert.Contains(t, output, "failed to parse file")
}

func TestLoadCmd_ValidYAMLWithoutContainers_Succeeds(t *testing.T) {
	file := clitest.File(t, `services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(sessiontest.CreateStubStore(t), dockertest.EmptyDockerProvider())

	_, err := clitest.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
}

func TestLoadCmd_YAMLWithoutServices_ReturnsError(t *testing.T) {
	file := clitest.File(t, `version: "3.8"`)

	sessionStore := sessiontest.CreateStubStore(t)
	cmd := NewLoadCmd(sessionStore, dockertest.EmptyDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd, file)

	assert.Error(t, err)

	assert.Contains(t, output, "no services defined in file")
}

func TestLoadCmd_UnreachableDockerDaemon_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: test
services:
  web:
    image: nginx
`)

	sessionStore := sessiontest.CreateStubStore(t)
	cmd := NewLoadCmd(sessionStore, dockertest.UnreachableDockerProvider())

	output, err := clitest.ExecuteCommand(t, cmd, file)

	assert.Error(t, err)

	assert.Contains(t, output, "failed to ping docker daemon")
}

func TestLoadCmd_ValidComposeFile_CreatesSession(t *testing.T) {
	f := clitest.File(t, `name: project-load-1
services:
  web:
    image: nginx
`)

	sessionStore := sessiontest.CreateStubStore(t)

	cmd := NewLoadCmd(
		sessionStore,
		dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-app-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-load-1",
					},
					State: "running",
				},
			},
		))

	output, err := clitest.ExecuteCommand(t, cmd, f)

	assert.NoError(t, err)

	sessionID := strings.TrimSpace(output)
	_, err = uuid.Parse(sessionID)

	assert.NoError(t, err)

	s, err := sessionStore.Get(sessionID)

	assert.NoError(t, err)

	assert.Equal(t, "project-load-1", s.Project)
	assert.Equal(t, f, s.ComposeFile)
}

func TestLoadCmd_SameComposeFileLoadedTwice_CreatesDistinctSessions(t *testing.T) {
	f := clitest.File(t, `name: project-load-1
services:
  web:
    image: nginx
`)

	sessionStore := sessiontest.CreateStubStore(t)

	cmd := NewLoadCmd(
		sessionStore,
		dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-app-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-load-1",
					},
					State: "running",
				},
			},
		))

	output1, err := clitest.ExecuteCommand(t, cmd, f)
	assert.NoError(t, err)

	output2, err := clitest.ExecuteCommand(t, cmd, f)
	assert.NoError(t, err)

	sessionID1 := strings.TrimSpace(output1)
	sessionID2 := strings.TrimSpace(output2)

	assert.NotEqual(t, sessionID1, sessionID2)
}
