package consumerStuttering

import (
	"strconv"
	"testing"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
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

	checkNoUnbondEarly(r.t, r.sk.RefCnt, r.sk.VscIdToOpIds, awaitedVscIds)

	checkNoUnbondLate(r.t, r.sk.RefCnt, r.sk.VscIdToOpIds, awaitedVscIds, valUpdateID)

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

// go test -v -timeout 10m -run TestTraces
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
