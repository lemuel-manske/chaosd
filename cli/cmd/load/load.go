package load

import (
	"context"
	"fmt"
	"io"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/topology"

	"github.com/spf13/cobra"
)

const missingState = "missing"

const reportFormat = "%s -> %s%s\n"

func doLoad(
	compose *docker.ComposeFile,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
) error {
	t, err := topology.LoadTopology(compose, ctx, cli)

	if err != nil {
		return err
	}

	t.Print(stdout)

	return nil
}

func runLoadCmd(
	cmd *cobra.Command,
	args []string,
	dockerProvider docker.DockerProvider,
) error {
	composeFile, err := docker.Parse(args[0])

	if err != nil {
		return err
	}

	cli, err := dockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	stdout := cmd.OutOrStdout()
	ctx := cmd.Context()

	return doLoad(
		composeFile,
		ctx,
		stdout,
		cli,
	)
}
