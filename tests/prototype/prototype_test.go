package core

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	simapp "github.com/cosmos/interchain-security/testutil/simapp"
	simibc "github.com/cosmos/interchain-security/testutil/simibc"
	providertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
)

const P = "provider"
const C = "consumer"

type InitState struct {
	PKSeeds         []string
	NumValidators   int
	MaxValidators   int
	SlashDoublesign sdk.Dec
	SlashDowntime   sdk.Dec
	UnbondingP      time.Duration
	UnbondingC      time.Duration
	Trusting        time.Duration
	MaxClockDrift   time.Duration
	BlockDuration   time.Duration
	ConsensusParams *abci.ConsensusParams
	MaxEntries      int
}

var initState InitState

func init() {
	//	tokens === power
	sdk.DefaultPowerReduction = sdk.NewInt(1)
	initState = InitState{
		PKSeeds: []string{
			// Fixed seeds are used to create the private keys for validators.
			// The seeds are chosen to ensure that the resulting validators are
			// sorted in descending order by the staking module.
			"bbaaaababaabbaabababbaabbbbbbaaa",
			"abbbababbbabaaaaabaaabbbbababaab",
			"bbabaabaabbbbbabbbaababbbbabbbbb",
			"aabbbabaaaaababbbabaabaabbbbbbba"},
		NumValidators:   4,
		MaxValidators:   2,
		SlashDoublesign: sdk.NewDec(0),
		SlashDowntime:   sdk.NewDec(0),
		UnbondingP:      time.Second * 70,
		UnbondingC:      time.Second * 50,
		Trusting:        time.Second * 49,
		MaxClockDrift:   time.Second * 10000,
		BlockDuration:   time.Second * 6,
		MaxEntries:      1000000,
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxBytes: 9223372036854775807,
				MaxGas:   9223372036854775807,
			},
			Evidence: &tmproto.EvidenceParams{
				MaxAgeNumBlocks: 302400,
				MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
				MaxBytes:        10000,
			},
			Validator: &tmproto.ValidatorParams{
				PubKeyTypes: []string{
					tmtypes.ABCIPubKeyTypeEd25519,
				},
			},
		},
	}
}

type PrototypeSuite struct {
	suite.Suite

	// simulate a relayed path
	simibc simibc.RelayedPath

	// keep around validators for easy access
	valAddresses []sdk.ValAddress

	// offsets: the model time and heights start at 0
	// so offsets are needed for comparisons.
	offsetTimeUnix int64
	offsetHeight   int64
}

type builder struct {
	suite       *suite.Suite
	link        simibc.OrderedLink
	path        *ibctesting.Path
	coordinator *ibctesting.Coordinator
}

func (b *builder) t() *testing.T { return b.suite.T() }

func GetZeroState(suite *suite.Suite, initState InitState) (*ibctesting.Path, []sdk.ValAddress, int64, int64) {
	b := builder{}
	// Get only the provider
	// coord, prov, cons := simapp.NewProviderConsumerCoordinator(suite.T())
	b.coordinator = simapp.NewBasicCoordinator(b.t())
	chainID := ibctesting.GetChainID(1)
	consChainID := ibctesting.GetChainID(2)
	coordinator.Chains[chainID] = ibctesting.NewTestChain(t, coordinator, simapp.SetupTestingappProvider, chainID)
	prov := coordinator.Chains[chainID]

	height := clienttypes.NewHeight(0, uint64(b.consumerChain().CurrentHeader.Height))
	proposal := providertypes.NewConsumerAdditionProposal("", "",
		consChainID,
		height,
		[]byte("a"),
		[]byte("a"),
		prov.CurrentHeader.Time.Add(-time.Hour)).(*providertypes.ConsumerAdditionProposal)

	b.providerKeeper().HandleConsumerAdditionProposal(b.ctx(P), proposal)
	consumerGenesis, found := b.providerKeeper().GetConsumerGenesis(b.ctx(P), b.chainID(C))

	// prov.NextBlock()

	_ = coord

	path := ibctesting.NewPath(prov, cons)
	addr := []sdk.ValAddress{}
	return path, addr, 0, 0
}

func (s *PrototypeSuite) TestAssumptions() {
	// test assumptions about the model
	s.Require().Equal(0, s.offsetHeight)
	s.Require().Equal(0, s.offsetTimeUnix)
}

// SetupTest sets up the test suite in a 'zero' state which matches
// the initial state in the model.
func (s *PrototypeSuite) SetupTest() {
	state := initState
	path, valAddresses, offsetHeight, offsetTimeUnix := GetZeroState(&s.Suite, state)
	s.valAddresses = valAddresses
	s.offsetHeight = offsetHeight
	s.offsetTimeUnix = offsetTimeUnix
	s.simibc = simibc.MakeRelayedPath(s.Suite.T(), path)
}

func TestPrototypeSuite(t *testing.T) {
	suite.Run(t, new(PrototypeSuite))
}
