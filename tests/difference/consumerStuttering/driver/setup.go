package consumerStuttering

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	clientkeeper "github.com/cosmos/ibc-go/v3/modules/core/02-client/keeper"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	connectionkeeper "github.com/cosmos/ibc-go/v3/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v3/modules/core/04-channel/keeper"
	portkeeper "github.com/cosmos/ibc-go/v3/modules/core/05-port/keeper"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spm/cosmoscmd"
	"github.com/tendermint/spm/ibckeeper"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

	cdc, storeKey, paramsSubspace, ctx := testkeeper.SetupInMemKeeper(t)

	// paramsSubspace = fixParamSubspace(ctx, paramsSubspace)

	upgradeKeeper := upgradekeeper.NewKeeper(
		map[int64]bool{},
		sdk.NewKVStoreKey(upgradetypes.StoreKey),
		cdc,
		"", //TODO:
		app.BaseApp,
	)

	ibcParamsSpace := paramstypes.NewSubspace(cdc,
		cdc,
		ibckey,
		memStoreKey,
		paramstypes.ModuleName,
	)

	clientKeeper := clientkeeper.NewKeeper(cdc, ibckey, paramSpace, stakingKeeper, upgradeKeeper)
	connectionKeeper := connectionkeeper.NewKeeper(cdc, ibckey, paramSpace, clientKeeper)
	portKeeper := portkeeper.NewKeeper(scopedKeeper)
	channelKeeper := channelkeeper.NewKeeper(cdc, ibckey, clientKeeper, connectionKeeper, portKeeper, scopedKeeper)

	ibcKeeper := ibckeeper.NewKeeper(
		cdc,
		ibckey,
		app.GetSubspace(ibchost.ModuleName),
		stakingKeeper,
		upgradeKeeper,
		scopedIBCKeeper,
	)

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

func SetupInMemIbcKeeper(t testing.TB) (*codec.ProtoCodec, *storetypes.KVStoreKey, paramstypes.Subspace, sdk.Context) {
	// TODO: need to reuse context here
	storeKey := sdk.NewKVStoreKey(ibchost.StoreKey)
	memStoreKeyS := "mem_ibc"
	memStoreKey := storetypes.NewMemoryStoreKey(memStoreKeyS)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := paramstypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		memStoreKey,
		paramstypes.ModuleName,
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())
	return cdc, storeKey, paramsSubspace, ctx
}

func SetupTestingappProvider() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := tmdb.NewMemDB()
	// encCdc := app.MakeTestEncodingConfig()
	encoding := cosmoscmd.MakeEncodingConfig(appProvider.ModuleBasics)
	testApp := appProvider.New(log.NewNopLogger(), db, nil, true, map[int64]bool{}, simapp.DefaultNodeHome, 5, encoding, simapp.EmptyAppOptions{}).(ibctesting.TestingApp)
	return testApp, appProvider.NewDefaultGenesisState(encoding.Marshaler)
}
