package restart

import (
	"context"
	"fmt"
	"io"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/lifecycle"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"

	"github.com/spf13/cobra"
)

func doRestart(
	composeFile *docker.ComposeFile,
	serviceName string,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
) error {
	t, err := topology.LoadTopology(composeFile, ctx, cli)

	if err != nil {
		return err
	}

	nodes := t.NodesByService(serviceName)

	if len(nodes) == 0 {
		return fmt.Errorf(
			"service %s not found in project %s",
			serviceName,
			t.Project,
		)
	}

	manager := lifecycle.NewLifecycle(cli)

	results := manager.Restart(ctx, nodes)

	var restartFailed bool

	for _, r := range results {
		if r.Err != nil {
			restartFailed = true

			fmt.Fprintf(
				stdout,
				"%s failed to restart\n",
				r.Node.ContainerName,
			)

			continue
		}

		fmt.Fprintf(
			stdout,
			"%s restarted\n",
			r.Node.ContainerName,
		)
	}

	if restartFailed {
		return fmt.Errorf("one or more containers failed to restart")
	}

	return nil
}

func runRestartCmd(
	cmd *cobra.Command,
	args []string,
	sessionStore *session.Store,
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

	serviceName := args[1]

	return doRestart(
		composeFile,
		serviceName,
		ctx,
		stdout,
		cli,
	)
}
