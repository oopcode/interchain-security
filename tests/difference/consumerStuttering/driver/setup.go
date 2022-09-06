package consumerStuttering

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	appProvider "github.com/cosmos/interchain-security/app/provider"
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/tendermint/spm/cosmoscmd"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	// simapp "github.com/cosmos/interchain-security/testutil/simapp"
	simapp "github.com/cosmos/cosmos-sdk/simapp"
)

type Runner struct {
	t         *testing.T
	ctx       *sdk.Context
	am        provider.AppModule
	k         *providerkeeper.Keeper
	sk        *SpecialStakingKeeper
	lastState State
}

func fixParamSubspace(ctx sdk.Context, ss paramstypes.Subspace) paramstypes.Subspace {
	pair := paramstypes.NewParamSetPair(types.KeyTemplateClient, &ibctmtypes.ClientState{}, func(value interface{}) error { return nil })
	keyTable := paramstypes.NewKeyTable(pair)
	ss = ss.WithKeyTable(keyTable)

	expectedClientState :=
		ibctmtypes.NewClientState("", ibctmtypes.DefaultTrustLevel, 0, 0,
			time.Second*10, clienttypes.Height{}, commitmenttypes.GetSDKSpecs(), []string{"upgrade", "upgradedIBCState"}, true, true)

	ss.Set(ctx, types.KeyTemplateClient, expectedClientState)

	return ss
}

func GetProviderKeeperAndCtx(t testing.TB, stakingKeeper ccv.StakingKeeper) (providerkeeper.Keeper, sdk.Context) {

	cdc, storeKey, paramsSubspace, ctx := testkeeper.SetupInMemKeeper(t)

	paramsSubspace = fixParamSubspace(ctx, paramsSubspace)

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
	// providerKeeper, ctx := GetProviderKeeperAndCtx(t, stakingKeeper)
	providerKeeper, ctx := Alternative(t, stakingKeeper)

	r := Runner{t, &ctx, provider.NewAppModule(&providerKeeper), &providerKeeper, stakingKeeper, initState}
	_ = r

	return &r
}

func SetupApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	// encCdc := app.MakeTestEncodingConfig()
	encoding := cosmoscmd.MakeEncodingConfig(appProvider.ModuleBasics)
	// TODO: figure out a way to set a mock staking module
	testApp := appProvider.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 5, encoding, simapp.EmptyAppOptions{}).(abci.Application)
	return testApp, appProvider.NewDefaultGenesisState(encoding.Marshaler)
}

func Alternative(t *testing.T, stakingKeeper ccv.StakingKeeper) (providerkeeper.Keeper, sdk.Context) {

	appInit := SetupApp

	app, genesis := appInit()

	app.GetStakingKeeper()

	chainID := ibctesting.GetChainID(0)

	// TODO: I need to put the mocked staking keeper in here somehow
	stateBytes := getAppBytes(chainID, app, genesis)

	app.InitChain(
		abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: consensusParams(),
			AppStateBytes:   stateBytes,
		},
	)

	app.Commit()

	h := tmproto.Header{
		ChainID:            chainID,
		Height:             app.LastBlockHeight() + 1,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     []byte{},
		NextValidatorsHash: []byte{},
		// ValidatorsHash:     validators.Hash(),
		// NextValidatorsHash: validators.Hash(),
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: h})

	k := app.(*appProvider.App).ProviderKeeper
	ctx := app.GetBaseApp().NewContext(false, h)
	// ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	return k, ctx
}

/*
	tmtypes "github.com/tendermint/tendermint/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
*/

func getAppBytes(chainID string, app ibctesting.TestingApp,
	genesis map[string]json.RawMessage) []byte {
	stateBytes, _ := json.MarshalIndent(genesis, "", " ")
	return stateBytes

}

func consensusParams() *abci.ConsensusParams {
	return &abci.ConsensusParams{
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
	}
}
