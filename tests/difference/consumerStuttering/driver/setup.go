package consumerStuttering

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	appProvider "github.com/cosmos/interchain-security/app/provider"
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/x/ccv/provider/types"
	providertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/tendermint/spm/cosmoscmd"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
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

	// TODO: perhaps I can mimic the keeper creation in app.Go, only swapping out the staking keeper
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

func Alternative(t testing.TB,
	stakingKeeper ccv.StakingKeeper) (providerkeeper.Keeper, sdk.Context) {

	chainID := "testchain0"

	app, genesis := createTestingApp()

	// TODO: need something inbetween?
	stateBytes, _ := json.MarshalIndent(genesis, "", " ")

	app.InitChain(abci.RequestInitChain{
		ChainId:         chainID,
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: consensusParams(),
		AppStateBytes:   stateBytes,
	})

	app.Commit()

	h := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		// TODO: this is taken from testing/coordinator.go
		Time: time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC).UTC(),
	}
	ctx := app.GetBaseApp().NewContext(false, h)
	return app.ProviderKeeper, ctx
}

func createTestingApp() (*appProvider.App, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	encoding := cosmoscmd.MakeEncodingConfig(appProvider.ModuleBasics)
	app := appProvider.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 5, encoding, simapp.EmptyAppOptions{}).(*appProvider.App)

	k := providerkeeper.NewKeeper(
		app.AppCodec(),
		sdk.NewKVStoreKey(providertypes.StoreKey),
		app.GetSubspace(providertypes.ModuleName),
		app.ScopedIBCProviderKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.IBCKeeper.ConnectionKeeper,
		app.IBCKeeper.ClientKeeper,
		NewSpecialStakingKeeper(),
		app.SlashingKeeper,
		app.AccountKeeper,
		authtypes.FeeCollectorName,
	)

	app.ProviderKeeper = k

	return app, appProvider.NewDefaultGenesisState(encoding.Marshaler)

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
