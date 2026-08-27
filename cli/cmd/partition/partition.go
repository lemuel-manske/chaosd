package partition

import (
	"context"
	"fmt"
	"io"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"

	"github.com/spf13/cobra"
)

func getRunningNode(t *topology.Topology, name string) (*topology.Node, error) {
	node := t.NodeByName(name)
	if node == nil {
		return nil, fmt.Errorf("%s missing", name)
	}

	if node.State != "running" {
		return nil, fmt.Errorf("%s is not running", name)
	}

	return node, nil
}

func doPartition(
	composeFile *docker.ComposeFile,
	nodeAName string,
	nodeBName string,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
	networkManager network.Manager,
) error {
	t, err := topology.Load(composeFile, ctx, cli)

	if err != nil {
		return err
	}

	nodeA, err := getRunningNode(t, nodeAName)

	if err != nil {
		return err
	}

	nodeB, err := getRunningNode(t, nodeBName)

	if err != nil {
		return err
	}

	err = networkManager.Partition(ctx, *nodeA, *nodeB)

	if err != nil {
		return err
	}

	fmt.Fprintf(stdout, "%s and %s partitioned\n", nodeAName, nodeBName)

	return nil
}

func runPartitionCmd(
	cmd *cobra.Command,
	args []string,
	sessionStore session.Store,
	dockerProvider docker.DockerProvider,
	networkManager network.Manager,
) error {
	sessionID := args[0]
	serviceA := args[1]
	serviceB := args[2]

	s, err := sessionStore.Get(sessionID)

	if err != nil {
		return err
	}

	composeFile, err := docker.Parse(s.ComposeFile)

	if err != nil {
		return err
	}

	cli, err := dockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	stdout := cmd.OutOrStdout()
	ctx := cmd.Context()

	return doPartition(
		composeFile,
		serviceA,
		serviceB,
		ctx,
		stdout,
		cli,
		networkManager,
	)
}
