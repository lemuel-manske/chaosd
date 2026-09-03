package application

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"chaosd/cli/internal/docker"
	"chaosd/cli/internal/event"
	"chaosd/cli/internal/lifecycle"
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"
	"chaosd/cli/internal/topology"
)

// Application represents the application layer of the CLI, to keep it (the CLI) thin
type Application struct {
	SessionStore   session.SessionStore
	DockerProvider docker.DockerProvider
	NetworkManager network.Manager
	Lifecycle      lifecycle.Lifecycle

	events event.EventStore
}

func NewApplication(
	sessionStore session.SessionStore,
	eventStore event.EventStore,
	dockerProvider docker.DockerProvider,
	networkManager network.Manager,
) *Application {
	dockerClient, err := dockerProvider.NewClient()

	if err != nil {
		panic(err)
	}

	lifecycle := lifecycle.NewLifecycle(dockerClient)

	return &Application{
		DockerProvider: dockerProvider,
		Lifecycle:      lifecycle,
		NetworkManager: networkManager,
		SessionStore:   sessionStore,

		events: eventStore,
	}
}

func (app *Application) GetTopology(
	ctx context.Context,
	sessionID session.SessionID,
) (*topology.Topology, error) {
	foundSession, err := app.SessionStore.Get(sessionID)

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

func (app *Application) Load(
	ctx context.Context,
	composeFilePath string,
) (session.SessionID, error) {
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
	sessionID session.SessionID,
	serviceName string,
) ([]lifecycle.ActionResult, error) {
	results := []lifecycle.ActionResult{}

	session, err := app.SessionStore.Get(sessionID)

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

	app.events.Append(sessionID, event.Event{
		Type:      event.RestartEvent,
		CreatedAt: time.Now(),
		Data: event.RestartEventData{
			ServiceName: serviceName,
		},
	})

	return results, nil
}

func (app *Application) ListEvents(
	ctx context.Context,
	sessionID session.SessionID,
) ([]event.Event, error) {
	events, err := app.events.List(sessionID)

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (app *Application) Partition(
	ctx context.Context,
	sessionID session.SessionID,
	nodeAName string,
	nodeBName string,
) error {
	_session, err := app.SessionStore.Get(sessionID)

	if err != nil {
		return err
	}

	composeFile, err := docker.Parse(_session.ComposeFile)

	if err != nil {
		return err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return newDockerClientError(err)
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

	faultID, err := app.SessionStore.AddPartitionFault(sessionID, nodeAName, nodeBName)

	if err != nil {
		return err
	}

	err = app.NetworkManager.Partition(ctx, *nodeA, *nodeB, string(faultID))

	if err != nil {
		return err
	}

	app.events.Append(sessionID, event.Event{
		Type:      event.PartitionAppliedEvent,
		CreatedAt: time.Now(),
		Data: event.PartitionAppliedEventData{
			NodeAName: nodeAName,
			NodeBName: nodeBName,
		},
	})

	return nil
}

func (app *Application) Heal(
	ctx context.Context,
	sessionID session.SessionID,
	nodeAName string,
	nodeBName string,
) error {
	_session, err := app.SessionStore.Get(sessionID)

	if err != nil {
		return err
	}

	composeFile, err := docker.Parse(_session.ComposeFile)

	if err != nil {
		return err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return newDockerClientError(err)
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

	fault := _session.GetFault(nodeAName, nodeBName)

	if fault == nil {
		return fmt.Errorf("no partition fault found between %s and %s", nodeAName, nodeBName)
	}

	if fault.IsHealed() {
		return fmt.Errorf("partition fault between %s and %s is already healed", nodeAName, nodeBName)
	}

	err = app.NetworkManager.Heal(ctx, *nodeA, *nodeB, string(fault.ID))

	if err != nil {
		return err
	}

	err = app.SessionStore.HealPartitionFault(sessionID, nodeAName, nodeBName)

	if err != nil {
		return err
	}

	app.events.Append(sessionID, event.Event{
		Type:      event.HealAppliedEvent,
		CreatedAt: time.Now(),
		Data: event.HealAppliedEventData{
			NodeAName: nodeAName,
			NodeBName: nodeBName,
		},
	})

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
