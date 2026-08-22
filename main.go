package main

import (
	"chaosd/cli/cmd"
	"chaosd/cli/cmd/root"
)

func main() {
	rootCmd := root.NewRootCmd()

	cmd.Init(rootCmd)
	cmd.Execute(rootCmd)
}
