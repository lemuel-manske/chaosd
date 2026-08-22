package load

import (
	"chaosd/cli/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCmdWithValidYamlAndRealDockerThenSucceed(t *testing.T) {
	file := test.StubFile(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.RealDockerProvider())

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
}
