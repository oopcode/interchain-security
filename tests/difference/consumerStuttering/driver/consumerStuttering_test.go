package consumerStuttering

import (
	"strconv"
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

type Runner struct {
	t         *testing.T
	ctx       *sdk.Context
	am        provider.AppModule
	k         *providerkeeper.Keeper
	sk        *SpecialStakingKeeper
	lastState State
}

func GetProviderKeeperAndCtx(t testing.TB, stakingKeeper ccv.StakingKeeper) (providerkeeper.Keeper, sdk.Context) {

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
		stakingKeeper,
		// &SpecialStakingKeeper{},
		// &testkeeper.MockStakingKeeper{},
		&testkeeper.MockSlashingKeeper{},
		&testkeeper.MockAccountKeeper{},
		"",
	)
	return k, ctx
}

func NewRunner(t *testing.T, initState State) *Runner {
	stakingKeeper := NewSpecialStakingKeeper()
	providerKeeper, ctx := GetProviderKeeperAndCtx(t, stakingKeeper)

	r := Runner{t, &ctx, provider.NewAppModule(&providerKeeper), &providerKeeper, stakingKeeper, initState}
	_ = r

	return &r
}

func stringifyChainId(id int) string {
	return "chain" + strconv.Itoa(id)
}

func (r *Runner) InitConsumer(c int) {
	var chainID string
	chainID = stringifyChainId(c)
	var ih clienttypes.Height
	ih.RevisionHeight = 0
	ih.RevisionNumber = 0
	var lockUbdOnTimeout bool
	lockUbdOnTimeout = false
	r.k.CreateConsumerClient(*r.ctx, chainID, ih, lockUbdOnTimeout)
}

// ActivateConsumer TODO:
// This is called by OnChanOpenConfirm
func (r *Runner) ActivateConsumer(c int) {
	var channelID string
	r.k.SetConsumerChain(*r.ctx, channelID)
}

func (r *Runner) StopConsumer(c int) {
	var chainID string
	chainID = stringifyChainId(c)
	var lockUbd bool
	var closeChan bool
	r.k.StopConsumerChain(*r.ctx, chainID, lockUbd, closeChan)
}

func (r *Runner) EndBlock(awaitedVscIds [][]int) {
	// option 1
	// r.am.EndBlock() TODO: which option?
	// option 2
	r.k.CompleteMaturedUnbondingOps(*r.ctx)

	// Add some more unbonding operations to be sent with this VSC
	valUpdateID := r.k.GetValidatorSetUpdateId(*r.ctx)
	for i := 0; i < 3; i++ { // TODO: param
		r.sk.AddUnbondingOperation(valUpdateID, func(id uint64) {
			r.k.TrackNewUnbondingOperation(*r.ctx, id)
		})
	}

	var updates []abci.ValidatorUpdate
	r.k.SendValidatorUpdates(*r.ctx, updates)

	checkNoUnbondEarly(r.t, r.sk.refCnt, r.sk.vscIdToOpids, awaitedVscIds)

	checkNoUnbondLate(r.t, r.sk.refCnt, r.sk.vscIdToOpids, awaitedVscIds, valUpdateID)

}

func (r *Runner) RecvMaturity(c int, vscId int) {
	var packet channeltypes.Packet
	packet.DestinationChannel = ""
	var data ccv.VSCMaturedPacketData
	data.ValsetUpdateId = 0
	r.k.OnRecvVSCMaturedPacket(*r.ctx, packet, data)
}

// apa simulate --output-traces --length=20 --max-run=10 main.tla

func (r *Runner) handleState(s State) {
	if s.Kind == "InitConsumer" {
		// Get newly initialised
		id := getDifferentInt(s.InitialisingConsumers, r.lastState.InitialisingConsumers)
		_ = id
		r.InitConsumer(*id)
	}
	if s.Kind == "ActivateConsumer" {
		// Get newly active
		id := getDifferentInt(s.ActiveConsumers, r.lastState.ActiveConsumers)
		_ = id
		r.ActivateConsumer(*id)
	}
	if s.Kind == "StopConsumer" {
		knew := []int{}
		knew = append(knew, r.lastState.InitialisingConsumers...)
		knew = append(knew, r.lastState.ActiveConsumers...)
		id := getDifferentInt(knew, []int{})
		_ = id
		r.StopConsumer(*id)
	}
	if s.Kind == "EndBlock" {
		r.EndBlock(s.AwaitedVscIds)
	}
	if s.Kind == "RecvMaturity" {
		pair := getDifferentIntPair(r.lastState.AwaitedVscIds, s.AwaitedVscIds)
		_ = pair
		r.RecvMaturity(pair[0], pair[1])
	}
	r.lastState = s
}

func TestTraces(t *testing.T) {
	data := LoadTraces("traces.json")
	for i, trace := range data {
		_ = i
		_ = trace
		initState := trace.States[0]
		runner := NewRunner(t, initState)
		for j, s := range trace.States[1:] {
			_ = j
			_ = s
			runner.handleState(s)
		}
	}
}

// Properties

// checkNoUnbondEarly checks that for all vscIds which are still awaited, the refCnts
// for all unbonding operations associated to the vscId are positive
func checkNoUnbondEarly(t *testing.T, refCnt map[uint64]int, vscIdToOpids map[uint64][]uint64,
	awaitedVscIds [][]int) {
	for _, pair := range awaitedVscIds {
		vscId := pair[1]
		for _, opId := range vscIdToOpids[uint64(vscId)] {
			if refCnt[opId] < 1 {
				t.Fatalf("fail checkNoUnbondEarly")
			}
		}
	}
}

// checkNoUnbondLate checks that for all vscId < valUpdateId: if there is NOT an awaited
// maturity for that vscID, then the refCnts for all unbonding operations associated to
// the vscID are 0
func checkNoUnbondLate(t *testing.T, refCnt map[uint64]int, vscIdToOpids map[uint64][]uint64,
	awaitedVscIds [][]int, maxVscIdToCheck uint64) {
	stillAwaiting := make([]bool, maxVscIdToCheck)
	for _, pair := range awaitedVscIds {
		vscId := pair[1]
		stillAwaiting[vscId] = true
	}
	for vscId, mustWaitLonger := range stillAwaiting {
		if !mustWaitLonger {
			for _, opId := range vscIdToOpids[uint64(vscId)] {
				if refCnt[opId] != 0 {
					t.Fatalf("fail checkNoUnbondLate")
				}
			}
		}
	}
}

//// Temporary below here (scratch code)

type SpecialStakingKeeper struct {
	// Controlled by this
	nextOpId uint64
	// Unbonding op id to reference count
	// Initialised by this, modified by staking module
	refCnt map[uint64]int
	// Initialised by this, modified by staking module
	vscIdToOpids map[uint64][]uint64
}

func NewSpecialStakingKeeper() *SpecialStakingKeeper {
	return &SpecialStakingKeeper{
		0,
		map[uint64]int{},
		map[uint64][]uint64{},
	}
}

func (k *SpecialStakingKeeper) AddUnbondingOperation(vscId uint64, callback func(uint64)) {
	if _, ok := k.vscIdToOpids[vscId]; !ok {
		//do something here
		k.vscIdToOpids[vscId] = []uint64{}
	}
	k.vscIdToOpids[vscId] = append(k.vscIdToOpids[vscId], k.nextOpId)
	k.refCnt[k.nextOpId] = 0
	callback(k.nextOpId)
	k.nextOpId += 1
}

func (k *SpecialStakingKeeper) GetValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
func (k *SpecialStakingKeeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	k.refCnt[id] -= 1
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
	k.refCnt[id] += 1
	return nil
}
