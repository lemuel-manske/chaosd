//go:build integration

package network_test

import (
	"context"
	"testing"
	"time"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/topology"

	"chaosd/cli/internal/docker/dockertest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetemInjector_DockerBridgeDelayAndHeal(t *testing.T) {
	app := dockertest.StartCompose(t, "project-netem-1", `name: project-netem-1
services:
  node-a:
    image: nginx:alpine

  node-b:
    image: nginx:alpine
`)

	ctx := context.Background()

	dockerProvider := dockertest.NewRealDockerProvider()
	dockerClient, err := dockerProvider.NewClient()
	require.NoError(t, err)

	composeFile, err := docker.Parse(app.ComposeFile)
	require.NoError(t, err)

	tl, err := topology.Load(composeFile, ctx, dockerClient)
	require.NoError(t, err)

	nodeA := tl.NodesByServiceName("node-a")[0]
	nodeB := tl.NodesByServiceName("node-b")[0]
	assert.NotNil(t, nodeA)
	assert.NotNil(t, nodeB)

	containerA := dockertest.ContainerByServiceName(t, "project-netem-1", "node-a")

	dockertest.AssertReachable(
		t,
		containerA,
		"node-b",
	)

	injector := network.NewNetemInjector()

	faultID := "test-fault-id"

	req := network.NewDelayRequest(nodeA, nodeB, faultID, 200*time.Millisecond)
	results := injector.Delay(ctx, req)

	require.NoError(t, results[0].Err)

	afterDelay := dockertest.MeasurePingFrom(
		t,
		containerA,
		"node-b",
	)

	assert.GreaterOrEqual(
		t,
		afterDelay,
		150*time.Millisecond,
	)

	healReq := network.NewHealRequest(nodeA, nodeB, faultID)
	results = injector.Heal(ctx, healReq)

	require.NoError(t, results[0].Err)

	afterHeal := dockertest.MeasurePingFrom(
		t,
		containerA,
		"node-b",
	)

	assert.Less(
		t,
		afterHeal,
		afterDelay-150*time.Millisecond,
	)
}
