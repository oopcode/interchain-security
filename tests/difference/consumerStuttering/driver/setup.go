package consumerStuttering

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
)

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
