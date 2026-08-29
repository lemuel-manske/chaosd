package partition

import (
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/moby/moby/api/types/container"

	"github.com/stretchr/testify/assert"
)

func TestPartitionCmd_NoArguments_ReturnsError(t *testing.T) {
	output, err := executePartition(t, sessiontest.CreateStubStore(t))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 3 arg(s), received 0")
}

func TestPartitionCmd_TooManyArguments_ReturnsError(t *testing.T) {
	output, err := executePartition(
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

func TestPartitionCmd_NonexistentSession_ReturnsError(t *testing.T) {
	output, err := executePartition(
		t,
		sessiontest.CreateStubStore(t),
		"session1",
		"service-a",
		"service-b",
	)

	assert.Error(t, err)
	assert.Contains(t, output, `read session "session1"`)
}

func TestPartitionCmd_UnknownNodeNames_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	s, _ := store.Create("project", file)

	cmd := NewPartitionCmd(
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

	sessionID := string(s.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "wrong-node-name", "another-wrong-node-name")

	assert.Error(t, err)

	assert.Contains(t, output, "wrong-node-name missing")
}

func TestPartitionCmd_StoppedNodes_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	s, _ := store.Create("project", file)

	cmd := NewPartitionCmd(
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

	sessionID := string(s.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.Error(t, err)

	assert.Contains(t, output, "chaosd-web-1 is not running")
}

func TestPartitionCmd_ValidArguments_ReturnsSuccess(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	s, _ := store.Create("project", file)

	cmd := NewPartitionCmd(
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

	sessionID := string(s.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.NoError(t, err)

	assert.Contains(t, output, "chaosd-web-1 and chaosd-db-1 partitioned")
}

func TestPartitionCmd_ValidArguments_StoresFault(t *testing.T) {
	file := clitest.File(t, `name: project-ps-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	manager := networktest.NewStubManager()

	store := sessiontest.CreateStubStore(t)
	s, _ := store.Create("project", file)

	cmd := NewPartitionCmd(
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

	sessionID := string(s.ID)

	output, err := clitest.ExecuteCommand(t, cmd, sessionID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	assert.Contains(t, output, "chaosd-web-1 and chaosd-db-1 partitioned")

	sessionAfterPartition, err := store.Get(s.ID)
	assert.NoError(t, err)

	assert.Len(t, sessionAfterPartition.Faults, 1)

	actualFault := sessionAfterPartition.Faults[0]

	expectedFault := session.Fault{
		ID:     actualFault.ID,
		Type:   "partition",
		NodeA:  "chaosd-web-1",
		NodeB:  "chaosd-db-1",
		Status: "active",
	}

	assert.Equal(t, expectedFault, actualFault)
}

func executePartition(t *testing.T, store session.Store, args ...string) (string, error) {
	t.Helper()

	return clitest.ExecuteCommand(
		t,
		NewPartitionCmd(store, dockertest.EmptyDockerProvider(), networktest.NewStubManager()),
		args...,
	)
}
