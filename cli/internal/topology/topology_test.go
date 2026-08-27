//go:build !integration

package topology

import (
	"context"
	"net/netip"
	"testing"

	"chaosd/cli/internal/docker"

	"chaosd/cli/internal/docker/dockertest"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"

	"github.com/stretchr/testify/assert"
)

func TestLoadReturnsRunningContainer(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
			},
		},
	)

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	actualTopology, err := Load(
		&composeFile,
		ctx,
		dockerClient,
	)

	expectedTopology := Topology{
		Project: "project1",
		Nodes: []Node{
			{
				Service:       "web",
				ContainerID:   "1234567890",
				ContainerName: "chaosd-app-1",
				State:         "running",
			},
		},
	}

	assert.NoError(t, err)

	assert.Equal(t, expectedTopology, *actualTopology)
}

func TestLoadNormalizesDockerContainerName(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"/project-web-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
			},
		})

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name: "project1",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	actualTopology, err := Load(&composeFile, context.Background(), dockerClient)

	assert.NoError(t, err)

	assert.Equal(t, "project-web-1", actualTopology.Nodes[0].ContainerName)
}

func TestLoadMarksMissingService(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{},
	)

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	actualTopology, err := Load(
		&composeFile,
		ctx,
		dockerClient,
	)

	expectedTopology := Topology{
		Project: "project1",
		Nodes: []Node{
			{
				Service: "web",
				State:   "missing",
			},
		},
	}

	assert.NoError(t, err)

	assert.Equal(t, expectedTopology, *actualTopology)
}

func TestNodeByNameReturnsNodeWithMatchingContainerName(t *testing.T) {
	nodes := []Node{
		{
			ContainerName: "node-a",
		},
		{
			ContainerName: "node-b",
		},
	}

	topology := Topology{
		Project: "project1",
		Nodes:   nodes,
	}

	actual := topology.NodeByName("node-a")

	expected := &nodes[0]

	assert.Equal(t, expected, actual)
}

func TestNodesByNetwork(t *testing.T) {
	nodes := []Node{
		{
			Service:       "web",
			ContainerID:   "1",
			ContainerName: "project-web-1",
			State:         "running",
			Networks: []NetworkEndpoint{
				{
					NetworkName: "chaosd_default",
					IPAddress:   "198.162.10.1",
				},
				{
					NetworkName: "chaosd_other",
					IPAddress:   "198.162.10.2",
				},
			},
		},
		{
			Service:       "web",
			ContainerID:   "2",
			ContainerName: "project-web-2",
			State:         "running",
			Networks: []NetworkEndpoint{
				{
					NetworkName: "chaosd_default",
					IPAddress:   "198.162.20.1",
				},
			},
		},
	}

	topology := Topology{
		Project: "project1",
		Nodes:   nodes,
	}

	actual := topology.Networks()

	expected := map[string][]NetworkNode{
		"chaosd_default": {
			{
				Node:    nodes[0],
				Address: "198.162.10.1",
			},
			{
				Node:    nodes[1],
				Address: "198.162.20.1",
			},
		},
		"chaosd_other": {
			{
				Node:    nodes[0],
				Address: "198.162.10.2",
			},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestGroupByNetworks(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1",
				Names: []string{"project-web-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
				NetworkSettings: &container.NetworkSettingsSummary{
					Networks: map[string]*network.EndpointSettings{
						"chaosd_default": {
							IPAddress: netip.MustParseAddr("198.162.10.1"),
						},
						"chaosd_other": {
							IPAddress: netip.MustParseAddr("198.162.10.2"),
						},
					},
				},
			},
			{
				ID:    "2",
				Names: []string{"project-web-2"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
				NetworkSettings: &container.NetworkSettingsSummary{
					Networks: map[string]*network.EndpointSettings{
						"chaosd_default": {
							IPAddress: netip.MustParseAddr("198.162.20.1"),
						},
					},
				},
			},
		},
	)

	nodes := []Node{
		{
			Service:       "web",
			ContainerID:   "1",
			ContainerName: "project-web-1",
			State:         "running",
			Networks: []NetworkEndpoint{
				{
					NetworkName: "chaosd_default",
					IPAddress:   "198.162.10.1",
				},
				{
					NetworkName: "chaosd_other",
					IPAddress:   "198.162.10.2",
				},
			},
		},
		{
			Service:       "web",
			ContainerID:   "2",
			ContainerName: "project-web-2",
			State:         "running",
			Networks: []NetworkEndpoint{
				{
					NetworkName: "chaosd_default",
					IPAddress:   "198.162.20.1",
				},
			},
		},
	}

	expected := map[string][]NetworkNode{
		"chaosd_default": {
			{
				Node:    nodes[0],
				Address: "198.162.10.1",
			},
			{
				Node:    nodes[1],
				Address: "198.162.20.1",
			},
		},
		"chaosd_other": {
			{
				Node:    nodes[0],
				Address: "198.162.10.2",
			},
		},
	}

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	tl, err := Load(
		&composeFile,
		ctx,
		dockerClient,
	)

	actual := tl.Networks()

	assert.NoError(t, err)

	assert.Len(t, actual, 2)

	assert.ElementsMatch(
		t,
		expected["chaosd_default"],
		actual["chaosd_default"],
	)

	assert.ElementsMatch(
		t,
		expected["chaosd_other"],
		actual["chaosd_other"],
	)
}

func TestSharedNetworkEndpointsReturnsOnlyNetworksSharedByBothNodes(t *testing.T) {
	tl := Topology{
		Nodes: []Node{
			{
				ContainerName: "node-a",
				Networks: []NetworkEndpoint{
					{
						NetworkName: "frontend",
						IPAddress:   "172.20.0.2",
					},
					{
						NetworkName: "backend",
						IPAddress:   "172.21.0.2",
					},
				},
			},
			{
				ContainerName: "node-b",
				Networks: []NetworkEndpoint{
					{
						NetworkName: "backend",
						IPAddress:   "172.21.0.3",
					},
				},
			},
		},
	}

	actual := tl.SharedNetworkEndpoints("node-a", "node-b")

	expected := []NetworkPair{
		{
			NetworkName: "backend",
			NodeA: NetworkNode{
				Node:    tl.Nodes[0],
				Address: "172.21.0.2",
			},
			NodeB: NetworkNode{
				Node:    tl.Nodes[1],
				Address: "172.21.0.3",
			},
		},
	}

	assert.ElementsMatch(t, expected, actual)
}
