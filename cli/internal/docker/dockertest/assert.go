package dockertest

import (
	"context"
	"io"
	"testing"

	"github.com/testcontainers/testcontainers-go"

	"github.com/stretchr/testify/require"
)

func AssertReachable(
	t *testing.T,
	container testcontainers.Container,
	target string,
) {
	t.Helper()

	ctx := context.Background()

	exitCode, reader, err := container.Exec(
		ctx,
		[]string{
			"ping",
			"-c", "1",
			"-W", "2",
			target,
		},
	)
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	require.Equalf(
		t,
		0,
		exitCode,
		"expected %s to be reachable, output: %s",
		target,
		string(output),
	)
}

func AssertNotReachable(
	t *testing.T,
	container testcontainers.Container,
	target string,
) {
	t.Helper()

	ctx := context.Background()

	exitCode, reader, err := container.Exec(
		ctx,
		[]string{
			"ping",
			"-c", "1",
			"-W", "2",
			target,
		},
	)
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	require.NotEqualf(
		t,
		0,
		exitCode,
		"expected %s to not be reachable, output: %s",
		target,
		string(output),
	)
}
