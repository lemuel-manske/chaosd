package docker

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerClient interface {
	Ping(ctx context.Context) error

	ContainerList(
		ctx context.Context,
		options client.ContainerListOptions,
	) ([]container.Summary, error)
}

type DockerProvider interface {
	NewClient() (DockerClient, error)
}

func NewDockerProvider() DockerProvider {
	return &dockerProvider{}
}
