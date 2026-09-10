package network

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type InterfaceResolver interface {
	ResolveEgressInterface(
		ctx context.Context,
		link Link,
	) (string, error)
}

type LinuxInterfaceResolver struct{}

func NewLinuxInterfaceResolver() *LinuxInterfaceResolver {
	return &LinuxInterfaceResolver{}
}

func (r *LinuxInterfaceResolver) ResolveEgressInterface(
	ctx context.Context,
	link Link,
) (string, error) {
	targetIP := link.TargetIP

	// Ping the target IP to ensure it is in the neighbor table
	_ = exec.CommandContext(
		ctx,
		"ping",
		"-c",
		"1",
		"-W",
		"1",
		targetIP,
	).Run()

	mac, err := resolveMACFromNeighborTable(ctx, targetIP)
	if err != nil {
		return "", fmt.Errorf(
			"resolve MAC for target %s: %w",
			targetIP,
			err,
		)
	}

	iface, err := resolveInterfaceFromFDB(ctx, mac)
	if err != nil {
		return "", fmt.Errorf(
			"resolve interface for MAC %s: %w",
			mac,
			err,
		)
	}

	return iface, nil
}

func resolveMACFromNeighborTable(
	ctx context.Context,
	targetIP string,
) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"ip",
		"neigh",
		"show",
		targetIP,
	)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ip neigh show: %w", err)
	}

	fields := strings.Fields(string(out))

	for i, field := range fields {
		if field == "lladdr" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}

	return "", fmt.Errorf(
		"MAC address not found in neighbor entry: %q",
		strings.TrimSpace(string(out)),
	)
}

func resolveInterfaceFromFDB(
	ctx context.Context,
	mac string,
) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"bridge",
		"fdb",
		"show",
	)

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("bridge fdb show: %w", err)
	}

	for line := range strings.SplitSeq(string(out), "\n") {
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}

		if !strings.EqualFold(fields[0], mac) {
			continue
		}

		for i, field := range fields {
			if field == "dev" && i+1 < len(fields) {
				return fields[i+1], nil
			}
		}
	}

	return "", fmt.Errorf(
		"interface not found in FDB for MAC %s",
		mac,
	)
}
