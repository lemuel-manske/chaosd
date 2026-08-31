package network_test

import (
	"context"
	"testing"

	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/topology"

	"github.com/stretchr/testify/assert"
)

func TestPartition(t *testing.T) {
	manager := networktest.NewStubManager()

	ctx := context.Background()

	nodeA := topology.Node{
		ContainerName: "node-a",
		Networks: []topology.NetworkEndpoint{
			{
				NetworkName: "frontend",
				IPAddress:   "198.162.10.1",
			},
		},
	}
	nodeB := topology.Node{
		ContainerName: "node-b",
		Networks: []topology.NetworkEndpoint{
			{
				NetworkName: "frontend",
				IPAddress:   "192.168.10.2",
			},
		},
	}

	faultID := "test-fault-id"

	err := manager.Partition(
		ctx,
		nodeA,
		nodeB,
		faultID,
	)

	assert.NoError(t, err)
}
