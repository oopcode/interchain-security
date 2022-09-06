package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Runner struct {
	am  provider.AppModule
	k   *providerkeeper.Keeper
	ctx *sdk.Context
}

func (r *Runner) CreateConsumer() {
	var chainID string
	var ih clienttypes.Height
	var lockUbdOnTimeout bool
	r.k.CreateConsumerClient(*r.ctx, chainID, ih, lockUbdOnTimeout)
}

// SetConsumerChain TODO:
// This is called by OnChanOpenConfirm
func (r *Runner) SetConsumerChain() {
	var channelID string
	r.k.SetConsumerChain(*r.ctx, channelID)
}

func (r *Runner) StopConsumer() {
	var chainID string
	var lockUbd bool
	var closeChan bool
	r.k.StopConsumerChain(*r.ctx, chainID, lockUbd, closeChan)
}

// TrackNewUnbondingOperation TODO:
// This is called by StakingHooks::AfterUnbondingInitiated
func (r *Runner) TrackNewUnbondingOperation() {
	var id uint64
	r.k.TrackNewUnbondingOperation(*r.ctx, id)
}

func (r *Runner) SimOnRecvVSCMaturedPacket() {
	var packet channeltypes.Packet
	var data ccv.VSCMaturedPacketData
	r.k.OnRecvVSCMaturedPacket(*r.ctx, packet, data)
}

func (r *Runner) SimEndBlock() {
	// option 1
	// r.am.EndBlock() TODO: which option?
	// option 2
	r.k.CompleteMaturedUnbondingOps(*r.ctx)
	var updates []abci.ValidatorUpdate
	r.k.SendValidatorUpdates(*r.ctx, updates)
}

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

	r := Runner{provider.NewAppModule(&providerKeeper), &providerKeeper, &ctx}
	r.SimEndBlock()

}

//// Temporary below here

// Constructs a provider keeper and context object for unit tests, backed by an in-memory db.
func GetProviderKeeperAndCtx(t testing.TB) (providerkeeper.Keeper, sdk.Context) {

	cdc, storeKey, paramsSubspace, ctx := testkeeper.SetupInMemKeeper(t)

	k := providerkeeper.NewKeeper(
		cdc,
		storeKey,
		paramsSubspace,
		&testkeeper.MockScopedKeeper{},
		&testkeeper.MockChannelKeeper{},
		&testkeeper.MockPortKeeper{},
		&testkeeper.MockConnectionKeeper{},
		&testkeeper.MockClientKeeper{},
		&SpecialStakingKeeper{},
		// &testkeeper.MockStakingKeeper{},
		&testkeeper.MockSlashingKeeper{},
		&testkeeper.MockAccountKeeper{},
		"",
	)
	return k, ctx
}

type SpecialStakingKeeper struct {
}

func (k *SpecialStakingKeeper) GetValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
func (k *SpecialStakingKeeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	return nil
}
func (k *SpecialStakingKeeper) UnbondingTime(ctx sdk.Context) time.Duration {
	return 0
}
func (k *SpecialStakingKeeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool) {
	return stakingtypes.Validator{}, false
}
func (k *SpecialStakingKeeper) Jail(sdk.Context, sdk.ConsAddress) {

}
func (k *SpecialStakingKeeper) Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec, stakingtypes.InfractionType) {

}
func (k *SpecialStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	return stakingtypes.Validator{}, false
}
func (k *SpecialStakingKeeper) IterateLastValidatorPowers(ctx sdk.Context, cb func(addr sdk.ValAddress, power int64) (stop bool)) {

}
func (k *SpecialStakingKeeper) PowerReduction(ctx sdk.Context) sdk.Int {
	return sdk.ZeroInt()
}
func (k *SpecialStakingKeeper) PutUnbondingOnHold(ctx sdk.Context, id uint64) error {
	return nil
}
