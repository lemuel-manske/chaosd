package dockertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertCannotReach(
	t *testing.T,
	project string,
	fromService string,
	target string,
) {
	t.Helper()

	ctr := ContainerByServiceName(t, project, fromService)

	exitCode, output, err := ctr.Exec(
		context.Background(),
		[]string{
			"curl",
			"--silent",
			"--show-error",
			"--fail",
			"--max-time",
			"2",
			target,
		},
	)

	require.NoError(t, err)
	require.NotEqualf(
		t,
		0,
		exitCode,
		"expected %s to not reach %s, output: %s",
		fromService,
		target,
		output,
	)
}

func AssertCanReach(
	t *testing.T,
	project string,
	fromService string,
	target string,
) {
	t.Helper()

	ctr := ContainerByServiceName(t, project, fromService)

	exitCode, output, err := ctr.Exec(
		context.Background(),
		[]string{
			"curl",
			"--silent",
			"--show-error",
			"--fail",
			"--max-time",
			"2",
			target,
		},
	)

	require.NoError(t, err)
	require.Equalf(
		t,
		0,
		exitCode,
		"expected %s to reach %s, output: %s",
		fromService,
		target,
		output,
	)
}
