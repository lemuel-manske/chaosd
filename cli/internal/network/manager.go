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

	request := PartitionRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: createFaultID(a, b),
		},
	}

	results := m.injector.Partition(ctx, request)

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

	request := HealRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: createFaultID(a, b),
		},
	}

	results := m.injector.Heal(ctx, request)

	for _, r := range results {
		if r.Result != nil {
			return r.Result
		}
	}

	return nil
}

func createFaultID(a topology.Node, b topology.Node) string {
	return a.ContainerName + "-" + b.ContainerName
}
