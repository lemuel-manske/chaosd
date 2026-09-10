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
	app := dockertest.StartCompose(t, "project-manager-1", `name: project-manager-1

services:
  node-a:
    image: nginx:alpine

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

	containerA := dockertest.ContainerByServiceName(t, "project-manager-1", "node-a")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	manager := network.NewManager(network.NewLinuxFirewallInjector(), network.NewNetemInjector())
	faultID := "test-fault-id"

	effects, err := manager.Partition(ctx, nodeA, nodeB, faultID)
	assert.NoError(t, err)

	dockertest.AssertNotReachable(
		t,
		containerA,
		"node-b",
	)

	err = manager.Heal(ctx, faultID, network.NetworkPartitionFaultType, effects)
	assert.NoError(t, err)

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)
}

func TestManager_PartitionAndHeal_IsBidirectional(t *testing.T) {
	app := dockertest.StartCompose(t, "project-manager-2", `name: project-manager-2

services:
  node-a:
    image: nginx:alpine

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

	containerA := dockertest.ContainerByServiceName(t, "project-manager-2", "node-a")
	containerB := dockertest.ContainerByServiceName(t, "project-manager-2", "node-b")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	manager := network.NewManager(network.NewLinuxFirewallInjector(), network.NewNetemInjector())
	faultID := "test-fault-id"

	effects, err := manager.Partition(ctx, nodeA, nodeB, faultID)
	assert.NoError(t, err)

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

	err = manager.Heal(ctx, faultID, network.NetworkPartitionFaultType, effects)
	assert.NoError(t, err)

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
