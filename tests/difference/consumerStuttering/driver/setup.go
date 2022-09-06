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
	testkeeper "github.com/cosmos/interchain-security/testutil/keeper"
	"github.com/cosmos/interchain-security/testutil/simapp"
	provider "github.com/cosmos/interchain-security/x/ccv/provider"
	providerkeeper "github.com/cosmos/interchain-security/x/ccv/provider/keeper"
	"github.com/cosmos/interchain-security/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/x/ccv/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
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

func Alternative(t *testing.T, stakingKeeper ccv.StakingKeeper) (providerkeeper.Keeper, sdk.Context) {

	appInit := simapp.SetupTestingappProvider

	app, genesis := appInit()

	chainID := ibctesting.GetChainID(0)

	stateBytes := getAppBytesAndSenders(chainID, app, genesis)

	app.InitChain(
		abci.RequestInitChain{
			ChainId:         chainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: consensusParams(),
			AppStateBytes:   stateBytes,
		},
	)
}

func getAppBytesAndSenders(chainID string, app ibctesting.TestingApp,
	genesis map[string]json.RawMessage) []byte {

	accounts := []authtypes.GenesisAccount{}
	balances := []banktypes.Balance{}
	senderAccounts := []ibctesting.SenderAccount{}

	// Create genesis accounts.
	for i := 0; i < 2; i++ {
		pk := secp256k1.GenPrivKey()
		acc := authtypes.NewBaseAccount(pk.PubKey().Address().Bytes(), pk.PubKey(), uint64(i), 0)

		// Give enough funds for many delegations
		// Extra units are to delegate to extra validators created later
		// in order to bond them and still have INITIAL_DELEGATOR_TOKENS remaining
		extra := 0
		for j := 0; j < b.initState.NumValidators; j++ {
			if b.initState.ValStates.Status[j] != stakingtypes.Bonded {
				extra += b.initState.ValStates.Delegation[j]
			}
		}
		amt := uint64(b.initState.InitialDelegatorTokens + extra)

		bal := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(amt))),
		}

		accounts = append(accounts, acc)
		balances = append(balances, bal)

		senderAccount := ibctesting.SenderAccount{
			SenderAccount: acc,
			SenderPrivKey: pk,
		}

		senderAccounts = append(senderAccounts, senderAccount)
	}

	// set genesis accounts
	genesisAuth := authtypes.NewGenesisState(authtypes.DefaultParams(), accounts)
	genesis[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(genesisAuth)

	stakingValidators := make([]stakingtypes.Validator, 0, len(validators.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(validators.Validators))

	// Sum bonded is needed for BondedPool account
	sumBonded := sdk.NewInt(0)

	for i, val := range validators.Validators {
		status := b.initState.ValStates.Status[i]
		delegation := b.initState.ValStates.Delegation[i]
		extra := b.initState.ValStates.ValidatorExtraTokens[i]

		tokens := sdk.NewInt(int64(delegation + extra))
		b.suite.Require().Equal(status, stakingtypes.Bonded, "All genesis validators should be bonded")
		sumBonded = sumBonded.Add(tokens)
		// delegator account receives delShares shares
		delShares := sdk.NewDec(int64(delegation))
		// validator has additional sumShares due to extra units
		sumShares := sdk.NewDec(int64(delegation + extra))

		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		require.NoError(b.suite.T(), err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(b.suite.T(), err)

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            status,
			Tokens:            tokens,
			DelegatorShares:   sumShares,
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}

		stakingValidators = append(stakingValidators, validator)

		// Store delegation from the model delegator account
		delegations = append(delegations, stakingtypes.NewDelegation(accounts[0].GetAddress(), val.Address.Bytes(), delShares))
		// Remaining delegation is from extra account
		delegations = append(delegations, stakingtypes.NewDelegation(accounts[1].GetAddress(), val.Address.Bytes(), sumShares.Sub(delShares)))
	}

	bondDenom := sdk.DefaultBondDenom
	genesisStaking := stakingtypes.GenesisState{}

	if genesis[stakingtypes.ModuleName] != nil {
		// If staking module genesis already exists
		app.AppCodec().MustUnmarshalJSON(genesis[stakingtypes.ModuleName], &genesisStaking)
		bondDenom = genesisStaking.Params.BondDenom
	}

	// Set model parameters
	genesisStaking.Params.MaxEntries = uint32(b.initState.MaxEntries)
	genesisStaking.Params.MaxValidators = uint32(b.initState.MaxValidators)
	genesisStaking.Params.UnbondingTime = b.initState.UnbondingP
	genesisStaking = *stakingtypes.NewGenesisState(genesisStaking.Params, stakingValidators, delegations)
	genesis[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(&genesisStaking)

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(bondDenom, sumBonded)},
	})

	// add unbonded amount
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(bondDenom, sdk.ZeroInt())},
	})

	// update total funds supply
	genesisBank := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, sdk.NewCoins(), []banktypes.Metadata{})
	genesis[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(genesisBank)

	stateBytes, err := json.MarshalIndent(genesis, "", " ")
	require.NoError(b.suite.T(), err)

	return stateBytes, senderAccounts

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
