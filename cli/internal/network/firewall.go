package network

import (
	"chaosd/cli/internal/topology"
	"context"
	"fmt"
	"os/exec"
)

type ActionResult struct {
	Link   Link
	Result error
}

type Injector interface {
	Partition(ctx context.Context, links []Link) []ActionResult
	Heal(ctx context.Context, links []Link) []ActionResult
}

type LinuxFirewallInjector struct{}

type Link struct {
	NetworkName string
	SourceIP    string
	TargetIP    string
}

func NewLinuxFirewallInjector() *LinuxFirewallInjector {
	return &LinuxFirewallInjector{}
}

func LinksBetween(a, b topology.Node) []Link {
	var links []Link

	for _, netA := range a.Networks {
		for _, netB := range b.Networks {
			if netA.NetworkName == netB.NetworkName {
				links = append(links, Link{
					NetworkName: netA.NetworkName,
					SourceIP:    netA.IPAddress,
					TargetIP:    netB.IPAddress,
				})
			}
		}
	}

	return links
}

func (i *LinuxFirewallInjector) Partition(
	ctx context.Context,
	links []Link,
) []ActionResult {
	results := make([]ActionResult, 0)

	for _, l := range links {

		targetIP := l.TargetIP
		sourceIP := l.SourceIP

		cmd := exec.CommandContext(
			ctx,
			"iptables",
			"-I", "DOCKER-USER",
			"-s", sourceIP,
			"-d", targetIP,
			"-j", "DROP",
		)

		output, err := cmd.CombinedOutput()

		if err != nil {
			results = append(results, ActionResult{
				Link:   l,
				Result: fmt.Errorf("failed to partition %s -> %s: %v, output: %s", sourceIP, targetIP, err, string(output)),
			})
		} else {
			results = append(results, ActionResult{
				Link:   l,
				Result: nil,
			})
		}
	}

	return results
}

func (i *LinuxFirewallInjector) Heal(
	ctx context.Context,
	links []Link,
) []ActionResult {
	results := make([]ActionResult, 0)

	for _, l := range links {

		targetIP := l.TargetIP
		sourceIP := l.SourceIP

		cmd := exec.CommandContext(
			ctx,
			"iptables",
			"-D", "DOCKER-USER",
			"-s", sourceIP,
			"-d", targetIP,
			"-j", "DROP",
		)

		output, err := cmd.CombinedOutput()

		if err != nil {
			results = append(results, ActionResult{
				Link:   l,
				Result: fmt.Errorf("failed to heal %s -> %s: %v, output: %s", sourceIP, targetIP, err, string(output)),
			})
		} else {
			results = append(results, ActionResult{
				Link:   l,
				Result: nil,
			})
		}
	}

	return results
}
