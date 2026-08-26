package load

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"

	"github.com/spf13/cobra"
)

const missingState = "missing"

const reportFormat = "%s -> %s%s\n"

func doLoad(
	sessionStore *session.Store,
	composeFileAbsPath string,
	compose *docker.ComposeFile,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
) error {
	_, err := topology.LoadTopology(compose, ctx, cli)

	if err != nil {
		return err
	}

	projectName := compose.Name

	s, err := sessionStore.Create(
		projectName,
		composeFileAbsPath,
	)

	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, s.ID)

	return nil
}

func runLoadCmd(
	cmd *cobra.Command,
	args []string,
	sessionStore *session.Store,
	dockerProvider docker.DockerProvider,
) error {
	filePath := args[0]

	composeFile, err := docker.Parse(filePath)

	if err != nil {
		return err
	}

	cli, err := dockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	stdout := cmd.OutOrStdout()
	ctx := cmd.Context()

	composeFileAbsPath, err := filepath.Abs(filePath)

	return doLoad(
		sessionStore,
		composeFileAbsPath,
		composeFile,
		ctx,
		stdout,
		cli,
	)
}
