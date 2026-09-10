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

func TestLinuxFirewallInjector_PartitionAndHeal(t *testing.T) {
	app := dockertest.StartCompose(t, "project-firewall-1", `name: project-firewall-1

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

	nodaA := tl.NodesByServiceName("node-a")[0]
	nodeB := tl.NodesByServiceName("node-b")[0]
	assert.NotNil(t, nodaA)
	assert.NotNil(t, nodeB)

	containerA := dockertest.ContainerByServiceName(t, "project-firewall-1", "node-a")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	linuxFirewallInjector := network.NewLinuxFirewallInjector()

	fakeFaultID := "test-fault-id"

	partitionRequest := network.NewPartitionRequest(nodaA, nodeB, fakeFaultID)
	results := linuxFirewallInjector.Partition(ctx, partitionRequest)

	for _, result := range results {
		assert.NoError(t, result.Err)
	}

	dockertest.AssertNotReachable(
		t,
		containerA,
		"node-b",
	)

	healRequest := network.NewHealRequest(nodaA, nodeB, fakeFaultID)
	results = linuxFirewallInjector.Heal(ctx, healRequest)

	for _, result := range results {
		assert.NoError(t, result.Err)
	}

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)
}
