package test

import (
	"context"
	"testing"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type ComposeApp struct {
	ComposeFile string
	ProjectName string
}

func writeComposeFile(t *testing.T, composeYAML string) string {
	t.Helper()

	file := File(t, composeYAML)

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
