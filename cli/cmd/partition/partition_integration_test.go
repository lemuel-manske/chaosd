//go:build integration

package partition

import (
	"testing"

	"chaosd/cli/application"
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

	sessionStore := sessiontest.NewTmpSessionStore(t)

	session, _ := sessionStore.Create("project-partition-1", app.ComposeFile)

	dockertest.AssertCanReach(
		t,
		"project-partition-1",
		"node-a",
		"http://node-b",
	)

	output, err := runPartition(
		t,
		sessionStore,
		session.ID,
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
		sessionStore,
		session.ID,
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

func TestPartitionCmd_RunningNodes_BlocksCommunicationBetweenThem_Bidirectional(t *testing.T) {
	app := dockertest.StartComposeApp(t, "project-partition-2", `name: project-partition-2

services:
  node-a:
    image: curlimages/curl
    command: ["sleep", "infinity"]
  node-b:
    image: nginx:alpine
`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	session, _ := sessionStore.Create("project-partition-2", app.ComposeFile)

	dockertest.AssertCanReach(
		t,
		"project-partition-2",
		"node-a",
		"http://node-b",
	)

	output, err := runPartition(
		t,
		sessionStore,
		session.ID,
		"project-partition-2-node-a-1",
		"project-partition-2-node-b-1",
	)
	assert.NoError(t, err)

	assert.Contains(t, output, "partitioned")

	dockertest.AssertCannotReach(
		t,
		"project-partition-2",
		"node-a",
		"http://node-b",
	)

	dockertest.AssertCannotReach(
		t,
		"project-partition-2",
		"node-b",
		"http://node-a",
	)
}

func runPartition(
	t *testing.T,
	sessionStore session.SessionStore,
	sessionID session.SessionID,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := NewPartitionCmd(*app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		nodeA,
		nodeB,
	)
}

func runHeal(
	t *testing.T,
	sessionStore session.SessionStore,
	sessionID session.SessionID,
	nodeA string,
	nodeB string,
) (string, error) {
	t.Helper()

	dockerProvider := dockertest.NewRealDockerProvider()
	networkManager := networktest.NewRealManager()

	app := application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)

	cmd := NewHealCmd(*app)

	return clitest.ExecuteCommand(
		t,
		cmd,
		string(sessionID),
		nodeA,
		nodeB,
	)
}
