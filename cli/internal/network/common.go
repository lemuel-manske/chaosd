package network

import (
	"chaosd/cli/internal/topology"
	"sort"
)

type RuleMetadata struct {
	ID string
}

type HealRequest struct {
	Links    []Link
	Metadata RuleMetadata
}

func NewHealRequest(
	nodeA topology.Node,
	nodeB topology.Node,
	faultID string,
) HealRequest {
	links := LinksBetween(nodeA, nodeB)

	return HealRequest{
		Links: links,
		Metadata: RuleMetadata{
			ID: faultID,
		},
	}
}

type ActionResult struct {
	Link Link
	Err  error
}

type Link struct {
	NetworkName string
	SourceIP    string
	TargetIP    string
}

func LinksBetween(a, b topology.Node) []Link {
	var links []Link

	for _, netA := range a.Networks {
		for _, netB := range b.Networks {
			if netA.NetworkName == netB.NetworkName {
				links = append(links, Link{
					NetworkName: netA.NetworkName,
					SourceIP:    netA.IPAddress,
					TargetIP:    netB.IPAddress,
				})
			}
		}
	}

	sort.Slice(links, func(i, j int) bool {
		return links[i].NetworkName < links[j].NetworkName
	})

	return links
}
