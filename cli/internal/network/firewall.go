package network

import (
	"context"
	"fmt"
	"os/exec"

	"chaosd/cli/internal/topology"
)

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

type Partitioner interface {
	Partition(ctx context.Context, request PartitionRequest) []ActionResult
	Heal(ctx context.Context, request HealRequest) []ActionResult
}

type LinuxFirewallInjector struct{}

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
