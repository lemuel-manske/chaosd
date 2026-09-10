package network

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"chaosd/cli/internal/topology"
)

const (
	netemRootHandle = "1:"
	netemParent     = "1:3"
	netemHandle     = "30:"
)

type DelayRequest struct {
	Links    []Link
	Metadata RuleMetadata
	Delay    time.Duration
}

func NewDelayRequest(
	nodeA topology.Node,
	nodeB topology.Node,
	faultID string,
	delay time.Duration,
) DelayRequest {
	return DelayRequest{
		Links: LinksBetween(nodeA, nodeB),
		Metadata: RuleMetadata{
			ID: faultID,
		},
		Delay: delay,
	}
}

type Delayer interface {
	Delay(ctx context.Context, request DelayRequest) []ActionResult
	Heal(ctx context.Context, request HealRequest) []ActionResult
}

type NetemInjector struct{}

func NewNetemInjector() *NetemInjector {
	return &NetemInjector{}
}

// Delay applies network delay to each requested link.
//
// Current limitation:
//
//	Only one netem delay can be active per interface.
//	Multiple target filters can share that delay as long as they use
//	the same interface and delay configuration.
//
// Commands must run in the network namespace where the source interface
// exists.
func (i *NetemInjector) Delay(
	ctx context.Context,
	request DelayRequest,
) []ActionResult {
	results := make([]ActionResult, 0, len(request.Links))

	// Multiple links can use the same interface. We only need to install
	// the qdisc once per interface for this request.
	configuredInterfaces := make(map[string]bool)

	for _, link := range request.Links {
		iface, err := i.resolveInterface(ctx, link.TargetIP)
		if err != nil {
			results = append(results, ActionResult{
				Link: link,
				Err: fmt.Errorf(
					"resolve interface for target %s: %w",
					link.TargetIP,
					err,
				),
			})

			continue
		}

		if !configuredInterfaces[iface] {
			if err := i.ensureRootQdisc(ctx, iface); err != nil {
				results = append(results, ActionResult{
					Link: link,
					Err:  err,
				})

				continue
			}

			if err := i.addDelayQdisc(ctx, iface, request.Delay); err != nil {
				results = append(results, ActionResult{
					Link: link,
					Err:  err,
				})

				continue
			}

			configuredInterfaces[iface] = true
		}

		if err := i.addTargetFilter(
			ctx,
			iface,
			link.TargetIP,
			request.Metadata,
		); err != nil {
			results = append(results, ActionResult{
				Link: link,
				Err:  err,
			})

			continue
		}

		results = append(results, ActionResult{
			Link: link,
		})
	}

	return results
}

func (i *NetemInjector) Heal(
	ctx context.Context,
	request HealRequest,
) []ActionResult {
	results := make([]ActionResult, len(request.Links))

	// Keep track of which result belongs to which interface so that,
	// if qdisc removal fails, we can report it for the affected links.
	type interfaceUsage struct {
		indexes []int
	}

	interfaces := make(map[string]*interfaceUsage)

	for idx, link := range request.Links {
		results[idx] = ActionResult{
			Link: link,
		}

		iface, err := i.resolveInterface(ctx, link.TargetIP)
		if err != nil {
			results[idx].Err = fmt.Errorf(
				"resolve interface for target %s: %w",
				link.TargetIP,
				err,
			)

			continue
		}

		usage := interfaces[iface]

		if usage == nil {
			usage = &interfaceUsage{}
			interfaces[iface] = usage
		}

		usage.indexes = append(usage.indexes, idx)

		err = i.removeTargetFilter(
			ctx,
			iface,
			link.TargetIP,
			request.Metadata,
		)

		if err != nil {
			results[idx].Err = err
		}
	}

	// Filters are removed first. Once none of this fault's links need the
	// netem qdisc anymore, remove it once per interface.
	for iface, usage := range interfaces {
		if err := i.removeDelayQdisc(ctx, iface); err != nil {
			for _, idx := range usage.indexes {
				results[idx].Err = errors.Join(
					results[idx].Err,
					err,
				)
			}
		}
	}

	return results
}

