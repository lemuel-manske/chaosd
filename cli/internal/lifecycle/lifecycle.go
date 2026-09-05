package lifecycle

import (
	"context"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/topology"

	"github.com/moby/moby/client"
)

type Restarter interface {
	Restart(
		ctx context.Context,
		nodes []topology.Node,
	) []ActionResult
}

type ActionResult struct {
	Node topology.Node
	Err  error
}

type DockerRestarter struct {
	dockerClient docker.DockerClient
}

func NewDockerRestarter(dockerClient docker.DockerClient) *DockerRestarter {
	return &DockerRestarter{
		dockerClient: dockerClient,
	}
}

func (r *DockerRestarter) Restart(ctx context.Context, nodes []topology.Node) []ActionResult {
	results := make([]ActionResult, 0, len(nodes))

	// best-effort restart: we attempt to restart all nodes, even if some fail
	for _, node := range nodes {
		opts := client.ContainerRestartOptions{}

		_, err := r.dockerClient.RestartContainer(
			ctx,
			node.ContainerID,
			opts,
		)

		results = append(results, ActionResult{
			Node: node,
			Err:  err,
		})
	}

	return results
}
