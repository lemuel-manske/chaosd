package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"chaosd/cli/internal/docker"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type dockerProviderMock struct {
	client docker.DockerClient
}

func (m *dockerProviderMock) NewClient() (docker.DockerClient, error) {
	return m.client, nil
}

type dockerClientMock struct {
	err error
}

func (f dockerClientMock) Ping(context.Context) error {
	return f.err
}

func fakeDockerProvider() docker.DockerProvider {
	return &dockerProviderMock{
		client: &dockerClientMock{},
	}
}

func unreachableDockerProvider() docker.DockerProvider {
	return &dockerProviderMock{
		client: &dockerClientMock{
			err: os.ErrNotExist,
		},
	}
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
