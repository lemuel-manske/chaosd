package application

import (
	"context"
	"errors"
	"testing"

	"chaosd/cli/internal/event"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"
	"chaosd/cli/test"

	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestGetTopology_NonexistentSession_ReturnsError(t *testing.T) {
	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithSessionStore(sessionStore),
	)

	sessionID := session.SessionID("session1")

	topology, err := app.GetTopology(context.Background(), sessionID)

	assert.Nil(t, topology)

	assert.Error(t, err)

	assert.Contains(t, err.Error(), `invalid session id`)
}

func TestGetTopology_RunningContainer_ReturnsTopology(t *testing.T) {
	file := test.File(t, `name: project-1
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

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-1", file)

	topology, err := app.GetTopology(context.Background(), session.ID)

	assert.NoError(t, err)

	assert.Len(t, topology.Nodes, 1)

	assert.Equal(t, "project-1", topology.Project)
	assert.Equal(t, "chaosd-web-1", topology.Nodes[0].ContainerName)
}

func TestLoad_InvalidYAML_ReturnsError(t *testing.T) {
	file := test.File(t, `services:
  web:
    ports: [
`)

	app := NewApplication()

	sessionID, err := app.Load(context.Background(), file)

	assert.Error(t, err)

	assert.Empty(t, sessionID)

	assert.Contains(t, err.Error(), "failed to parse file")
}

func TestLoad_ValidComposeFile_CreatesSession(t *testing.T) {
	file := test.File(t, `name: project-load-1
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
				"project-load-1",
				"web",
				"chaosd:192.168.10.2",
			),
		),
	)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithSessionStore(sessionStore),
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
	file := test.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-restart-1", file)

	results, err := app.RestartService(context.Background(), session.ID, "api")

	assert.Error(t, err)

	assert.Empty(t, results)

	assert.EqualError(t, err, "service api not found in project project-restart-1")
}

func TestRestartService_RunningContainers_ReturnsResults(t *testing.T) {
	file := test.File(t, `name: project-restart-1
services:
  web:
    image: nginx
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-restart-1", file)

	results, err := app.RestartService(context.Background(), session.ID, "web")

	assert.NoError(t, err)

	assert.Len(t, results, 1)

	assert.Equal(t, "chaosd-web-1", results[0].Node.ContainerName)

	assert.NoError(t, results[0].Err)
}

func TestPartition_RunningNodes_PartitionsNodes(t *testing.T) {
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networktest.NewStubManager()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	_, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")

	assert.NoError(t, err)
}

func TestHeal_RunningNodes_HealsNodes(t *testing.T) {
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networktest.NewStubManager()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	faultID, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), session.ID, faultID)
	assert.NoError(t, err)
}

func TestApplyPartition_PersistsPartialEffectsAndReturnsApplyError(t *testing.T) {
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"a:192.168.10.1,b:192.168.20.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"a:192.168.10.2,b:192.168.20.2",
			),
		),
	)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	networkManager := network.NewManager(
		networktest.NewStubPartitionerWithError(1, errors.New("partition error")),
		networktest.NewStubDelayer(),
	)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networkManager),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	faultID, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")

	assert.Error(t, err)
	assert.EqualError(t, err, "partition error")

	sessionAfter, _ := sessionStore.Get(session.ID)

	assert.Len(t, sessionAfter.Faults, 1)

	fault := sessionAfter.Faults[0]

	assert.Equal(t, faultID, fault.ID)
	assert.Equal(t, "chaosd-web-1", fault.NodeA)
	assert.Equal(t, "chaosd-db-1", fault.NodeB)

	// saves 1st Link effect, but fails on 2nd Link

	assert.Len(t, fault.Effects, 1)

	effect := fault.Effects[0]

	// FLAKY: sometimes, it's the 20.1 and 20.1 IPs

	assert.Equal(t, "192.168.10.1", effect.SourceIP)
	assert.Equal(t, "192.168.10.2", effect.TargetIP)
}

func TestPartition_IsBidirectional(t *testing.T) {
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	dockerProvider := dockertest.NewFakeDockerProvider(
		dockertest.NewContainers(
			dockertest.NewRunningContainer(
				"1",
				"chaosd-web-1",
				"project-network-1",
				"web",
				"chaosd:192.168.10.1",
			),
			dockertest.NewRunningContainer(
				"2",
				"chaosd-db-1",
				"project-network-1",
				"db",
				"chaosd:192.168.10.2",
			),
		),
	)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networktest.NewStubManager()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	faultID, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), session.ID, faultID)
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
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networktest.NewStubManager()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	_, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	events, err := app.EventStore.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 1)

	expectedData := event.PartitionAppliedEventData{
		NodeAName: "chaosd-web-1",
		NodeBName: "chaosd-db-1",
	}

	assert.Equal(t, event.PartitionAppliedEvent, events[0].Type)
	assert.Equal(t, expectedData, events[0].Data)
}

func TestHeal_RunningNodes_KeepsEvent(t *testing.T) {
	file := test.File(t, `name: project-network-1
services:
  web:
    image: nginx
  db:
    image: postgres
`)

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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithNetworkManager(networktest.NewStubManager()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	faultID, err := app.Partition(context.Background(), session.ID, "chaosd-web-1", "chaosd-db-1")
	assert.NoError(t, err)

	err = app.Heal(context.Background(), session.ID, faultID)
	assert.NoError(t, err)

	events, err := app.EventStore.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 2)

	expectedData := event.HealAppliedEventData{
		NodeAName: "chaosd-web-1",
		NodeBName: "chaosd-db-1",
	}

	assert.Equal(t, event.HealAppliedEvent, events[1].Type)
	assert.Equal(t, expectedData, events[1].Data)
}

func TestRestartService_RunningContainers_KeepsEvent(t *testing.T) {
	file := test.File(t, `name: project-network-1
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

	app := NewApplication(
		WithDockerProvider(dockerProvider),
		WithEventStore(event.NewInMemoryEventStore()),
		WithSessionStore(sessionStore),
	)

	session, _ := sessionStore.Create("project-network-1", file)

	results, err := app.RestartService(context.Background(), session.ID, "web")

	assert.NoError(t, err)

	assert.Len(t, results, 1)

	events, err := app.EventStore.List(session.ID)
	assert.NoError(t, err)

	assert.Len(t, events, 1)

	expectedData := event.RestartEventData{
		ServiceName: "web",
	}

	assert.Equal(t, event.RestartEvent, events[0].Type)
	assert.Equal(t, expectedData, events[0].Data)
}
