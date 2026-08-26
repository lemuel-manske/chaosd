package topology

import (
	"context"
	"testing"

	"chaosd/cli/internal/docker"

	"chaosd/cli/internal/docker/dockertest"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestLoadTopologyWithValidComposeFileAndRunningContainer(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
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
		},
	)

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	actualTopology, err := LoadTopology(
		&composeFile,
		ctx,
		dockerClient,
	)

	expectedTopology := Topology{
		Project: "project1",
		Nodes: []Node{
			{
				Service:       "web",
				ContainerID:   "1234567890",
				ContainerName: "chaosd-app-1",
				State:         "running",
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedTopology, *actualTopology)
}

func TestLoadTopologyWithValidComposeFileAndNoRunningContainer(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{},
	)

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	actualTopology, err := LoadTopology(
		&composeFile,
		ctx,
		dockerClient,
	)

	expectedTopology := Topology{
		Project: "project1",
		Nodes: []Node{
			{
				Service: "web",
				State:   "missing",
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedTopology, *actualTopology)
}

func TestLoadTopologyWithValidComposeFileAndMultipleRunningContainers(t *testing.T) {
	dockerProvider := dockertest.FakeDockerProvider(
		[]container.Summary{
			{
				ID:    "1",
				Names: []string{"project-web-1"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
			},
			{
				ID:    "2",
				Names: []string{"project-web-2"},
				Labels: map[string]string{
					"com.docker.compose.service": "web",
					"com.docker.compose.project": "project1",
				},
				State: "running",
			},
		},
	)

	dockerClient, _ := dockerProvider.NewClient()

	composeFile := docker.ComposeFile{
		Name:    "project1",
		Version: "3",
		Services: map[string]any{
			"web": map[string]any{
				"image": "nginx",
			},
		},
	}

	ctx := context.Background()

	actualTopology, err := LoadTopology(
		&composeFile,
		ctx,
		dockerClient,
	)

	expectedTopology := Topology{
		Project: "project1",
		Nodes: []Node{
			{
				Service:       "web",
				ContainerID:   "1",
				ContainerName: "project-web-1",
				State:         "running",
			},
			{
				Service:       "web",
				ContainerID:   "2",
				ContainerName: "project-web-2",
				State:         "running",
			},
		},
	}

	assert.NoError(t, err)
	assert.Equal(t, expectedTopology, *actualTopology)
}
