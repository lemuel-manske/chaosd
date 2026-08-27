package topology

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"chaosd/cli/internal/docker"

	"github.com/moby/moby/client"
)

type NetworkNode struct {
	Node    Node
	Address string
}

type NetworkEndpoint struct {
	NetworkName string
	IPAddress   string
}

type NetworkPair struct {
	NetworkName string
	NodeA       NetworkNode
	NodeB       NetworkNode
}

type Node struct {
	Service       string
	ContainerID   string
	ContainerName string
	State         string
	Networks      []NetworkEndpoint
}

type Topology struct {
	Project string
	Nodes   []Node
}

func Load(
	compose *docker.ComposeFile,
	ctx context.Context,
	cli docker.DockerClient,
) (*Topology, error) {
	err := cli.Ping(ctx)

	pName := compose.Name

	if err != nil {
		return nil, fmt.Errorf("failed to ping docker daemon: %v", err)
	}

	var topology = Topology{
		Project: pName,
		Nodes:   []Node{},
	}

	for service := range compose.Services {
		filters := client.Filters{}

		filters.Add(
			"label",
			fmt.Sprintf("com.docker.compose.service=%s", service),
			fmt.Sprintf("com.docker.compose.project=%s", pName),
		)

		containers, err := cli.ContainerList(ctx, client.ContainerListOptions{
			All:     true,
			Filters: filters,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to list containers: %v", err)
		}

		if len(containers) == 0 {
			topology.Nodes = append(topology.Nodes, Node{
				Service: service,
				State:   "missing",
			})
			continue
		}

		for _, container := range containers {
			names := container.Names

			if len(names) == 0 {
				// containers always have names
				// if the container does not have one, panic!
				// at least by this moment

				panic(fmt.Sprintf("container %s has no names", container.ID))
			}

			// docker's API may prefix containers names with a slash,
			// so we need to trim it
			containerName := strings.TrimPrefix(names[0], "/")

			topology.Nodes = append(topology.Nodes, Node{
				Service:       service,
				ContainerID:   container.ID,
				ContainerName: containerName,
				State:         string(container.State),
				Networks: func() []NetworkEndpoint {
					var endpoints []NetworkEndpoint

					netSettings := container.NetworkSettings

					if netSettings == nil {
						return endpoints
					}

					networks := netSettings.Networks

					if networks == nil {
						return endpoints
					}

					for networkName, network := range networks {
						endpoints = append(endpoints, NetworkEndpoint{
							NetworkName: networkName,
							IPAddress:   network.IPAddress.String(),
						})
					}

					return endpoints
				}(),
			})
		}
	}

	return &topology, nil
}

func (t *Topology) NodesByServiceName(serviceName string) []Node {
	var nodes []Node

	for _, node := range t.Nodes {
		if node.Service == serviceName {
			nodes = append(nodes, node)
		}
	}

	return nodes
}

func (t *Topology) NodeByName(nodeName string) *Node {
	for _, node := range t.Nodes {
		if node.ContainerName == nodeName {
			return &node
		}
	}

	return nil
}

func (t Topology) Networks() map[string][]NetworkNode {
	nodesByNetwork := make(map[string][]NetworkNode)

	for _, node := range t.Nodes {

		// keep order of networks consistent
		sort.Slice(node.Networks, func(i, j int) bool {
			return node.Networks[i].NetworkName < node.Networks[j].NetworkName
		})

		for _, network := range node.Networks {
			nodesByNetwork[network.NetworkName] = append(
				nodesByNetwork[network.NetworkName],
				NetworkNode{
					Node:    node,
					Address: network.IPAddress,
				},
			)
		}
	}

	return nodesByNetwork
}

func (t *Topology) SharedNetworkEndpoints(nodeA string, nodeB string) []NetworkPair {
	var sharedNetworks []NetworkPair

	nodeAObj := t.NodeByName(nodeA)
	nodeBObj := t.NodeByName(nodeB)

	if nodeAObj == nil || nodeBObj == nil {
		return sharedNetworks
	}

	for _, networkA := range nodeAObj.Networks {
		for _, networkB := range nodeBObj.Networks {
			if networkA.NetworkName == networkB.NetworkName {
				sharedNetworks = append(sharedNetworks, NetworkPair{
					NetworkName: networkA.NetworkName,
					NodeA: NetworkNode{
						Node:    *nodeAObj,
						Address: networkA.IPAddress,
					},
					NodeB: NetworkNode{
						Node:    *nodeBObj,
						Address: networkB.IPAddress,
					},
				})
			}
		}
	}

	return sharedNetworks
}

func (t *Topology) Print(stdout io.Writer) {
	reportFormat := "%-20s %-30s %-10s\n"
	missingState := "missing"

	fmt.Fprintf(stdout, reportFormat, "Service", "Container Name", "State")
	fmt.Fprintf(stdout, reportFormat, "-------", "--------------", "-----")

	for _, node := range t.Nodes {
		fmt.Fprintf(stdout, reportFormat, node.Service, node.ContainerName, node.State)
	}

	if len(t.Nodes) == 0 {
		fmt.Fprintf(stdout, reportFormat, missingState, "", "")
	}
}
