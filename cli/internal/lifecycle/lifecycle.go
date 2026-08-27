package lifecycle

import (
	"context"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/topology"

	"github.com/moby/moby/client"
)

type Lifecycle struct {
	docker docker.DockerClient
}

type ActionResult struct {
	Node topology.Node
	Err  error
}

func NewLifecycle(docker docker.DockerClient) *Lifecycle {
	return &Lifecycle{
		docker: docker,
	}
}

func (l *Lifecycle) Restart(ctx context.Context, nodes []topology.Node) []ActionResult {
	results := make([]ActionResult, 0, len(nodes))

	// best-effort restart: we attempt to restart all nodes, even if some fail
	for _, node := range nodes {
		opts := client.ContainerRestartOptions{}

		_, err := l.docker.RestartContainer(
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
