//go:build !integration

package network

import (
	"testing"

	"chaosd/cli/internal/topology"

	"github.com/stretchr/testify/assert"
)

func TestLinksBetween(t *testing.T) {
	tl := topology.Topology{
		Nodes: []topology.Node{
			{
				ContainerName: "node-a",
				Networks: []topology.NetworkEndpoint{
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
				Networks: []topology.NetworkEndpoint{
					{
						NetworkName: "backend",
						IPAddress:   "172.21.0.3",
					},
				},
			},
		},
	}

	actual := LinksBetween(tl.Nodes[0], tl.Nodes[1])

	expected := []Link{
		{
			NetworkName: "backend",
			SourceIP:    "172.21.0.2",
			TargetIP:    "172.21.0.3",
		},
	}

	assert.ElementsMatch(t, expected, actual)
}

func TestLinksBetween_NoCommonNetworks(t *testing.T) {
	tl := topology.Topology{
		Nodes: []topology.Node{
			{
				ContainerName: "node-a",
				Networks: []topology.NetworkEndpoint{
					{
						NetworkName: "frontend",
						IPAddress:   "172.21.0.3",
					},
				},
			},
			{
				ContainerName: "node-b",
				Networks: []topology.NetworkEndpoint{
					{
						NetworkName: "backend",
						IPAddress:   "172.21.0.2",
					},
				},
			},
		},
	}

	actual := LinksBetween(tl.Nodes[0], tl.Nodes[1])

	expected := []Link{}

	assert.ElementsMatch(t, expected, actual)
}
