package network

import (
	"context"

	"chaosd/cli/internal/topology"
)

type Manager interface {
	Partition(ctx context.Context, a topology.Node, b topology.Node, faultID string) error
	Heal(ctx context.Context, a topology.Node, b topology.Node, faultID string) error
}

type concreteManager struct {
	injector Injector
}

func NewManager(injector Injector) Manager {
	return &concreteManager{
		injector: injector,
	}
}

// Partition isolates node A from node B in both directions
// across every network they share. It's bidirectional.
func (m *concreteManager) Partition(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
	faultID string,
) error {
	request := NewPartitionRequest(a, b, faultID)

	results := m.injector.Partition(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}

func (m *concreteManager) Heal(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
	faultID string,
) error {
	request := NewHealRequest(a, b, faultID)

	results := m.injector.Heal(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}
