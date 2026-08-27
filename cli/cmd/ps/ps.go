package ps

import (
	"context"
	"fmt"
	"io"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"

	"github.com/spf13/cobra"
)

func doPs(
	compose *docker.ComposeFile,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
) error {
	t, err := topology.Load(compose, ctx, cli)

	if err != nil {
		return err
	}

	t.Print(stdout)

	return nil
}

func runPsCmd(
	cmd *cobra.Command,
	args []string,
	sessionStore session.Store,
	dockerProvider docker.DockerProvider,
) error {
	s, err := sessionStore.Get(args[0])

	if err != nil {
		return err
	}

	composeFile, err := docker.Parse(s.ComposeFile)

	if err != nil {
		return err
	}

	cli, err := dockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	stdout := cmd.OutOrStdout()
	ctx := cmd.Context()

	return doPs(
		composeFile,
		ctx,
		stdout,
		cli,
	)
}
