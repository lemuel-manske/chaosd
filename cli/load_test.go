package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadCmdWithNoArgsThenFail(t *testing.T) {
	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestLoadCmdWithMultipleArgsThenFail(t *testing.T) {
	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), "file1", "file2")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestLoadCmdWithNonExistentFileThenFail(t *testing.T) {
	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), "file1")

	assert.Error(t, err)
	assert.Contains(t, output, "file file1 does not exist")
}

func TestLoadCmdWithInvalidYamlThenFail(t *testing.T) {
	file := stubFile(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestLoadCmdWithValidYamlThenSucceed(t *testing.T) {
	file := stubFile(t, `services:
  web:
    image: nginx
`)

	_, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), file)

	assert.NoError(t, err)
}

func TestLoadCmdWithValidYamlButNoServicesThenFail(t *testing.T) {
	file := stubFile(t, `version: "3.8"`)

	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), file)

	assert.Error(t, err)
	assert.Contains(t, output, "no services defined in file")
}

func TestLoadCmdWithValidYamlThenPrintServiceNames(t *testing.T) {
	file := stubFile(t, `services:
  web:
    image: nginx
  db:
    image: postgres
`)

	output, err := executeCommand(t, NewLoadCmd(fakeDockerProvider()), file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "db")
}

func TestLoadCmdWithValidYamlButUnreachableDockerThenFail(t *testing.T) {
	file := stubFile(t, `services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(unreachableDockerProvider())

	output, err := executeCommand(t, cmd, file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to ping docker daemon")
}
