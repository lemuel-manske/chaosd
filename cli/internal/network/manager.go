package network

import (
	"context"

	"chaosd/cli/internal/topology"
)

type Manager interface {
	Partition(ctx context.Context, a topology.Node, b topology.Node) error
	Heal(ctx context.Context, a topology.Node, b topology.Node) error
}

type concreteManager struct {
	injector Injector
}

func NewManager(injector Injector) Manager {
	return &concreteManager{
		injector: injector,
	}
}

func (m *concreteManager) Partition(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
) error {
	links := LinksBetween(a, b)

	results := m.injector.Partition(ctx, links)

	for _, r := range results {
		if r.Result != nil {
			return r.Result
		}
	}

	return nil
}

func (m *concreteManager) Heal(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
) error {
	links := LinksBetween(a, b)

	results := m.injector.Heal(ctx, links)

	for _, r := range results {
		if r.Result != nil {
			return r.Result
		}
	}

	return nil
}
