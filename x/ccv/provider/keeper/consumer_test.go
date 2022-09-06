package keeper_test

import (
	"testing"

	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
)

// TestMultipleConsumers TODO:
func TestMultipleConsumers(t *testing.T) {
	/*
		Actions should be:
		-
	*/
	/*
		Notes:
			Provider EndBlock does
				- CompleteMaturedUnbondingOps
				- SendValidatorUpdates

			Provider OnRecvVSCMaturedPacket does
				- check if packet.DestinationChannel channel exists
				- uses the packet data.ValsetUpdateId to do business logic

			Provider proposals (create and stop) call into (respectively)
				- CreateConsumerClient
				- StopConsumerChain

			Provider OnChanOpenConfirm (last handshake step) does
				- SetConsumerChain

			Provider AfterUnbondingInitiated does
				- uses iterator IterateConsumerChains
				- increments ref cnt and tracks opId for each chain



	*/

	providerKeeper, ctx := testkeeper.GetProviderKeeperAndCtx(t)

}
