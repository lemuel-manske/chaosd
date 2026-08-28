package network

import (
	"chaosd/cli/internal/topology"
	"context"
	"fmt"
	"os/exec"
)

type RuleMetadata struct {
	FaultID string
}

type PartitionRequest struct {
	Links    []Link
	Metadata RuleMetadata
}

type HealRequest struct {
	Links    []Link
	Metadata RuleMetadata
}

type ActionResult struct {
	Link   Link
	Result error
}

type Injector interface {
	Partition(ctx context.Context, request PartitionRequest) []ActionResult
	Heal(ctx context.Context, request HealRequest) []ActionResult
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

const (
	chaosdChainName = "CHAOSD"
	commentFormat   = "chaosd:%s"
)

func (i *LinuxFirewallInjector) Partition(
	ctx context.Context,
	request PartitionRequest,
) []ActionResult {
	results := make([]ActionResult, 0)

	for _, l := range request.Links {

		targetIP := l.TargetIP
		sourceIP := l.SourceIP

		err := ensureChaosdChain()

		if err != nil {
			results = append(results, ActionResult{
				Link:   l,
				Result: fmt.Errorf("failed to ensure CHAOSD chain exists: %v", err),
			})

			continue
		}

		cmd := exec.CommandContext(
			ctx,
			"iptables",
			"-I", chaosdChainName,
			"-s", sourceIP,
			"-d", targetIP,
			"-m", "comment", "--comment", fmt.Sprintf(commentFormat, request.Metadata.FaultID),
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
	request HealRequest,
) []ActionResult {
	results := make([]ActionResult, 0)

	for _, l := range request.Links {
		targetIP := l.TargetIP
		sourceIP := l.SourceIP

		err := ensureChaosdChain()

		if err != nil {
			results = append(results, ActionResult{
				Link:   l,
				Result: fmt.Errorf("failed to ensure CHAOSD chain exists: %v", err),
			})

			continue
		}

		cmd := exec.CommandContext(
			ctx,
			"iptables",
			"-D", chaosdChainName,
			"-s", sourceIP,
			"-d", targetIP,
			"-m", "comment", "--comment", fmt.Sprintf(commentFormat, request.Metadata.FaultID),
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

// checks if the CHAOSD chain exists in iptables, and creates it if it doesn't.
func ensureChaosdChain() error {
	cmd := exec.Command("iptables", "-L", "CHAOSD")
	err := cmd.Run()

	if err != nil {
		cmd := exec.Command("iptables", "-N", "CHAOSD")
		err = cmd.Run()

		if err != nil {
			return fmt.Errorf("failed to create CHAOSD chain: %v", err)
		}

		cmd = exec.Command("iptables", "-I", "DOCKER-USER", "-j", "CHAOSD")
		err = cmd.Run()

		if err != nil {
			return fmt.Errorf("failed to insert CHAOSD chain into DOCKER-USER: %v", err)
		}
	}

	return nil
}
