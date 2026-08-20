package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func runLoadCmd(cmd *cobra.Command, args []string) error {
	file := args[0]

	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", file)
		}

		return err
	}

	yamlFile, err := os.ReadFile(file)

	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", file, err)
	}

	var compose ComposeFile

	if err := yaml.Unmarshal(yamlFile, &compose); err != nil {
		return fmt.Errorf("failed to parse file %s", file)
	}

	if len(compose.Services) == 0 {
		return fmt.Errorf("no services defined in file %s", file)
	}

	for service := range compose.Services {
		fmt.Fprintln(cmd.OutOrStdout(), service)
	}

	return nil
}