func (i *NetemInjector) resolveInterface(
	ctx context.Context,
	targetIP string,
) (string, error) {
	out, err := exec.CommandContext(
		ctx,
		"ip",
		"route",
		"get",
		targetIP,
	).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf(
			"ip route get %s: %w: %s",
			targetIP,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	fields := strings.Fields(string(out))

	for idx, field := range fields {
		if field == "dev" && idx+1 < len(fields) {
			return fields[idx+1], nil
		}
	}

	return "", fmt.Errorf(
		"interface not found in route to %s",
		targetIP,
	)
}

func (i *NetemInjector) ensureRootQdisc(
	ctx context.Context,
	iface string,
) error {
	// We deliberately use replace here for the first implementation.
	//
	// This means chaosd owns the root qdisc of this interface while the
	// fault is active. Supporting arbitrary pre-existing tc configuration
	// requires a more advanced composition strategy.
	out, err := exec.CommandContext(
		ctx,
		"tc",
		"qdisc",
		"replace",
		"dev", iface,
		"root",
		"handle", netemRootHandle,
		"prio",
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"configure root qdisc on %s: %w: %s",
			iface,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func (i *NetemInjector) addDelayQdisc(
	ctx context.Context,
	iface string,
	delay time.Duration,
) error {
	if delay <= 0 {
		return fmt.Errorf("delay must be greater than zero")
	}

	out, err := exec.CommandContext(
		ctx,
		"tc",
		"qdisc",
		"add",
		"dev", iface,
		"parent", netemParent,
		"handle", netemHandle,
		"netem",
		"delay", formatDelay(delay),
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"add netem delay on %s: %w: %s",
			iface,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func (i *NetemInjector) addTargetFilter(
	ctx context.Context,
	iface string,
	targetIP string,
	metadata RuleMetadata,
) error {
	pref := filterPreference(metadata, targetIP)

	out, err := exec.CommandContext(
		ctx,
		"tc",
		"filter",
		"add",
		"dev", iface,
		"protocol", "ip",
		"parent", "1:0",
		"pref", strconv.Itoa(pref),
		"u32",
		"match", "ip",
		"dst", targetIP+"/32",
		"flowid", netemParent,
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"add traffic filter for %s on %s: %w: %s",
			targetIP,
			iface,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func (i *NetemInjector) removeTargetFilter(
	ctx context.Context,
	iface string,
	targetIP string,
	metadata RuleMetadata,
) error {
	pref := filterPreference(metadata, targetIP)

	out, err := exec.CommandContext(
		ctx,
		"tc",
		"filter",
		"del",
		"dev", iface,
		"protocol", "ip",
		"parent", "1:0",
		"pref", strconv.Itoa(pref),
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"remove traffic filter for %s on %s: %w: %s",
			targetIP,
			iface,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func (i *NetemInjector) removeDelayQdisc(
	ctx context.Context,
	iface string,
) error {
	out, err := exec.CommandContext(
		ctx,
		"tc",
		"qdisc",
		"del",
		"dev", iface,
		"parent", netemParent,
		"handle", netemHandle,
	).CombinedOutput()

	if err != nil {
		return fmt.Errorf(
			"remove netem qdisc on %s: %w: %s",
			iface,
			err,
			strings.TrimSpace(string(out)),
		)
	}

	return nil
}

func filterPreference(
	metadata RuleMetadata,
	targetIP string,
) int {
	hash := fnv.New32a()

	_, _ = hash.Write([]byte(metadata.ID))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(targetIP))

	// Stay away from very small priorities and keep the result within
	// a comfortable uint16 range.
	return 1000 + int(hash.Sum32()%50000)
}

func formatDelay(delay time.Duration) string {
	// Preserve sub-millisecond precision when necessary.
	if delay%time.Millisecond == 0 {
		return fmt.Sprintf("%dms", delay.Milliseconds())
	}

	return fmt.Sprintf("%dus", delay.Microseconds())
}
