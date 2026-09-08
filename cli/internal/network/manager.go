package network

import (
	"context"
	"fmt"
	"time"

	"chaosd/cli/internal/topology"
)

type Manager interface {
	Delay(ctx context.Context, a topology.Node, b topology.Node, faultID string, delay time.Duration) error
	Heal(ctx context.Context, a topology.Node, b topology.Node, faultID string) error
	Partition(ctx context.Context, a topology.Node, b topology.Node, faultID string) error
}

type concreteManager struct {
	partitioner Partitioner
	delayer     Delayer
}

func NewManager(
	partitioner Partitioner,
	delayer Delayer,
) Manager {
	return &concreteManager{
		partitioner: partitioner,
		delayer:     delayer,
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

	if len(request.Links) == 0 {
		return fmt.Errorf("no shared network found between %s and %s", a.ContainerName, b.ContainerName)
	}

	results := m.partitioner.Partition(ctx, request)

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

	if len(request.Links) == 0 {
		return fmt.Errorf("no shared network found between %s and %s", a.ContainerName, b.ContainerName)
	}

	results := m.partitioner.Heal(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}

func (m *concreteManager) Delay(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
	faultID string,
	delay time.Duration,
) error {
	request := NewDelayRequest(a, b, faultID, delay)

	if len(request.Links) == 0 {
		return fmt.Errorf("no shared network found between %s and %s", a.ContainerName, b.ContainerName)
	}

	results := m.delayer.Delay(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}
