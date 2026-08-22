package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"chaosd/cli/internal/docker"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
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
	containers []container.Summary
	err        error
}

func (f dockerClientMock) Ping(context.Context) error {
	return f.err
}

func (f *dockerClientMock) ContainerList(
	ctx context.Context,
	options client.ContainerListOptions,
) ([]container.Summary, error) {
	labelFilters := options.Filters["label"]

	if len(labelFilters) == 0 {
		return f.containers, f.err
	}

	var filteredContainers []container.Summary

	for _, ctr := range f.containers {
		matchesAll := true

		for labelFilter := range labelFilters {
			key, value, hasValue := strings.Cut(labelFilter, "=")

			if hasValue {
				if actualValue, ok := ctr.Labels[key]; !ok || actualValue != value {
					matchesAll = false
					break
				}
			} else {
				if _, ok := ctr.Labels[key]; !ok {
					matchesAll = false
					break
				}
			}
		}

		if matchesAll {
			filteredContainers = append(filteredContainers, ctr)
		}
	}

	return filteredContainers, f.err
}

func realDockerProvider() docker.DockerProvider {
	return docker.NewDockerProvider()
}

func stillDockerProvider() docker.DockerProvider {
	return &dockerProviderMock{
		client: &dockerClientMock{
			containers: []container.Summary{},
		},
	}
}

func fakeDockerProvider(
	containers []container.Summary,
) docker.DockerProvider {
	return &dockerProviderMock{
		client: &dockerClientMock{
			containers: containers,
		},
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
