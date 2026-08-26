package session

import (
	"testing"

	"chaosd/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestSessionIsPersisted(t *testing.T) {
	f := test.File(t, `name: project`)

	s, err := CreateSession(`project1`, f)

	assert.NoError(t, err)

	assert.Equal(t, `project1`, s.Project)
}

func TestGetSession(t *testing.T) {
	f := test.File(t, `name: project`)

	s, err := CreateSession(`project1`, f)

	assert.NoError(t, err)

	s2, err := GetSession(s.ID)

	assert.NoError(t, err)

	assert.Equal(t, s.ID, s2.ID)
	assert.Equal(t, s.Project, s2.Project)
	assert.Equal(t, s.ComposeFile, s2.ComposeFile)
}

func TestGetSessionNotFound(t *testing.T) {
	_, err := GetSession(`nonexistent`)

	assert.Error(t, err)
}

func TestDeleteSession(t *testing.T) {
	f := test.File(t, `name: project`)

	s, err := CreateSession(`project1`, f)

	assert.NoError(t, err)

	err = DeleteSession(s.ID)

	assert.NoError(t, err)

	_, err = GetSession(s.ID)

	assert.Error(t, err)
}
