package load

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

const missingState = "missing"

const reportFormat = "%s -> %s%s\n"

func parseComposeFile(file string) (*docker.ComposeFile, error) {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist", file)
		}

		return nil, err
	}

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", file, err)
	}

	var compose docker.ComposeFile

	if err := yaml.Unmarshal(yamlFile, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse file %s", file)
	}

	if len(compose.Services) == 0 {
		return nil, fmt.Errorf("no services defined in file %s", file)
	}

	if compose.Name == "" {
		absFile, err := filepath.Abs(file)

		if err != nil {
			return nil, fmt.Errorf("failed to resolve file path %s: %v", file, err)
		}

		compose.Name = filepath.Base(filepath.Dir(absFile))
	}

	return &compose, nil
}

func doLoad(
	compose *docker.ComposeFile,
	ctx context.Context,
	stdout io.Writer,
	cli docker.DockerClient,
) error {
	topology, err := loadTopology(compose, ctx, cli)

	if err != nil {
		return err
	}

	printTopology(topology, stdout)

	return nil
}

func runLoadCmd(
	cmd *cobra.Command,
	args []string,
	dockerProvider docker.DockerProvider,
) error {
	composeFile, err := parseComposeFile(args[0])

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
