package root

import (
	"testing"

	"chaosd/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestRootCmdGeneratesNoErrors(t *testing.T) {
	_, err := test.ExecuteCommand(t, NewRootCmd())

	assert.NoError(t, err)
}
