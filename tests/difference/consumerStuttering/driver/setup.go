package consumerStuttering

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	mocks "github.com/cosmos/interchain-security/testutil/keeper"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/x/ccv/provider/types"
	providertypes "github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccvtypes "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

type Runner struct {
	t         *testing.T
	ctx       *sdk.Context
	k         *providerkeeper.Keeper
	sk        *SpecialStakingKeeper
	lastState State
}

func NewRunner(t *testing.T, initState State) *Runner {
	sk := NewSpecialStakingKeeper()

	pk, ctx := GetProviderKeeperAndCtx(t, sk)

	r := Runner{t, &ctx, &pk, sk, initState}
	_ = r

	return &r
}

func GetProviderKeeperAndCtx(t testing.TB,
	stakingKeeper *SpecialStakingKeeper) (providerkeeper.Keeper, sdk.Context) {

	cdc, storeKey, paramsSubspace, ctx := SetupInMemKeeper(t)

	fixParams(ctx, paramsSubspace)

	// ibcKeeper := GetIBCKeeper(stakingKeeper)

	k := providerkeeper.NewKeeper(
		cdc,
		storeKey,
		paramsSubspace,
		&mocks.MockScopedKeeper{},
		&MockChannelKeeper{},
		// ibcKeeper.ChannelKeeper,
		&MockPortKeeper{},
		// &ibcKeeper.PortKeeper,
		&MockConnectionKeeper{},
		// ibcKeeper.ConnectionKeeper,
		&MockClientKeeper{},
		// ibcKeeper.ClientKeeper,
		// &mocks.MockStakingKeeper{},
		stakingKeeper,
		&mocks.MockSlashingKeeper{},
		&mocks.MockAccountKeeper{},
		"",
	)
	return k, ctx
}

func SetupInMemKeeper(t testing.TB) (*codec.ProtoCodec, *storetypes.KVStoreKey, paramstypes.Subspace, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(ccvtypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(ccvtypes.MemStoreKey)

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

func fixParams(ctx sdk.Context, ss paramstypes.Subspace) paramstypes.Subspace {
	keyTable := paramstypes.NewKeyTable(paramstypes.NewParamSetPair(providertypes.KeyTemplateClient, &ibctmtypes.ClientState{}, func(value interface{}) error { return nil }))
	ss = ss.WithKeyTable(keyTable)

	expectedClientState :=
		ibctmtypes.NewClientState("", ibctmtypes.DefaultTrustLevel, 0, 0,
			time.Second*10, clienttypes.Height{}, commitmenttypes.GetSDKSpecs(), []string{"upgrade", "upgradedIBCState"}, true, true)

	ss.Set(ctx, types.KeyTemplateClient, expectedClientState)
	return ss
}
