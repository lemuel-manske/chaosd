package partition

import (
	"testing"

	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/moby/moby/api/types/container"

	"github.com/stretchr/testify/assert"
)

func TestHealCmd_NoArguments_ReturnsError(t *testing.T) {
	output, err := executeHeal(t, sessiontest.CreateStubStore(t))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 3 arg(s), received 0")
}

func TestHealCmd_TooManyArguments_ReturnsError(t *testing.T) {
	output, err := executeHeal(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"session2",
		"session3",
		"session4",
	)

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 3 arg(s), received 4")
}

func TestHealCmd_NonexistentSession_ReturnsError(t *testing.T) {
	output, err := executeHeal(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"service-a",
		"service-b",
	)

	assert.Error(t, err)
	assert.Contains(t, output, `read session "session1"`)
}

func TestHealCmd_UnknownNodeNames_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	cmd := NewHealCmd(
		store,
		dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "running",
				},
				{
					ID:    "0987654321",
					Names: []string{"chaosd-db-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "db",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "running",
				},
			},
		),
		manager,
	)

	sessionID := string(session.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "wrong-node-name", "another-wrong-node-name")

	assert.Error(t, err)

	assert.Contains(t, output, "wrong-node-name missing")
}

func TestHealCmd_StoppedNodes_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	cmd := NewHealCmd(
		store,
		dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "exited",
				},
				{
					ID:    "0987654321",
					Names: []string{"chaosd-db-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "db",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "exited",
				},
			},
		),
		manager,
	)

	sessionID := string(session.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-web-1 is not running")
}

func TestHealCmd_NodesNotPartitioned_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	cmd := NewHealCmd(
		store,
		dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "running",
				},
				{
					ID:    "0987654321",
					Names: []string{"chaosd-db-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "db",
						"com.docker.compose.project": "project-ps-1",
					},
					State: "running",
				},
			},
		),
		manager,
	)

	sessionID := string(session.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.Error(t, err)

	assert.Contains(t, output, "no partition fault found between chaosd-web-1 and chaosd-db-1")
}

func TestHealCmd_PartitionedNodes_HealsSuccessfully(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-web-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project-ps-1",
				},
				State: "running",
			},
			{
				ID:    "0987654321",
				Names: []string{"chaosd-db-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "db",
					"com.docker.compose.project": "project-ps-1",
				},
				State: "running",
			},
		},
	)

	partitionCmd := NewPartitionCmd(
		store,
		dockerProvider,
		manager,
	)

	healCmd := NewHealCmd(
		store,
		dockerProvider,
		manager,
	)

	sessionID := string(session.ID)

	partitionOutput, partitionErr := clitest.ExecuteCommand(
		t,
		partitionCmd,
		sessionID,
		"chaosd-web-1",
		"chaosd-db-1",
	)

	assert.NoError(t, partitionErr)
	assert.Contains(t, partitionOutput, "chaosd-web-1 and chaosd-db-1 partitioned")

	healOutput, healErr := clitest.ExecuteCommand(
		t,
		healCmd,
		sessionID,
		"chaosd-web-1",
		"chaosd-db-1",
	)

	assert.NoError(t, healErr)
	assert.Contains(t, healOutput, "chaosd-web-1 and chaosd-db-1 healed")
}

func TestHealCmd_AlreadyHealedNodes_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	session, _ := store.Create("project", file)

	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-web-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project-ps-1",
				},
				State: "running",
			},
			{
				ID:    "0987654321",
				Names: []string{"chaosd-db-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "db",
					"com.docker.compose.project": "project-ps-1",
				},
				State: "running",
			},
		},
	)

	partitionCmd := NewPartitionCmd(
		store,
		dockerProvider,
		manager,
	)

	healCmd := NewHealCmd(
		store,
		dockerProvider,
		manager,
	)

	sessionID := string(session.ID)

	partitionOutput, partitionErr := clitest.ExecuteCommand(
		t,
		partitionCmd,
		sessionID,
		"chaosd-web-1",
		"chaosd-db-1",
	)

	assert.NoError(t, partitionErr)
	assert.Contains(t, partitionOutput, "chaosd-web-1 and chaosd-db-1 partitioned")

	healOutput, healErr := clitest.ExecuteCommand(
		t,
		healCmd,
		sessionID,
		"chaosd-web-1",
		"chaosd-db-1",
	)

	assert.NoError(t, healErr)
	assert.Contains(t, healOutput, "chaosd-web-1 and chaosd-db-1 healed")

	secondHealOutput, secondHealErr := clitest.ExecuteCommand(
		t,
		healCmd,
		sessionID,
		"chaosd-web-1",
		"chaosd-db-1",
	)

	assert.Error(t, secondHealErr)
	assert.Contains(t, secondHealOutput, "partition fault between chaosd-web-1 and chaosd-db-1 is already healed")
}

func executeHeal(t *testing.T, store session.Store, args ...string) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewHealCmd(store, dockertest.EmptyDockerProvider(), networktest.NewStubManager()),
		args...,
	)
}
