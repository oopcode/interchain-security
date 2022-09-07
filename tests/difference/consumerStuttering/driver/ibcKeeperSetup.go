package consumerStuttering

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

func GetIBCKeeper(staking *SpecialStakingKeeper) *ibckeeper.Keeper {

	var cdc codec.BinaryCodec
	var key sdk.StoreKey
	var paramSpace paramtypes.Subspace
	// var stakingKeeper clienttypes.StakingKeeper
	var upgradeKeeper clienttypes.UpgradeKeeper
	var scopedKeeper capabilitykeeper.ScopedKeeper

	return ibckeeper.NewKeeper(cdc, key, paramSpace, staking, upgradeKeeper, scopedKeeper)
}
