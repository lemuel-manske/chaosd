package clitest

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func ExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()

	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)

	err := cmd.Execute()

	return output.String(), err
}

func File(t *testing.T, content string) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), "stub.yaml")

	err := os.WriteFile(file, []byte(content), 0o600)
	require.NoError(t, err)

	return file
}
