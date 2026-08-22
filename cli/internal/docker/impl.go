package docker

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type dockerClient struct {
	cli *client.Client
}

type dockerProvider struct{}

func (d *dockerProvider) NewClient() (DockerClient, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &dockerClient{
		cli: cli,
	}, nil
}

func (d *dockerClient) Ping(ctx context.Context) error {
	_, err := d.cli.Ping(ctx, client.PingOptions{})
	return err
}

func (d *dockerClient) ContainerList(
	ctx context.Context,
	options client.ContainerListOptions,
) ([]container.Summary, error) {
	list, err := d.cli.ContainerList(ctx, options)

	return list.Items, err
}
