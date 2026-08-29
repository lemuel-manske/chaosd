package application

import (
	"context"
	"testing"

	"chaosd/cli/internal/topology"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestGetTopology_NonexistentSession_ReturnsError(t *testing.T) {
	app := Application{
		SessionStore:   sessiontest.CreateStubStore(t),
		DockerProvider: dockertest.EmptyDockerProvider(),
	}

	topology, err := app.GetTopology(context.Background(), "session1")

	assert.Nil(t, topology)

	assert.Error(t, err)

	assert.Contains(t, err.Error(), `read session "session1"`)
}

func TestGetTopology_RunningContainer_ReturnsTopology(t *testing.T) {
	file := clitest.File(t, `name: project-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)

	session, _ := store.Create("project-1", file)

	app := Application{
		SessionStore: store,
		DockerProvider: dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1234567890",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-1",
					},
					State: "running",
				},
			},
		),
	}

	topology, err := app.GetTopology(context.Background(), string(session.ID))

	assert.NoError(t, err)

	assert.Len(t, topology.Nodes, 1)

	assert.Equal(t, "project-1", topology.Project)
	assert.Equal(t, "chaosd-web-1", topology.Nodes[0].ContainerName)
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	file := clitest.File(t, `services:
  web:
    ports: [
`)

	app := Application{
		SessionStore:   sessiontest.CreateStubStore(t),
		DockerProvider: dockertest.EmptyDockerProvider(),
	}

	sessionID, err := app.Load(context.Background(), file)

	assert.Error(t, err)

	assert.Empty(t, sessionID)

	assert.Contains(t, err.Error(), "failed to parse file")
}

func TestLoad_ValidComposeFile_CreatesSession(t *testing.T) {
	file := clitest.File(t, `name: project-load-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)
	app := Application{
		SessionStore:   store,
		DockerProvider: dockertest.EmptyDockerProvider(),
	}

	sessionID, err := app.Load(context.Background(), file)

	assert.NoError(t, err)

	assert.NotEmpty(t, sessionID)

	createdSession, err := store.Get(sessionID)

	assert.NoError(t, err)

	assert.Equal(t, "project-load-1", createdSession.Project)
	assert.Equal(t, file, createdSession.ComposeFile)
}

func TestRestartService_UnknownService_ReturnsError(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)

	session, _ := store.Create("project-restart-1", file)

	app := Application{
		SessionStore:   store,
		DockerProvider: dockertest.EmptyDockerProvider(),
	}

	results, err := app.RestartService(context.Background(), string(session.ID), "api")

	assert.Error(t, err)

	assert.Empty(t, results)

	assert.EqualError(t, err, "service api not found in project project-restart-1")
}

func TestRestartService_RunningContainers_ReturnsResults(t *testing.T) {
	file := clitest.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

	store := sessiontest.CreateStubStore(t)

	session, _ := store.Create("project-restart-1", file)

	app := Application{
		SessionStore: store,
		DockerProvider: dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-restart-1",
					},
					State: "running",
				},
			},
		),
	}

	results, err := app.RestartService(context.Background(), string(session.ID), "web")

	assert.NoError(t, err)

	assert.Len(t, results, 1)

	assert.Equal(t, "chaosd-web-1", results[0].Node.ContainerName)

	assert.NoError(t, results[0].Err)
}

func TestPartition_RunningNodes_PartitionsNodes(t *testing.T) {
	app, sessionID := createNetworkApplication(t)

	err := app.Partition(context.Background(), sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.NoError(t, err)
}

func TestHeal_RunningNodes_HealsNodes(t *testing.T) {
	app, sessionID := createNetworkApplication(t)

	err := app.Partition(context.Background(), sessionID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), sessionID, "chaosd-web-1", "chaosd-db-1")

	assert.NoError(t, err)
}

func TestGetRunningNode_UnknownNode_ReturnsError(t *testing.T) {
	topology := &topology.Topology{}

	node, err := getRunningNode(topology, "chaosd-web-1")

	assert.Error(t, err)

	assert.Nil(t, node)

	assert.EqualError(t, err, "chaosd-web-1 missing")
}

func TestGetRunningNode_StoppedNode_ReturnsError(t *testing.T) {
	topology := &topology.Topology{
		Nodes: []topology.Node{
			{
				ContainerName: "chaosd-web-1",
				State:         "exited",
			},
		},
	}

	node, err := getRunningNode(topology, "chaosd-web-1")

	assert.Error(t, err)

	assert.Nil(t, node)

	assert.EqualError(t, err, "chaosd-web-1 is not running")
}

func createNetworkApplication(t *testing.T) (*Application, string) {
	t.Helper()

	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	store := sessiontest.CreateStubStore(t)

	session, _ := store.Create("project-network-1", file)

	app := &Application{
		SessionStore: store,
		DockerProvider: dockertest.FakeDockerProvider(
			[]container.Summary{
				{
					ID:    "1",
					Names: []string{"chaosd-web-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "web",
						"com.docker.compose.project": "project-network-1",
					},
					State: "running",
				},
				{
					ID:    "2",
					Names: []string{"chaosd-db-1"},
					Labels: map[string]string{
						"com.docker.compose.service": "db",
						"com.docker.compose.project": "project-network-1",
					},
					State: "running",
				},
			},
		),
		NetworkManager: networktest.NewStubManager(),
	}

	return app, string(session.ID)
}
