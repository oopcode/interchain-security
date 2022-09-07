package consumerStuttering

import (
	types "github.com/cosmos/cosmos-sdk/types"
	types1 "github.com/cosmos/cosmos-sdk/x/capability/types"
	types5 "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	types6 "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	exported "github.com/cosmos/ibc-go/v3/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
)

type MockChannelKeeper struct {
}

func NewMockChannelKeeper() *MockChannelKeeper {
	return &MockChannelKeeper{}
}

func (m *MockChannelKeeper) ChanCloseInit(ctx types.Context, portID, channelID string, chanCap *types1.Capability) error {
	return nil
}

func (m *MockChannelKeeper) GetChannel(ctx types.Context, srcPort, srcChan string) (types6.Channel, bool) {
	return types6.Channel{}, false
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx types.Context, portID, channelID string) (uint64, bool) {
	return 0, false
}

func (m *MockChannelKeeper) SendPacket(ctx types.Context, channelCap *types1.Capability, packet exported.PacketI) error {
	return nil
}

func (m *MockChannelKeeper) WriteAcknowledgement(ctx types.Context, chanCap *types1.Capability, packet exported.PacketI, acknowledgement exported.Acknowledgement) error {
	return nil
}

type MockPortKeeper struct {
}

func NewMockPortKeeper() *MockPortKeeper {
	return &MockPortKeeper{}
}

func (m *MockPortKeeper) BindPort(ctx types.Context, portID string) *types1.Capability {
	return &types1.Capability{}
}

type MockConnectionKeeper struct {
}

func NewMockConnectionKeeper() *MockConnectionKeeper {
	return &MockConnectionKeeper{}
}

func (m *MockConnectionKeeper) GetConnection(ctx types.Context, connectionID string) (types5.ConnectionEnd, bool) {
	return types5.ConnectionEnd{}, false
}

type MockClientKeeper struct {
}

func NewMockClientKeeper() *MockClientKeeper {
	return &MockClientKeeper{}
}

func (m *MockClientKeeper) CreateClient(ctx types.Context, clientState exported.ClientState, consensusState exported.ConsensusState) (string, error) {
	return "", nil
}

func (m *MockClientKeeper) GetClientState(ctx types.Context, clientID string) (exported.ClientState, bool) {
	// return
	return &ibctmtypes.ClientState{}, false
}

func (m *MockClientKeeper) GetLatestClientConsensusState(ctx types.Context, clientID string) (exported.ConsensusState, bool) {
	return &ibctmtypes.ConsensusState{}, false
}

func (m *MockClientKeeper) GetSelfConsensusState(ctx types.Context, height exported.Height) (exported.ConsensusState, error) {
	return &ibctmtypes.ConsensusState{}, nil
}

/*

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.ClientStateKey())
	if bz == nil {
		return nil, false
	}

	clientState := k.MustUnmarshalClientState(bz)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Context, clientID string, clientState exported.ClientState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.ClientStateKey(), k.MustMarshalClientState(clientState))
}

// GetClientConsensusState gets the stored consensus state from a client at a given height.
func (k Keeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz := store.Get(host.ConsensusStateKey(height))
	if bz == nil {
		return nil, false
	}

	consensusState := k.MustUnmarshalConsensusState(bz)
	return consensusState, true
}
*/
