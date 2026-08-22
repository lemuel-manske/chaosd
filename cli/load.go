package cli

import (
	"fmt"
	"os"

	"chaosd/cli/internal/docker"

	"github.com/moby/moby/client"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func parseComposeFile(file string) (*ComposeFile, error) {
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

	var compose ComposeFile

	if err := yaml.Unmarshal(yamlFile, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse file %s", file)
	}

	if len(compose.Services) == 0 {
		return nil, fmt.Errorf("no services defined in file %s", file)
	}

	return &compose, nil
}

func runLoadCmd(cmd *cobra.Command, args []string, dockerProvider docker.DockerProvider) error {
	compose, err := parseComposeFile(args[0])

	if err != nil {
		return err
	}

	cli, err := dockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	err = cli.Ping(cmd.Context())

	if err != nil {
		return fmt.Errorf("failed to ping docker daemon: %v", err)
	}

	for service := range compose.Services {
		filters := client.Filters{}

		labelFilter := fmt.Sprintf("com.docker.compose.service=%s", service)
		filters.Add("label", labelFilter)

		containers, err := cli.ContainerList(cmd.Context(), client.ContainerListOptions{
			Filters: filters,
		})

		if err != nil {
			return fmt.Errorf("failed to list containers: %v", err)
		}

		if len(containers) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s -> missing\n", service)
			continue
		}

		containerName := containers[0].Names[0]

		fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s running\n", service, containerName)
	}

	return nil
}
