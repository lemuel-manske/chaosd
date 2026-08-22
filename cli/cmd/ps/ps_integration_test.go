package ps

import (
	"chaosd/cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPsWithRealCompose(t *testing.T) {
	app := test.StartComposeApp(t, `project1`, `
name: project1
services:
  web:
    image: nginx:alpine
`)

	output, err := runPs(t, app.ComposeFile)

	require.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "running")
}

func runPs(t *testing.T, composeFile string) (string, error) {
	t.Helper()

	cmd := NewPsCmd(test.RealDockerProvider())

	return test.ExecuteCommand(t, cmd, composeFile)
}
