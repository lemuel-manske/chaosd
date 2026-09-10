//go:build integration

package partition

import (
	"context"
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/event/eventtest"
	"chaosd/cli/internal/session"

	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionCmd_RunningNodes_BlocksCommunicationBetweenThem(t *testing.T) {
	app := dockertest.StartCompose(t, "project-partition-1", `name: project-partition-1

services:
  node-a:
    image: nginx:alpine

  node-b:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	_session, _ := sessionStore.Create("project-partition-1", app.ComposeFile)

	container := dockertest.ContainerByServiceName(t, "project-partition-1", "node-a")

	dockertest.AssertReachable(
		t,
		container,
		"node-b",
	)

	output, err := clitest.RunPartition(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		"project-partition-1-node-a-1",
		"project-partition-1-node-b-1",
	)
	assert.NoError(t, err)

	faultID := session.ParseFaultID(output)

	dockertest.AssertNotReachable(
		t,
		container,
		"node-b",
	)

	output, err = clitest.RunHeal(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		faultID,
	)
	assert.NoError(t, err)

	assert.Contains(t, output, "healed")

	dockertest.AssertReachable(
		t,
		container,
		"node-b",
	)
}

func TestPartitionCmd_RunningNodes_BlocksCommunicationBetweenThem_Bidirectional(t *testing.T) {
	app := dockertest.StartCompose(t, "project-partition-2", `name: project-partition-2

services:
  node-a:
    image: wbitt/network-multitool
  node-b:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	_session, _ := sessionStore.Create("project-partition-2", app.ComposeFile)

	containerA := dockertest.ContainerByServiceName(t, "project-partition-2", "node-a")

	containerB := dockertest.ContainerByServiceName(t, "project-partition-2", "node-b")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	output, err := clitest.RunPartition(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		"project-partition-2-node-a-1",
		"project-partition-2-node-b-1",
	)
	assert.NoError(t, err)

	faultID := session.ParseFaultID(output)

	dockertest.AssertNotReachable(
		t,
		containerA,
		"node-b",
	)

	dockertest.AssertNotReachable(
		t,
		containerB,
		"node-a",
	)

	output, err = clitest.RunHeal(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		faultID,
	)
	assert.NoError(t, err)
	assert.Contains(t, output, "healed")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	dockertest.AssertReachable(
		t,
		containerB,
		"node-a",
	)
}

func TestPartitionCmd_RunningNodes_BlocksCommunicationByIP(t *testing.T) {
	app := dockertest.StartCompose(t, "project-partition-3", `name: project-partition-3

services:
  node-a:
    image: wbitt/network-multitool
  node-b:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)
	eventStore := eventtest.NewTmpEventStore(t)

	_session, _ := sessionStore.Create("project-partition-3", app.ComposeFile)

	containerA := dockertest.ContainerByServiceName(t, "project-partition-3", "node-a")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	output, err := clitest.RunPartition(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		"project-partition-3-node-a-1",
		"project-partition-3-node-b-1",
	)
	assert.NoError(t, err)

	faultID := session.ParseFaultID(output)

	dockertest.AssertNotReachable(
		t,
		containerA,
		"node-b",
	)

	dockertest.StopContainerByServiceName(t, "project-partition-3", "node-a")

	output, err = clitest.RunHeal(
		t,
		sessionStore,
		eventStore,
		_session.ID,
		faultID,
	)
	assert.NoError(t, err)
	assert.Contains(t, output, "healed")

	err = containerA.Start(context.Background())
	require.NoError(t, err)

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)
}
