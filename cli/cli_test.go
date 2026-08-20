package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmdGeneratesNoErrors(t *testing.T) {
	_, err := executeCommand(t, RootCmd)

	assert.NoError(t, err)
}

func TestLoadCmdWithNoArgsThenFail(t *testing.T) {
	output, err := executeCommand(t, LoadCmd)

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestLoadCmdWithMultipleArgsThenFail(t *testing.T) {
	output, err := executeCommand(t, LoadCmd, "file1", "file2")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestLoadCmdWithNonExistentFileThenFail(t *testing.T) {
	output, err := executeCommand(t, LoadCmd, "file1")

	assert.Error(t, err)
	assert.Contains(t, output, "file file1 does not exist")
}

func TestLoadCmdWithInvalidYamlThenFail(t *testing.T) {
	file := stubFile(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := executeCommand(t, LoadCmd, file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestLoadCmdWithValidYamlThenSucceed(t *testing.T) {
	file := stubFile(t, `services:
  web:
    image: nginx
`)

	_, err := executeCommand(t, LoadCmd, file)

	assert.NoError(t, err)
}

func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()

	var output bytes.Buffer

	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs(args)

	err := cmd.Execute()

	return output.String(), err
}

func stubFile(t *testing.T, content string) string {
	t.Helper()

	file := filepath.Join(t.TempDir(), "stub.yaml")

	err := os.WriteFile(file, []byte(content), 0o600)
	require.NoError(t, err)

	return file
}
