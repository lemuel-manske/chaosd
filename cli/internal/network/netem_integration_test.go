// go:build integration

package network_test

import (
	"context"
	"testing"
	"time"

	"chaosd/cli/internal/network"

	"chaosd/cli/internal/network/networktest"

	"github.com/stretchr/testify/assert"
)

func TestNetemInjector_DelayAndHeal(t *testing.T) {
	ctx := context.Background()

	testNetwork := networktest.SetupNetwork(t)

	injector := network.NewNetemInjector()

	before := networktest.MeasurePing(t, testNetwork.TargetIP)

	results := injector.Delay(ctx, network.DelayRequest{
		Links: []network.Link{
			{
				SourceIP: testNetwork.SourceIP,
				TargetIP: testNetwork.TargetIP,
			},
		},
		Delay: 200 * time.Millisecond,
	})

	assert.NoError(t, results[0].Err)

	afterDelay := networktest.MeasurePing(t, testNetwork.TargetIP)

	assert.GreaterOrEqual(
		t,
		afterDelay,
		before+150*time.Millisecond,
	)

	results = injector.Heal(ctx, network.HealRequest{
		Links: []network.Link{
			{
				SourceIP: testNetwork.SourceIP,
				TargetIP: testNetwork.TargetIP,
			},
		},
	})

	assert.NoError(t, results[0].Err)

	afterHeal := networktest.MeasurePing(t, testNetwork.TargetIP)

	assert.Less(
		t,
		afterHeal,
		afterDelay-150*time.Millisecond,
	)
}
