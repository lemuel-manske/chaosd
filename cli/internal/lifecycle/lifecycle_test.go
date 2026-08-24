package lifecycle

import (
	"context"
	"testing"

	"chaosd/cli/internal/topology"
	"chaosd/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestRestartLifecycle(t *testing.T) {
	dockerClient := &test.DockerClientMock{}

	l := NewLifecycle(dockerClient)

	ctx := context.Background()

	res := l.Restart(ctx, []topology.Node{
		{
			Service:       "web",
			ContainerID:   "1234567890",
			ContainerName: "chaosd-app-1",
			State:         "running",
		},
	})

	assert.Len(t, res, 1)

	stNode := res[0]

	assert.Equal(t, "web", stNode.Node.Service)
	assert.Nil(t, stNode.Err)
}

func TestRestartLifecycleWithError(t *testing.T) {
	dockerClient := &test.DockerClientMock{
		RestartErr: map[string]error{
			"1234567890": assert.AnError,
		},
	}

	l := NewLifecycle(dockerClient)

	ctx := context.Background()

	res := l.Restart(ctx, []topology.Node{
		{
			Service:       "web",
			ContainerID:   "1234567890",
			ContainerName: "chaosd-app-1",
			State:         "running",
		},
	})

	assert.Len(t, res, 1)

	stNode := res[0]

	assert.Equal(t, "web", stNode.Node.Service)
	assert.NotNil(t, stNode.Err)
}
