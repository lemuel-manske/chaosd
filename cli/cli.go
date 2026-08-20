package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

type ComposeFile struct {
	Version  string         `yaml:"version"`
	Services map[string]any `yaml:"services"`
}

var LoadCmd = &cobra.Command{
	Use:   "load",
	Short: "Load a docker-compose file",
	Long:  `Load a docker-compose file to chaosd daemon.`,
	RunE:  runLoadCmd,
	Args:  cobra.ExactArgs(1),
}

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

var RootCmd = &cobra.Command{
	Use:   "chaosd",
	Short: "Chaosd is a laboratory for distributed systems",
	Run:   runChaosd,
}

func runChaosd(cmd *cobra.Command, args []string) {
	print("Chaosd command executed\n")
}

func Init() {
	RootCmd.AddCommand(LoadCmd)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
