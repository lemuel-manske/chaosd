//go:build !integration

package root

import (
	"testing"

	"chaosd/cli/clitest"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd_NoArguments_Succeeds(t *testing.T) {
	_, err := clitest.ExecuteCommand(t, NewRootCmd())

	assert.NoError(t, err)
}
