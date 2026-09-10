//go:build integration

package network

import (
	"context"
	"net"
	"testing"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/topology"

	"chaosd/cli/internal/docker/dockertest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinuxInterfaceResolver_DockerBridge_ReturnsTargetVeth(t *testing.T) {
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
	assert.NoError(t, err)

	composeFile, err := docker.Parse(app.ComposeFile)
	assert.NoError(t, err)

	tl, err := topology.Load(composeFile, ctx, dockerClient)
	assert.NoError(t, err)

	a := tl.NodeByContainerName("project-netem-1-node-a-1")
	b := tl.NodeByContainerName("project-netem-1-node-b-1")

	links := LinksBetween(*a, *b)
	ifaceResolver := NewLinuxInterfaceResolver()

	actualIface, err := ifaceResolver.ResolveEgressInterface(ctx, links[0])
	require.NoError(t, err)

	_, err = net.InterfaceByName(actualIface)
	require.NoError(t, err)

	assert.Regexp(t, `^veth`, actualIface)
}
