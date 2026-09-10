package network

import (
	"context"
	"fmt"
	"time"

	"chaosd/cli/internal/topology"
)

type NetworkFaultType string

const (
	FaultTypePartition NetworkFaultType = "partition"
	FaultTypeDelay     NetworkFaultType = "delay"
)

type AppliedEffect struct {
	NetworkName string
	SourceIP    string
	TargetIP    string
}

type Manager interface {
	Delay(
		ctx context.Context,
		a topology.Node,
		b topology.Node,
		faultID string,
		delay time.Duration,
	) ([]AppliedEffect, error)

	Partition(
		ctx context.Context,
		a topology.Node,
		b topology.Node,
		faultID string,
	) ([]AppliedEffect, error)

	Heal(
		ctx context.Context,
		faultID string,
		faultType NetworkFaultType,
		effects []AppliedEffect,
	) error
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
) ([]AppliedEffect, error) {
	request := NewPartitionRequest(a, b, faultID)

	effects := []AppliedEffect{}

	if len(request.Links) == 0 {
		return effects, fmt.Errorf("no shared network found between %s and %s", a.ContainerName, b.ContainerName)
	}

	results := m.partitioner.Partition(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return effects, r.Err
		}

		e := AppliedEffect{
			NetworkName: r.Link.NetworkName,
			SourceIP:    r.Link.SourceIP,
			TargetIP:    r.Link.TargetIP,
		}

		effects = append(effects, e)
	}

	return effects, nil
}

func (m *concreteManager) Delay(
	ctx context.Context,
	a topology.Node,
	b topology.Node,
	faultID string,
	delay time.Duration,
) ([]AppliedEffect, error) {
	request := NewDelayRequest(a, b, faultID, delay)

	effects := []AppliedEffect{}

	if len(request.Links) == 0 {
		return effects, fmt.Errorf("no shared network found between %s and %s", a.ContainerName, b.ContainerName)
	}

	results := m.delayer.Delay(ctx, request)

	for _, r := range results {
		if r.Err != nil {
			return effects, r.Err
		}

		e := AppliedEffect{
			NetworkName: r.Link.NetworkName,
			SourceIP:    r.Link.SourceIP,
			TargetIP:    r.Link.TargetIP,
		}

		effects = append(effects, e)
	}

	return effects, nil
}

func (m *concreteManager) Heal(
	ctx context.Context,
	faultID string,
	faultType NetworkFaultType,
	effects []AppliedEffect,
) error {
	links := make([]Link, 0, len(effects))

	for _, effect := range effects {
		links = append(links, Link{
			NetworkName: effect.NetworkName,
			SourceIP:    effect.SourceIP,
			TargetIP:    effect.TargetIP,
		})
	}

	request := HealRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: faultID,
		},
	}

	var results []ActionResult

	switch faultType {
	case FaultTypeDelay:
		results = m.delayer.Heal(ctx, request)

	case FaultTypePartition:
		results = m.partitioner.Heal(ctx, request)

	default:
		return fmt.Errorf("unsupported fault kind %q", faultType)
	}

	for _, r := range results {
		if r.Err != nil {
			return r.Err
		}
	}

	return nil
}
