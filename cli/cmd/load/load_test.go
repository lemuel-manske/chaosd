package load

import (
	"testing"

	"chaosd/cli/test"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestLoadCmdWithNoArgsThenFail(t *testing.T) {
	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()))

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestLoadCmdWithMultipleArgsThenFail(t *testing.T) {
	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), "file1", "file2")

	assert.Error(t, err)
	assert.Contains(t, output, "accepts 1 arg(s), received 2")
}

func TestLoadCmdWithNonExistentFileThenFail(t *testing.T) {
	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), "file1")

	assert.Error(t, err)
	assert.Contains(t, output, "file file1 does not exist")
}

func TestLoadCmdWithInvalidYamlThenFail(t *testing.T) {
	file := test.File(t, `services:
  app:
    image: nginx
    ports: [
`)

	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to parse file")
}

func TestLoadCmdWithValidYamlAndEmptyDockerProviderThenSucceed(t *testing.T) {
	file := test.File(t, `services:
  web:
    image: nginx
`)

	_, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), file)

	assert.NoError(t, err)
}

func TestLoadCmdWithValidYamlButNoServicesThenFail(t *testing.T) {
	file := test.File(t, `version: "3.8"`)

	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), file)

	assert.Error(t, err)
	assert.Contains(t, output, "no services defined in file")
}

func TestLoadCmdWithValidYamlThenPrintServiceNames(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
  db:
    image: postgres
`)

	output, err := test.ExecuteCommand(t, NewLoadCmd(test.EmptyDockerProvider()), file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "db")
}

func TestLoadCmdWithValidYamlButUnreachableDockerThenFail(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.UnreachableDockerProvider())

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.Error(t, err)
	assert.Contains(t, output, "failed to ping docker daemon")
}

func TestLoadCmdWithValidYamlButNoContainersRunningThenPrintMissingContainer(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "missing")
}

func TestLoadCmdWithValidYamlAndContainerIsRunningButNotLaunchedFromComposeThenPrintMissingContainer(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				State: "running",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "missing")
}

func TestLoadCmdWithValidYamlAndRunningContainerThenFilterProjectNameByDirIfKeyMissing(t *testing.T) {
	file := test.File(t, `services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "001", // 001 is the temp dir name
				},
				State: "running",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
}

func TestLoadCmdWithValidYamlAndRunningContainersThenPrintRunningContainer(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "test",
				},
				State: "running",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
}

func TestLoadCmdWithValidYamlAndMultipleRunningContainersThenPrintRunningContainers(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "test",
				},
				State: "running",
			},
			{
				ID:    "0987654321",
				Names: []string{"chaosd-app-2"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "test",
				},
				State: "running",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")

	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-2")
	assert.Contains(t, output, "running")
}

func TestLoadCmdWithValidYamlAndExitedContainerThenPrintExitedContainer(t *testing.T) {
	file := test.File(t, `name: test
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "test",
				},
				State: "exited",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "exited")
}

func TestLoadCmdWithValidYamlAndMultipleProjectsThenPrintOnlyMatchingProject(t *testing.T) {
	file := test.File(t, `name: project1
services:
  web:
    image: nginx
`)

	cmd := NewLoadCmd(test.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1234567890",
				Names: []string{"chaosd-app-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
			},
			{
				ID:    "0987654321",
				Names: []string{"chaosd-app-2"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project2",
				},
				State: "running",
			},
		},
	))

	output, err := test.ExecuteCommand(t, cmd, file)

	assert.NoError(t, err)
	assert.Contains(t, output, "web")
	assert.Contains(t, output, "chaosd-app-1")
	assert.Contains(t, output, "running")
	assert.NotContains(t, output, "chaosd-app-2")
}
