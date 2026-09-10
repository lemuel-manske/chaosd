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
