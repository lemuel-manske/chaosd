package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmdGeneratesNoErrors(t *testing.T) {
	_, err := executeCommand(t, RootCmd)

	assert.NoError(t, err)
}
