//go:build integration

package partition

import (
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestPartitionCmd_RunningNodes_BlocksCommunicationBetweenThem(t *testing.T) {
	app := dockertest.StartComposeApp(t, "project-partition-1", `name: project-partition-1

services:
  node-a:
    image: curlimages/curl
    command: ["sleep", "infinity"]

  node-b:
    image: nginx:alpine
`)

	store := sessiontest.CreateStubStore(t)

	session, _ := store.Create("project-partition-1", app.ComposeFile)

	dockertest.AssertCanReach(
		t,
		"project-partition-1",
		"node-a",
		"http://node-b",
	)

	output, err := runPartition(
		t,
		store,
		string(session.ID),
		"project-partition-1-node-a-1",
		"project-partition-1-node-b-1",
	)

	assert.NoError(t, err)

	assert.Contains(t, output, "partitioned")

	dockertest.AssertCannotReach(
		t,
		"project-partition-1",
		"node-a",
		"http://node-b",
	)

	output, err = runHeal(
		t,
		store,
		string(session.ID),
		"project-partition-1-node-a-1",
		"project-partition-1-node-b-1",
	)

	assert.NoError(t, err)

	assert.Contains(t, output, "healed")

	dockertest.AssertCanReach(
		t,
		"project-partition-1",
		"node-a",
		"http://node-b",
	)
}

func runPartition(
	t *testing.T,
	sessionStore session.Store,
	sessionID string,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	networkManager := networktest.NewRealManager()

	cmd := NewPartitionCmd(sessionStore, dockertest.RealDockerProvider(), networkManager)

	return clitest.ExecuteCommand(t, cmd, sessionID, nodeA, nodeB)
}

func runHeal(
	t *testing.T,
	sessionStore session.Store,
	sessionID string,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	networkManager := networktest.NewRealManager()

	cmd := NewHealCmd(sessionStore, dockertest.RealDockerProvider(), networkManager)

	return clitest.ExecuteCommand(t, cmd, sessionID, nodeA, nodeB)
}
