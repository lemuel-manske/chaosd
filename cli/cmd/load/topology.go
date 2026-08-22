package load

import (
	"chaosd/cli/internal/docker"
	"context"
	"fmt"
	"io"

	"github.com/moby/moby/client"
)

type Node struct {
	Service       string
	ContainerID   string
	ContainerName string
	State         string
}

type Topology struct {
	Nodes []Node
}

func loadTopology(
	compose *docker.ComposeFile,
	ctx context.Context,
	cli docker.DockerClient,
) (*Topology, error) {
	err := cli.Ping(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to ping docker daemon: %v", err)
	}

	var topology Topology

	for service := range compose.Services {
		filters := client.Filters{}

		serviceNameFilter := fmt.Sprintf("com.docker.compose.service=%s", service)
		projectNameFilter := fmt.Sprintf("com.docker.compose.project=%s", compose.Name)

		filters.Add("label", serviceNameFilter)
		filters.Add("label", projectNameFilter)

		containers, err := cli.ContainerList(ctx, client.ContainerListOptions{
			Filters: filters,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list containers: %v", err)
		}

		if len(containers) == 0 {
			topology.Nodes = append(topology.Nodes, Node{
				Service:       service,
				ContainerName: service,
				State:         "missing",
			})
			continue
		}

		for _, container := range containers {
			names := container.Names

			if len(names) == 0 {
				// containers always have names
				// if the container does not have one, panic! at least at this moment

				panic(fmt.Sprintf("container %s has no names", container.ID))
			}

			containerName := names[0]

			topology.Nodes = append(topology.Nodes, Node{
				Service:       service,
				ContainerID:   container.ID,
				ContainerName: containerName,
				State:         string(container.State),
			})
		}
	}

	return &topology, nil
}

func printTopology(topology *Topology, stdout io.Writer) {
	reportFormat := "%-20s %-30s %-10s\n"
	missingState := "missing"

	fmt.Fprintf(stdout, reportFormat, "Service", "Container Name", "State")
	fmt.Fprintf(stdout, reportFormat, "-------", "--------------", "-----")

	for _, node := range topology.Nodes {
		fmt.Fprintf(stdout, reportFormat, node.Service, node.ContainerName, node.State)
	}

	if len(topology.Nodes) == 0 {
		fmt.Fprintf(stdout, reportFormat, missingState, "", "")
	}
}
