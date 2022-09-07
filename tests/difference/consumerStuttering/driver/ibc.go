package consumerStuttering

import (
	"fmt"

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
	fmt.Println("ChanCloseInit")
	return nil
}

func (m *MockChannelKeeper) GetChannel(ctx types.Context, srcPort, srcChan string) (types6.Channel, bool) {
	fmt.Println("GetChannel")
	return types6.Channel{ConnectionHops: []string{}}, true
	return types6.Channel{}, false
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx types.Context, portID, channelID string) (uint64, bool) {
	fmt.Println("GetNextSequenceSend")
	return 0, false
}

func (m *MockChannelKeeper) SendPacket(ctx types.Context, channelCap *types1.Capability, packet exported.PacketI) error {
	fmt.Println("SendPacket")
	return nil
}

func (m *MockChannelKeeper) WriteAcknowledgement(ctx types.Context, chanCap *types1.Capability, packet exported.PacketI, acknowledgement exported.Acknowledgement) error {
	fmt.Println("WriteAck")
	return nil
}

type MockPortKeeper struct {
}

func NewMockPortKeeper() *MockPortKeeper {
	return &MockPortKeeper{}
}

func (m *MockPortKeeper) BindPort(ctx types.Context, portID string) *types1.Capability {
	fmt.Println("BindPort")

	return &types1.Capability{}
}

type MockConnectionKeeper struct {
}

func NewMockConnectionKeeper() *MockConnectionKeeper {
	return &MockConnectionKeeper{}
}

func (m *MockConnectionKeeper) GetConnection(ctx types.Context, connectionID string) (types5.ConnectionEnd, bool) {
	fmt.Println("GetConnection")

	return types5.ConnectionEnd{}, false
}

type MockClientKeeper struct {
}

func NewMockClientKeeper() *MockClientKeeper {
	return &MockClientKeeper{}
}

func (m *MockClientKeeper) CreateClient(ctx types.Context, clientState exported.ClientState, consensusState exported.ConsensusState) (string, error) {
	fmt.Println("CreateClient")
	return "", nil
}

func (m *MockClientKeeper) GetClientState(ctx types.Context, clientID string) (exported.ClientState, bool) {
	fmt.Println("GetClientState")
	return &ibctmtypes.ClientState{}, false
}

func (m *MockClientKeeper) GetLatestClientConsensusState(ctx types.Context, clientID string) (exported.ConsensusState, bool) {
	fmt.Println("GetLatestClientConsensusState")
	return &ibctmtypes.ConsensusState{}, false
}

func (m *MockClientKeeper) GetSelfConsensusState(ctx types.Context, height exported.Height) (exported.ConsensusState, error) {
	fmt.Println("GetSelfConsensusState")
	return &ibctmtypes.ConsensusState{}, nil
}
