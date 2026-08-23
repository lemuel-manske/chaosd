package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert/yaml"
)

type ComposeFile struct {
	Name     string         `yaml:"name"`
	Version  string         `yaml:"version"`
	Services map[string]any `yaml:"services"`
}

type DockerClient interface {
	Ping(ctx context.Context) error

	ContainerList(
		ctx context.Context,
		options client.ContainerListOptions,
	) ([]container.Summary, error)

	RestartContainer(
		ctx context.Context,
		containerID string,
		options client.ContainerRestartOptions,
	) (client.ContainerRestartResult, error)
}

type DockerProvider interface {
	NewClient() (DockerClient, error)
}

func NewDockerProvider() DockerProvider {
	return &dockerProvider{}
}

func Parse(file string) (*ComposeFile, error) {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s does not exist", file)
		}

		return nil, err
	}

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", file, err)
	}

	var compose ComposeFile

	if err := yaml.Unmarshal(yamlFile, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse file %s", file)
	}

	if len(compose.Services) == 0 {
		return nil, fmt.Errorf("no services defined in file %s", file)
	}

	if compose.Name == "" {
		absFile, err := filepath.Abs(file)

		if err != nil {
			return nil, fmt.Errorf("failed to resolve file path %s: %v", file, err)
		}

		compose.Name = filepath.Base(filepath.Dir(absFile))
	}

	return &compose, nil
}
