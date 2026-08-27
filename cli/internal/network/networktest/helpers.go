package networktest

import (
	"context"

	"chaosd/cli/internal/network"
	"chaosd/cli/internal/topology"
)

type StubManager struct {
	injector network.Injector
}

func NewRealManager() network.Manager {
	return network.NewManager(
		network.NewLinuxFirewallInjector(),
	)
}

func NewStubManager() network.Manager {
	return &StubManager{
		injector: NewStubInjector(),
	}
}

func (m *StubManager) Partition(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
) error {
	links := network.LinksBetween(a, b)

	results := m.injector.Partition(ctx, links)

	for _, r := range results {
		if r.Result != nil {
			return r.Result
		}
	}

	return nil
}

func (m *StubManager) Heal(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
) error {
	links := network.LinksBetween(a, b)

	results := m.injector.Heal(ctx, links)

	for _, r := range results {
		if r.Result != nil {
			return r.Result
		}
	}

	return nil
}

type StubInjector struct{}

func NewStubInjector() *StubInjector {
	return &StubInjector{}
}

func (i *StubInjector) Partition(ctx context.Context, links []network.Link) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range links {
		results = append(results, network.ActionResult{
			Link:   l,
			Result: nil,
		})
	}

	return results
}

func (i *StubInjector) Heal(ctx context.Context, links []network.Link) []network.ActionResult {
	results := make([]network.ActionResult, 0)

	for _, l := range links {
		results = append(results, network.ActionResult{
			Link:   l,
			Result: nil,
		})
	}

	return results
}
