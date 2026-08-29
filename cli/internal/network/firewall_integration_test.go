//go:build integration

package network

import (
	"context"
	"testing"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/topology"

	"github.com/stretchr/testify/assert"
)

func TestLinuxFirewallInjector_PartitionAndHeal(t *testing.T) {
	app := dockertest.StartComposeApp(t, "project-firewall-1", `name: project-firewall-1

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

	dockerProvider := dockertest.RealDockerProvider()
	dockerClient, err := dockerProvider.NewClient()
	assert.NoError(t, err)

	tl, err := topology.Load(composeFile, ctx, dockerClient)
	assert.NoError(t, err)

	nodaA := tl.NodesByServiceName("node-a")[0]
	nodeB := tl.NodesByServiceName("node-b")[0]
	assert.NotNil(t, nodaA)
	assert.NotNil(t, nodeB)

	dockertest.AssertCanReach(
		t,
		"project-firewall-1",
		"node-a",
		"http://node-b",
	)

	linuxFirewallInjector := NewLinuxFirewallInjector()

	partitionRequest := PartitionRequest{
		Links: LinksBetween(nodaA, nodeB),
		Metadata: RuleMetadata{
			FaultID: "test-fault-id",
		},
	}
	results := linuxFirewallInjector.Partition(ctx, partitionRequest)

	for _, result := range results {
		assert.NoError(t, result.Result)
	}

	dockertest.AssertCannotReach(
		t,
		"project-firewall-1",
		"node-a",
		"http://node-b",
	)

	healRequest := HealRequest{
		Links: LinksBetween(nodaA, nodeB),
		Metadata: RuleMetadata{
			FaultID: "test-fault-id",
		},
	}
	results = linuxFirewallInjector.Heal(ctx, healRequest)

	for _, result := range results {
		assert.NoError(t, result.Result)
	}

	dockertest.AssertCanReach(
		t,
		"project-firewall-1",
		"node-a",
		"http://node-b",
	)
}
