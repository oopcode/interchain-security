package consumerStuttering

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

//// Temporary below here (scratch code)

type SpecialStakingKeeper struct {
	// Controlled by this
	nextOpId uint64
	// Unbonding op id to reference count
	// Initialised by this, modified by staking module
	RefCnt map[uint64]int
	// Initialised by this, modified by staking module
	VscIdToOpIds map[uint64][]uint64
}

func NewSpecialStakingKeeper() *SpecialStakingKeeper {
	return &SpecialStakingKeeper{
		0,
		map[uint64]int{},
		map[uint64][]uint64{},
	}
}

func (k *SpecialStakingKeeper) AddUnbondingOperation(vscId uint64, callback func(uint64)) {
	if _, ok := k.VscIdToOpIds[vscId]; !ok {
		//do something here
		k.VscIdToOpIds[vscId] = []uint64{}
	}
	k.VscIdToOpIds[vscId] = append(k.VscIdToOpIds[vscId], k.nextOpId)
	k.RefCnt[k.nextOpId] = 0
	callback(k.nextOpId)
	k.nextOpId += 1
}

func (k *SpecialStakingKeeper) GetValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
func (k *SpecialStakingKeeper) UnbondingCanComplete(ctx sdk.Context, id uint64) error {
	k.RefCnt[id] -= 1
	return nil
}
func (k *SpecialStakingKeeper) UnbondingTime(ctx sdk.Context) time.Duration {
	return 0
}
func (k *SpecialStakingKeeper) GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool) {
	return stakingtypes.Validator{}, false
}
func (k *SpecialStakingKeeper) Jail(sdk.Context, sdk.ConsAddress) {

}
func (k *SpecialStakingKeeper) Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec, stakingtypes.InfractionType) {

}
func (k *SpecialStakingKeeper) GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool) {
	return stakingtypes.Validator{}, false
}
func (k *SpecialStakingKeeper) IterateLastValidatorPowers(ctx sdk.Context, cb func(addr sdk.ValAddress, power int64) (stop bool)) {

}
func (k *SpecialStakingKeeper) PowerReduction(ctx sdk.Context) sdk.Int {
	return sdk.ZeroInt()
}
func (k *SpecialStakingKeeper) PutUnbondingOnHold(ctx sdk.Context, id uint64) error {
	k.RefCnt[id] += 1
	return nil
}
