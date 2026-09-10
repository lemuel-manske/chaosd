package networktest

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TestNetwork struct {
	SourceIP string
	TargetIP string
	Iface    string
	NetNS    string
}

func SetupNetwork(t *testing.T) TestNetwork {
	t.Helper()

	requireNetworkTools(t)

	suffix := strconv.Itoa(os.Getpid() % 10000)

	hostIface := "ch" + suffix + "h"
	peerIface := "ch" + suffix + "p"
	netns := "chaosd-" + suffix

	sourceIP := "10.250.0.1"
	targetIP := "10.250.0.2"

	// defensive cleanup in case the previous test run failed
	_ = exec.Command("ip", "link", "del", hostIface).Run()
	_ = exec.Command("ip", "netns", "del", netns).Run()

	runCommand(t,
		"ip", "netns", "add", netns,
	)

	t.Cleanup(func() {
		_ = exec.Command(
			"ip", "netns", "del", netns,
		).Run()

		_ = exec.Command(
			"ip", "link", "del", hostIface,
		).Run()
	})

	// hostIface <--------> peerIface
	runCommand(t,
		"ip", "link", "add",
		hostIface,
		"type", "veth",
		"peer", "name", peerIface,
	)

	// moves peerIface to the network namespace
	runCommand(t,
		"ip", "link", "set",
		peerIface,
		"netns", netns,
	)

	// configures the host interface
	runCommand(t,
		"ip", "addr", "add",
		sourceIP+"/30",
		"dev", hostIface,
	)

	runCommand(t,
		"ip", "link", "set",
		hostIface,
		"up",
	)

	// configures the peer interface inside the network namespace
	runCommand(t,
		"ip", "netns", "exec",
		netns,
		"ip", "addr", "add",
		targetIP+"/30",
		"dev", peerIface,
	)

	runCommand(t,
		"ip", "netns", "exec",
		netns,
		"ip", "link", "set",
		peerIface,
		"up",
	)

	runCommand(t,
		"ip", "netns", "exec",
		netns,
		"ip", "link", "set",
		"lo",
		"up",
	)

	// proves that the network is working before returning
	runCommand(t,
		"ping",
		"-c", "1",
		"-W", "1",
		targetIP,
	)

	return TestNetwork{
		SourceIP: sourceIP,
		TargetIP: targetIP,
		Iface:    hostIface,
		NetNS:    netns,
	}
}

func requireNetworkTools(t *testing.T) {
	t.Helper()

	if runtime.GOOS != "linux" {
		t.Skip("network integration tests require Linux")
	}

	if os.Geteuid() != 0 {
		t.Skip("network integration tests require root")
	}

	for _, binary := range []string{
		"ip",
		"tc",
		"ping",
	} {
		if _, err := exec.LookPath(binary); err != nil {
			t.Skipf(
				"network integration tests require %s",
				binary,
			)
		}
	}
}

func MeasurePing(t *testing.T, targetIP string) time.Duration {
	t.Helper()

	start := time.Now()

	cmd := exec.Command(
		"ping",
		"-c", "1",
		"-W", "2",
		targetIP,
	)

	out, err := cmd.CombinedOutput()

	require.NoErrorf(
		t,
		err,
		"ping %s failed:\n%s",
		targetIP,
		string(out),
	)

	return time.Since(start)
}

func runCommand(
	t *testing.T,
	name string,
	args ...string,
) string {
	t.Helper()

	cmd := exec.Command(name, args...)

	out, err := cmd.CombinedOutput()

	require.NoErrorf(
		t,
		err,
		"command failed: %s %s\n%s",
		name,
		strings.Join(args, " "),
		string(out),
	)

	return string(out)
}
