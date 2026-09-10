package networktest

import (
	"context"

	"chaosd/cli/internal/network"
)

type StubPartitioner struct {
	PartitionErrIdx int
	PartitionErr    error
}

type StubDelayer struct {
}

func NewRealManager() network.Manager {
	return network.NewManager(
		network.NewLinuxFirewallInjector(),
		network.NewNetemInjector(),
	)
}

func NewStubManager() network.Manager {
	return network.NewManager(
		NewStubPartitioner(),
		NewStubDelayer(),
	)
}

func NewStubPartitionerWithError(idx int, err error) *StubPartitioner {
	return &StubPartitioner{
		PartitionErrIdx: idx,
		PartitionErr:    err,
	}
}

func NewStubPartitioner() *StubPartitioner {
	return &StubPartitioner{}
}

func NewStubDelayer() *StubDelayer {
	return &StubDelayer{}
}

func (i *StubPartitioner) Partition(
	ctx context.Context,
	request network.PartitionRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for idx, l := range request.Links {
		if idx == i.PartitionErrIdx {
			results = append(results, network.ActionResult{
				Link: l,
				Err:  i.PartitionErr,
			})

			continue
		}

		results = append(results, network.ActionResult{
			Link: l,
			Err:  nil,
		})
	}

	return results
}

func (i *StubPartitioner) Heal(
	ctx context.Context,
	request network.HealRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range request.Links {
		results = append(results, network.ActionResult{
			Link: l,
			Err:  nil,
		})
	}

	return results
}

func (i *StubDelayer) Delay(
	ctx context.Context,
	request network.DelayRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range request.Links {
		results = append(results, network.ActionResult{
			Link: l,
			Err:  nil,
		})
	}

	return results
}

func (i *StubDelayer) Heal(
	ctx context.Context,
	request network.HealRequest,
) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range request.Links {
		results = append(results, network.ActionResult{
			Link: l,
			Err:  nil,
		})
	}

	return results
}
