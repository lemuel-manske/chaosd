package application

import (
	"context"
	"testing"

	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestGetTopology_NonexistentSession_ReturnsError(t *testing.T) {
	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewEmptyDockerProvider()
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session_id := session.SessionID("session1")

	topology, err := app.GetTopology(context.Background(), session_id)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1234567890",
				"chaosd-web-1",
				"project-1",
				"web",
				"chaosd:198.162.10.1",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-1", file)

	topology, err := app.GetTopology(context.Background(), session.ID)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewEmptyDockerProvider()
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewEmptyDockerProvider()
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	sessionID, err := app.Load(context.Background(), file)

	assert.NoError(t, err)

	assert.NotEmpty(t, sessionID)

	createdSession, err := sessionStore.Get(sessionID)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewEmptyDockerProvider()
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-restart-1", file)

	results, err := app.RestartService(context.Background(), session.ID, "api")

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

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-restart-1",
				"web",
				"chaosd:198.162.10.1",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-restart-1", file)

	results, err := app.RestartService(context.Background(), session.ID, "web")

	assert.NoError(t, err)

	assert.Len(t, results, 1)

	assert.Equal(t, "chaosd-web-1", results[0].Node.ContainerName)

	assert.NoError(t, results[0].Err)
}

func TestPartition_RunningNodes_PartitionsNodes(t *testing.T) {
	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:198.162.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:198.162.10.2",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-network-1", file)

	err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")

	assert.NoError(t, err)
}

func TestHeal_RunningNodes_HealsNodes(t *testing.T) {
	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:198.162.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:198.162.10.2",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-network-1", file)

	err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")

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

func TestPartition_RunningNodes_KeepsEvent(t *testing.T) {
	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:198.162.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:198.162.10.2",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-network-1", file)

	err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	events, err := app.events.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 1)

	expectedData := PartitionAppliedEventData{
		NodeAName: "chaosd-web-1",
		NodeBName: "chaosd-db-1",
	}

	assert.Equal(t, PartitionAppliedEvent, events[0].Type)
	assert.Equal(t, expectedData, events[0].Data)
}

func TestHeal_RunningNodes_KeepsEvent(t *testing.T) {
	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:198.162.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:198.162.10.2",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	session, _ := sessionStore.Create("project-network-1", file)

	err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	events, err := app.events.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 2)

	expectedData := HealAppliedEventData{
		NodeAName: "chaosd-web-1",
		NodeBName: "chaosd-db-1",
	}

	assert.Equal(t, HealAppliedEvent, events[1].Type)
	assert.Equal(t, expectedData, events[1].Data)
}

func TestRestartService_RunningContainers_KeepsEvent(t *testing.T) {
	file := clitest.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:198.162.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:198.162.10.2",
			),
		),
	)
	networkManager := networktest.NewStubManager()

	session, _ := sessionStore.Create("project-network-1", file)

	app := NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	results, err := app.RestartService(context.Background(), session.ID, "web")

	assert.NoError(t, err)

	assert.Len(t, results, 1)

	events, err := app.events.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 1)

	expectedData := RestartEventData{
		ServiceName: "web",
	}

	assert.Equal(t, RestartEvent, events[0].Type)
	assert.Equal(t, expectedData, events[0].Data)
}
