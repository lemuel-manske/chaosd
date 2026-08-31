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

func NewPartitionRequest(
	nodeA topology.Node,
	nodeB topology.Node,
	faultID string,
) PartitionRequest {
	links := LinksBetween(nodeA, nodeB)

	return PartitionRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: faultID,
		},
	}
}

type HealRequest struct {
	Links    []Link
	Metadata RuleMetadata
}

func NewHealRequest(
	nodeA topology.Node,
	nodeB topology.Node,
	faultID string,
) HealRequest {
	links := LinksBetween(nodeA, nodeB)

	return HealRequest{
		Links: links,
		Metadata: RuleMetadata{
			FaultID: faultID,
		},
	}
}

type ActionResult struct {
	Link Link
	Err  error
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

const (
	chaosdChainName = "CHAOSD"
	dockerChainName = "DOCKER-USER"

	commentFormat = "chaosd:%s"
)

func (i *LinuxFirewallInjector) Partition(
	ctx context.Context,
	request PartitionRequest,
) []ActionResult {
	results := make([]ActionResult, 0)

	for _, l := range request.Links {

		targetIP := l.TargetIP
		sourceIP := l.SourceIP

		err := i.ensureCHAOSDChain()

		if err != nil {
			results = append(results, ActionResult{
				Link: l,
				Err:  fmt.Errorf("failed to ensure CHAOSD firewall setup: %v", err),
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
				Link: l,
				Err:  fmt.Errorf("failed to partition %s -> %s: %v, output: %s", sourceIP, targetIP, err, string(output)),
			})
		} else {
			results = append(results, ActionResult{
				Link: l,
				Err:  nil,
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

		err := i.ensureCHAOSDChain()

		if err != nil {
			results = append(results, ActionResult{
				Link: l,
				Err:  fmt.Errorf("failed to ensure CHAOSD firewall setup: %v", err),
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
				Link: l,
				Err:  fmt.Errorf("failed to heal %s -> %s: %v, output: %s", sourceIP, targetIP, err, string(output)),
			})
		} else {
			results = append(results, ActionResult{
				Link: l,
				Err:  nil,
			})
		}
	}

	return results
}

func (i *LinuxFirewallInjector) ensureCHAOSDChain() error {
	if err := i.ensureChain(); err != nil {
		return err
	}

	if err := i.ensureJump(); err != nil {
		return err
	}

	return nil
}

func (i *LinuxFirewallInjector) ensureChain() error {
	cmd := exec.Command("iptables", "-L", chaosdChainName)
	if err := cmd.Run(); err == nil {
		return nil
	}

	cmd = exec.Command("iptables", "-N", chaosdChainName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create CHAOSD chain: %v", err)
	}

	return nil
}

func (i *LinuxFirewallInjector) ensureJump() error {
	cmd := exec.Command(
		"iptables",
		"-C", dockerChainName,
		"-j", chaosdChainName,
	)
	if err := cmd.Run(); err == nil {
		return nil
	}

	cmd = exec.Command(
		"iptables",
		"-I", dockerChainName,
		"-j", chaosdChainName,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"failed to insert %s chain into %s: %v",
			chaosdChainName,
			dockerChainName,
			err,
		)
	}

	return nil
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
