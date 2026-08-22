package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
