package dockertest

import (
	"context"
	"os"
	"strings"
	"testing"

	"chaosd/cli/internal/docker"

	"chaosd/cli/clitest"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type DockerProviderMock struct {
	Client docker.DockerClient
}

func (m *DockerProviderMock) NewClient() (docker.DockerClient, error) {
	return m.Client, nil
}

type DockerClientMock struct {
	Containers []container.Summary

	PingErr          error
	ContainerListErr error
	RestartErr       map[string]error
}

func (f DockerClientMock) Ping(context.Context) error {
	return f.PingErr
}

func (f *DockerClientMock) ContainerList(
	ctx context.Context,
	options client.ContainerListOptions,
) ([]container.Summary, error) {
	labelFilters := options.Filters["label"]

	if len(labelFilters) == 0 {
		return f.Containers, f.ContainerListErr
	}

	var filteredContainers []container.Summary

	for _, ctr := range f.Containers {
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

	return filteredContainers, f.ContainerListErr
}

func (f *DockerClientMock) RestartContainer(
	ctx context.Context,
	containerID string,
	options client.ContainerRestartOptions,
) (client.ContainerRestartResult, error) {
	if err, ok := f.RestartErr[containerID]; ok {
		return client.ContainerRestartResult{}, err
	}

	return client.ContainerRestartResult{}, nil
}

func RealDockerProvider() docker.DockerProvider {
	return docker.NewDockerProvider()
}

func EmptyDockerProvider() docker.DockerProvider {
	return &DockerProviderMock{
		Client: &DockerClientMock{
			Containers: []container.Summary{},
		},
	}
}

func FakeDockerProvider(
	containers []container.Summary,
) docker.DockerProvider {
	return &DockerProviderMock{
		Client: &DockerClientMock{
			Containers: containers,
		},
	}
}

func UnreachableDockerProvider() docker.DockerProvider {
	return &DockerProviderMock{
		Client: &DockerClientMock{
			PingErr: os.ErrNotExist,
		},
	}
}

type ComposeApp struct {
	ComposeFile string
	ProjectName string
}

func writeComposeFile(t *testing.T, composeYAML string) string {
	t.Helper()

	file := clitest.File(t, composeYAML)

	return file
}

func containerByServiceName(
	t *testing.T,
	projectName string,
	serviceName string,
) testcontainers.Container {
	t.Helper()

	ctx := context.Background()

	provider, err := testcontainers.NewDockerProvider()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, provider.Close())
	})

	filters := client.Filters{}
	filters.Add("label", "com.docker.compose.project="+projectName)
	filters.Add("label", "com.docker.compose.service="+serviceName)

	result, err := provider.Client().ContainerList(
		ctx,
		client.ContainerListOptions{
			All:     true,
			Filters: filters,
		},
	)
	require.NoError(t, err)
	require.Len(t, result.Items, 1)

	container, err := provider.ContainerFromType(ctx, result.Items[0])
	require.NoError(t, err)

	return container
}

func StartComposeApp(t *testing.T, projectName string, composeYAML string) ComposeApp {
	t.Helper()

	composeFile := writeComposeFile(t, composeYAML)

	stack, err := compose.NewDockerComposeWith(
		compose.WithStackFiles(composeFile),
		compose.StackIdentifier(projectName),
	)

	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, stack.Up(ctx))

	t.Cleanup(func() {
		require.NoError(t, stack.Down(ctx))
	})

	return ComposeApp{
		ComposeFile: composeFile,
		ProjectName: projectName,
	}
}

func StopContainerByServiceName(t *testing.T, projectName string, serviceName string) {
	t.Helper()

	ctx := context.Background()
	c := containerByServiceName(t, projectName, serviceName)

	require.NoError(t, c.Stop(ctx, nil))
}

func RemoveContainerByServiceName(t *testing.T, projectName string, serviceName string) {
	t.Helper()

	ctx := context.Background()
	c := containerByServiceName(t, projectName, serviceName)

	require.NoError(t, c.Terminate(ctx))
}

func ContainerByServiceName(t *testing.T, projectName string, serviceName string) testcontainers.Container {
	t.Helper()

	return containerByServiceName(t, projectName, serviceName)
}

func InspectContainer(t *testing.T, containerID string) client.ContainerInspectResult {
	t.Helper()

	ctx := context.Background()

	provider, err := testcontainers.NewDockerProvider()
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, provider.Close())
	})

	opts := client.ContainerInspectOptions{}

	containerJSON, err := provider.Client().ContainerInspect(ctx, containerID, opts)
	require.NoError(t, err)

	return containerJSON
}
