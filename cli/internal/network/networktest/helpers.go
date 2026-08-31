package networktest

import (
	"context"

	"chaosd/cli/internal/network"
)

type StubInjector struct{}

func NewRealManager() network.Manager {
	return network.NewManager(
		network.NewLinuxFirewallInjector(),
	)
}

func NewStubManager() network.Manager {
	return network.NewManager(
		NewStubInjector(),
	)
}

func NewStubInjector() *StubInjector {
	return &StubInjector{}
}

func (i *StubInjector) Partition(
	ctx context.Context,
	request network.PartitionRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range request.Links {
		results = append(results, network.ActionResult{
			Link:   l,
			Err: nil,
		})
	}

	return results
}

func (i *StubInjector) Heal(
	ctx context.Context,
	request network.HealRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range request.Links {
		results = append(results, network.ActionResult{
			Link:   l,
			Err: nil,
		})
	}

	return results
}
