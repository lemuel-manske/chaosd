package apptest

import (
	"testing"

	"chaosd/cli/application"

	"chaosd/cli/internal/docker/dockertest"
	"chaosd/cli/internal/network/networktest"
	"chaosd/cli/internal/session/sessiontest"
)

func NewIntegrationTestApp(t *testing.T) application.Application {
	t.Helper()

	sessionStore := sessiontest.NewTmpSessionStore(t)

	dockerProvider := dockertest.NewRealDockerProvider()

	networkManager := networktest.NewRealManager()

	return *application.NewApplication(
		sessionStore,
		dockerProvider,
		networkManager,
	)
}
