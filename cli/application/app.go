package application

import (
	"context"
	"fmt"
	"path/filepath"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/lifecycle"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"
)

// Application represents the application layer of the CLI, to keep it (the CLI) thin
type Application struct {
	SessionStore   session.Store
	DockerProvider docker.DockerProvider
	NetworkManager network.Manager
	Lifecycle      lifecycle.Lifecycle
}

func (app *Application) GetTopology(
	ctx context.Context,
	sessionID string,
) (*topology.Topology, error) {
	id := session.SessionID(sessionID)

	foundSession, err := app.SessionStore.Get(id)

	if err != nil {
		return nil, err
	}

	composeFile, err := docker.Parse(foundSession.ComposeFile)

	if err != nil {
		return nil, err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return nil, newDockerClientError(err)
	}

	t, err := topology.Load(composeFile, ctx, cli)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (app *Application) Load(ctx context.Context, composeFilePath string) (session.SessionID, error) {
	composeFileAbsPath, err := filepath.Abs(composeFilePath)

	if err != nil {
		return "", err
	}

	composeFile, err := docker.Parse(composeFilePath)

	if err != nil {
		return "", err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return "", newDockerClientError(err)
	}

	_, err = topology.Load(composeFile, ctx, cli)

	if err != nil {
		return "", err
	}

	projectName := composeFile.Name

	createdSession, err := app.SessionStore.Create(
		projectName,
		composeFileAbsPath,
	)

	if err != nil {
		return "", err
	}

	return createdSession.ID, nil
}

func (app *Application) RestartService(
	ctx context.Context,
	sessionID string,
	serviceName string,
) ([]lifecycle.ActionResult, error) {
	id := session.SessionID(sessionID)

	results := []lifecycle.ActionResult{}

	session, err := app.SessionStore.Get(id)

	if err != nil {
		return results, err
	}

	composeFile, err := docker.Parse(session.ComposeFile)

	if err != nil {
		return results, err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return results, newDockerClientError(err)
	}

	t, err := topology.Load(composeFile, ctx, cli)

	if err != nil {
		return results, err
	}

	nodes := t.NodesByServiceName(serviceName)

	if len(nodes) == 0 {
		err := fmt.Errorf(
			"service %s not found in project %s",
			serviceName,
			t.Project,
		)

		return results, err
	}

	manager := lifecycle.NewLifecycle(cli)

	results = manager.Restart(ctx, nodes)

	return results, nil
}

func (app *Application) Partition(
	ctx context.Context,
	sessionID string,
	nodeAName string,
	nodeBName string,
) error {
	id := session.SessionID(sessionID)

	session, err := app.SessionStore.Get(id)

	if err != nil {
		return err
	}

	composeFile, err := docker.Parse(session.ComposeFile)

	if err != nil {
		return err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return fmt.Errorf("failed to create docker client: %v", err)
	}

	t, err := topology.Load(composeFile, ctx, cli)

	if err != nil {
		return err
	}

	nodeA, err := getRunningNode(t, nodeAName)

	if err != nil {
		return err
	}

	nodeB, err := getRunningNode(t, nodeBName)

	if err != nil {
		return err
	}

	err = app.NetworkManager.Partition(ctx, *nodeA, *nodeB)

	if err != nil {
		return err
	}

	return nil
}

func getRunningNode(t *topology.Topology, name string) (*topology.Node, error) {
	node := t.NodeByName(name)
	if node == nil {
		return nil, fmt.Errorf("%s missing", name)
	}

	if node.State != "running" {
		return nil, fmt.Errorf("%s is not running", name)
	}

	return node, nil
}

func newDockerClientError(err error) error {
	return fmt.Errorf("failed to create docker client: %v", err)
}
