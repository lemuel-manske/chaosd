package session_test

import (
	"testing"

	"chaosd/cli/clitest"
	"chaosd/cli/internal/session/sessiontest"

	"github.com/stretchr/testify/assert"
)

func TestSessionIsPersisted(t *testing.T) {
	f := clitest.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	assert.Equal(t, `project1`, s.Project)
}

func TestGetSession(t *testing.T) {
	f := clitest.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	s2, err := sessionStore.Get(s.ID)

	assert.NoError(t, err)

	assert.Equal(t, s.ID, s2.ID)
	assert.Equal(t, s.Project, s2.Project)
	assert.Equal(t, s.ComposeFile, s2.ComposeFile)
}

func TestGetSessionNotFound(t *testing.T) {
	sessionStore := sessiontest.NewTmpSessionStore(t)

	_, err := sessionStore.Get(`nonexistent`)

	assert.Error(t, err)
}

func TestDeleteSession(t *testing.T) {
	f := clitest.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	err = sessionStore.Delete(s.ID)

	assert.NoError(t, err)

	_, err = sessionStore.Get(s.ID)

	assert.Error(t, err)
}
