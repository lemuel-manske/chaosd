package docker

import "context"

type DockerClient interface {
	Ping(ctx context.Context) error
}

type DockerProvider interface {
	NewClient() (DockerClient, error)
}

func NewDockerProvider() DockerProvider {
	return &dockerProvider{}
}
