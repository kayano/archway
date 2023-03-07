package keeper

import (
	"time"

	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const Year = 24 * time.Hour * 365

// GetBlockProvisions gets the tokens to be minted in the current block and returns the new inflation amount as well
func (k Keeper) GetBlockProvisions(ctx sdk.Context) (tokens sdk.Coin, blockInflation sdk.Dec) {
	mintParams := k.GetParams(ctx)

	// getting last block info
	lbi, found := k.GetLastBlockInfo(ctx)
	if !found {
		currentTime := ctx.BlockTime()
		lbi = types.LastBlockInfo{
			Inflation: mintParams.MinInflation,
			Time:      &currentTime,
		}
	}

	// time since last mint
	elapsed := ctx.BlockTime().Sub(*lbi.GetTime())
	if elapsed > mintParams.GetMaxBlockDuration() {
		elapsed = mintParams.GetMaxBlockDuration()
	}

	// inflation for the current block
	bondedRatio := k.BondedRatio(ctx)
	bondDenom := k.BondDenom(ctx)
	blockInflation = getBlockInflation(lbi.Inflation, bondedRatio, mintParams, elapsed)

	// amount of bond tokens to mint in this block
	bondedTokenSupply := k.GetBondedTokenSupply(ctx)
	tokenAmount := blockInflation.MulInt(bondedTokenSupply.Amount).MulInt64(int64(elapsed / Year)) // amount := (inflation * bondedTokenSupply) * (elapsed/Year)
	tokens = sdk.NewInt64Coin(bondDenom, tokenAmount.BigInt().Int64())                             // as sdk.Coin
	return
}

// getBlockInflation adjusts the current block inflation amount based on tokens bonded
func getBlockInflation(inflation sdk.Dec, bondedRatio sdk.Dec, mintParams types.Params, elapsed time.Duration) sdk.Dec {
	switch {
	case bondedRatio.LT(mintParams.MinBonded): // if bondRatio is lower than we want, increase inflation
		inflation = inflation.Add(mintParams.InflationChange.MulInt64(int64(elapsed)))
	case bondedRatio.GT(mintParams.MaxBonded): // if bondRatio is higher than we want, decrease inflation
		inflation = inflation.Sub(mintParams.InflationChange.MulInt64(int64(elapsed)))
	}
	if inflation.GT(mintParams.MaxInflation) {
		inflation = mintParams.MaxInflation
	} else if inflation.LT(mintParams.MinInflation) {
		inflation = mintParams.MinInflation
	}
	return inflation
}
