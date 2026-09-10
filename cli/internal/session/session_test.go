package session_test

import (
	"testing"

	"chaosd/cli/internal/session"

	"chaosd/cli/internal/session/sessiontest"
	"chaosd/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestParseSessionID(t *testing.T) {
	id := "   sess-12345678   "

	parsedID := session.ParseSessionID(id)

	assert.Equal(t, session.SessionID("sess-12345678"), parsedID)
}

func TestNewSessionID(t *testing.T) {
	id := session.NewSessionID()

	expectedLen := 8
	prefix := "sess-"

	assert.Equal(t, expectedLen, len(id)-len(prefix))
}

func TestParseFaultID(t *testing.T) {
	id := "   fault-12345678   "

	parsedID := session.ParseFaultID(id)

	assert.Equal(t, session.FaultID("fault-12345678"), parsedID)
}

func TestNewFaultID(t *testing.T) {
	id := session.NewFaultID()

	expectedLen := 8
	prefix := "fault-"

	assert.Equal(t, expectedLen, len(id)-len(prefix))
}

func TestSessionIsPersisted(t *testing.T) {
	f := test.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	assert.Equal(t, `project1`, s.Project)
}

func TestGetSession(t *testing.T) {
	f := test.File(t, `name: project`)

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
	f := test.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	err = sessionStore.Delete(s.ID)

	assert.NoError(t, err)

	_, err = sessionStore.Get(s.ID)

	assert.Error(t, err)
}

func TestDeleteSessionNotFound(t *testing.T) {
	sessionStore := sessiontest.NewTmpSessionStore(t)

	err := sessionStore.Delete(`nonexistent`)

	assert.Error(t, err)
}

func TestAddFault(t *testing.T) {
	f := test.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	fault := session.Fault{
		ID: session.NewFaultID(),
	}

	faultID, err := sessionStore.AddFault(s.ID, fault)

	assert.NoError(t, err)

	s2, err := sessionStore.Get(s.ID)

	assert.NoError(t, err)

	assert.Equal(t, 1, len(s2.Faults))
	assert.Equal(t, faultID, s2.Faults[0].ID)
}

func TestHealFault(t *testing.T) {
	f := test.File(t, `name: project`)

	sessionStore := sessiontest.NewTmpSessionStore(t)

	s, err := sessionStore.Create(`project1`, f)

	assert.NoError(t, err)

	fault := session.Fault{
		ID:     session.NewFaultID(),
		Status: "active",
	}

	faultID, err := sessionStore.AddFault(s.ID, fault)

	assert.NoError(t, err)

	err = sessionStore.HealFault(s.ID, faultID)

	assert.NoError(t, err)

	s2, err := sessionStore.Get(s.ID)

	assert.NoError(t, err)

	assert.Equal(t, 1, len(s2.Faults))
	assert.Equal(t, "healed", s2.Faults[0].Status)
}
