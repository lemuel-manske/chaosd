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
	DockerProvider docker.DockerProvider
	EventStore     event.EventStore
	NetworkManager network.Manager
	Restarter      lifecycle.Restarter
	SessionStore   session.SessionStore
}

type AppOption func(*Application)

func WithDockerProvider(provider docker.DockerProvider) AppOption {
	return func(app *Application) {
		app.DockerProvider = provider
	}
}

func WithEventStore(store event.EventStore) AppOption {
	return func(app *Application) {
		app.EventStore = store
	}
}

func WithSessionStore(store session.SessionStore) AppOption {
	return func(app *Application) {
		app.SessionStore = store
	}
}

func WithNetworkManager(manager network.Manager) AppOption {
	return func(app *Application) {
		app.NetworkManager = manager
	}
}

func NewApplication(opts ...AppOption) Application {
	app := Application{}

	for _, opt := range opts {
		opt(&app)
	}

	return app
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

	restarter := lifecycle.NewDockerRestarter(cli)

	results = restarter.Restart(ctx, nodes)

	for _, result := range results {
		if result.Err != nil {
			return results, fmt.Errorf("failed to restart service %s: %v", serviceName, result.Err)
		}
	}

	app.EventStore.Append(sessionID, event.Event{
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
	events, err := app.EventStore.List(sessionID)

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
) (session.FaultID, error) {
	_session, err := app.SessionStore.Get(sessionID)

	if err != nil {
		return "", err
	}

	composeFile, err := docker.Parse(_session.ComposeFile)

	if err != nil {
		return "", err
	}

	cli, err := app.DockerProvider.NewClient()

	if err != nil {
		return "", newDockerClientError(err)
	}

	t, err := topology.Load(composeFile, ctx, cli)

	if err != nil {
		return "", err
	}

	nodeA, err := getRunningNode(t, nodeAName)

	if err != nil {
		return "", err
	}

	nodeB, err := getRunningNode(t, nodeBName)

	if err != nil {
		return "", err
	}

	faultID, err := app.SessionStore.AddPartitionFault(sessionID, nodeAName, nodeBName)

	if err != nil {
		return "", err
	}

	err = app.NetworkManager.Partition(ctx, *nodeA, *nodeB, string(faultID))

	if err != nil {
		return "", err
	}

	app.EventStore.Append(sessionID, event.Event{
		Type:      event.PartitionAppliedEvent,
		CreatedAt: time.Now(),
		Data: event.PartitionAppliedEventData{
			NodeAName: nodeAName,
			NodeBName: nodeBName,
		},
	})

	return faultID, nil
}

func (app *Application) Heal(
	ctx context.Context,
	sessionID session.SessionID,
	faultID session.FaultID,
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

	fault := _session.GetFault(faultID)

	if fault == nil {
		return fmt.Errorf("fault %s not found", faultID)
	}

	// TODO: do not recalculate IPs
	// is the containers are recreated, the IPs will change, and the partition will not be healed correctly

	nodeA, err := getRunningNode(t, fault.NodeA)

	if err != nil {
		return err
	}

	nodeB, err := getRunningNode(t, fault.NodeB)

	if err != nil {
		return err
	}

	if fault.IsHealed() {
		return fmt.Errorf("fault %s is already healed", faultID)
	}

	err = app.NetworkManager.Heal(ctx, *nodeA, *nodeB, string(fault.ID))

	if err != nil {
		return err
	}

	err = app.SessionStore.HealFault(sessionID, faultID)

	if err != nil {
		return err
	}

	app.EventStore.Append(sessionID, event.Event{
		Type:      event.HealAppliedEvent,
		CreatedAt: time.Now(),
		Data: event.HealAppliedEventData{
			NodeAName: nodeA.ContainerName,
			NodeBName: nodeB.ContainerName,
		},
	})

	return nil
}

func (app *Application) Delay(
	ctx context.Context,
	sessionID session.SessionID,
	nodeAName string,
	nodeBName string,
	delay time.Duration,
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

	faultID, err := app.SessionStore.AddDelayFault(sessionID, nodeAName, nodeBName, delay)

	if err != nil {
		return err
	}

	err = app.NetworkManager.Delay(ctx, *nodeA, *nodeB, string(faultID), delay)

	if err != nil {
		return err
	}

	app.EventStore.Append(sessionID, event.Event{
		Type:      event.DelayAppliedEvent,
		CreatedAt: time.Now(),
		Data: event.DelayAppliedEventData{
			Delay:     delay,
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
