package network

import (
	"context"
	"time"

	"chaosd/cli/internal/topology"
)

type DelayRequest struct {
	Links    []Link
	Metadata RuleMetadata
	Delay    time.Duration
}

func NewDelayRequest(
	nodeA topology.Node,
	nodeB topology.Node,
	faultID string,
	delay time.Duration,
) DelayRequest {
	links := LinksBetween(nodeA, nodeB)

	return DelayRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: faultID,
		},
		Delay: delay,
	}
}

type Delayer interface {
	Delay(ctx context.Context, request DelayRequest) []ActionResult
	Heal(ctx context.Context, request HealRequest) []ActionResult
}

type NetemInjector struct{}

func NewNetemInjector() *NetemInjector {
	return &NetemInjector{}
}

func (i *NetemInjector) Delay(
	ctx context.Context,
	request DelayRequest,
) []ActionResult {
	// Implement the delay logic using tc/netem here
	// This is a placeholder for the actual implementation
	results := make([]ActionResult, 0)

	for _, l := range request.Links {
		// Placeholder for applying delay to the link
		results = append(results, ActionResult{
			Link: l,
			Err:  nil, // Replace with actual error if any
		})
	}

	return results
}

func (i *NetemInjector) Heal(
	ctx context.Context,
	request HealRequest,
) []ActionResult {
	// Implement the heal logic to remove delay using tc/netem here
	// This is a placeholder for the actual implementation
	results := make([]ActionResult, 0)

	for _, l := range request.Links {
		// Placeholder for removing delay from the link
		results = append(results, ActionResult{
			Link: l,
			Err:  nil, // Replace with actual error if any
		})
	}

	return results
}
