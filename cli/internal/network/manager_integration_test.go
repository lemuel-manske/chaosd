//go:build integration

package network_test

import (
	"context"
	"testing"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/topology"

	"chaosd/cli/internal/docker/dockertest"

	"github.com/stretchr/testify/assert"
)

func TestManager_PartitionAndHeal(t *testing.T) {
	app := dockertest.StartComposeApp(t, "project-manager-1", `name: project-manager-1

services:
  node-a:
    image: curlimages/curl
    command: ["sleep", "infinity"]

  node-b:
    image: nginx:alpine
`)

	ctx := context.Background()

	composeFile, err := docker.Parse(app.ComposeFile)
	assert.NoError(t, err)

	dockerProvider := dockertest.NewRealDockerProvider()
	dockerClient, err := dockerProvider.NewClient()
	assert.NoError(t, err)

	tl, err := topology.Load(composeFile, ctx, dockerClient)
	assert.NoError(t, err)

	nodeA := tl.NodesByServiceName("node-a")[0]
	nodeB := tl.NodesByServiceName("node-b")[0]
	assert.NotNil(t, nodeA)
	assert.NotNil(t, nodeB)

	dockertest.AssertCanReach(
		t,
		"project-manager-1",
		"node-a",
		"http://node-b",
	)

	manager := network.NewManager(network.NewLinuxFirewallInjector())
	faultID := "test-fault-id"

	err = manager.Partition(ctx, nodeA, nodeB, faultID)
	assert.NoError(t, err)

	dockertest.AssertCannotReach(
		t,
		"project-manager-1",
		"node-a",
		"http://node-b",
	)

	err = manager.Heal(ctx, nodeA, nodeB, faultID)
	assert.NoError(t, err)

	dockertest.AssertCanReach(
		t,
		"project-manager-1",
		"node-a",
		"http://node-b",
	)
}
