package cli

import (
	"fmt"
	"os"

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

func runLoadCmd(cmd *cobra.Command, args []string) error {
	compose, err := parseComposeFile(args[0])

	if err != nil {
		return err
	}

	for service := range compose.Services {
		fmt.Fprintln(cmd.OutOrStdout(), service)
	}

	return nil
}
