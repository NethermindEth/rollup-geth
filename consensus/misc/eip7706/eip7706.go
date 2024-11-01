package eip7706

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	minTxGasPrice          = big.NewInt(params.TxMinGasPrice)
	gaspriceUpdateFraction = big.NewInt(params.BaseFeeUpdateFraction)
)

// VerifyEIP7706Header Verifies EIP-7706 header
func VerifyEIP7706Header(parent, header *types.Header) error {
	// Verify the header is not malformed
	if header.ExcessGas == nil {
		return errors.New("header is missing excessGas")
	}
	if header.GasUsedVector == nil {
		return errors.New("header is missing gasUsedVector")
	}
	if header.GasLimits == nil {
		return errors.New("header is missing gasLimits")
	}

	// Verify blob gas usage
	blobGasUsed := header.GasUsedVector[1]
	if blobGasUsed > params.MaxBlobGasPerBlock {
		return fmt.Errorf("blob gas used %d exceeds maximum allowance %d", blobGasUsed, params.MaxBlobGasPerBlock)
	}

	if blobGasUsed%params.BlobTxBlobGasPerBlob != 0 {
		return fmt.Errorf("blob gas used %d not a multiple of blob gas per blob %d", blobGasUsed, params.BlobTxBlobGasPerBlob)
	}

	// Verify calldata gas usage
	calldataGasUsed := header.GasUsedVector[2]
	if calldataGasUsed%params.CalldataGasPerToken != 0 {
		return fmt.Errorf("calldata gas used %d not a multiple of calldata gas per token %d", calldataGasUsed, params.CalldataGasPerToken)
	}

	// Verify the excessGas is correct based on the parent header
	var (
		parentGasUsed   types.VectorGasLimit
		parentExcessGas types.VectorGasLimit
		parentGasLimits types.VectorGasLimit
	)

	if parentIsEIP7706Block := parent.GasUsedVector != nil && parent.ExcessGas != nil && parent.GasLimits != nil; parentIsEIP7706Block {
		parentGasUsed = *parent.GasUsedVector
		parentExcessGas = *parent.ExcessGas
		parentGasLimits = *parent.GasLimits
	} else {
		parentGasLimits = types.VectorGasLimit{parent.GasLimit, params.MaxBlobGasPerBlock, parent.GasLimit / params.CallDataGasLimitRatio}
		parentGasUsed = types.VectorGasLimit{parent.GasUsed, *parent.BlobGasUsed, parent.GasUsed / params.CallDataGasLimitRatio}
		// TODO: what about[rollup-geth] EIP-7706 execution excess gas and calldata excess gas for non EIP-7706 parent?
		parentExcessGas = types.VectorGasLimit{0, *parent.ExcessBlobGas, 0}
	}

	expectedExcessGas := CalcExecGas(parentGasUsed, parentExcessGas, parentGasLimits)
	if !header.ExcessGas.VectorAllEq(expectedExcessGas) {
		return errors.New("invalid excessGas")
	}

	// Verify base fees are correct based on the parent header
	// NOTE: EIP-7706 doesn't specify base fees as part of header (they are not part of header protocol spec)
	// But for convenience we do calculate them and store as part of the header
	// So, if they do exist, make sure they are properly calculated
	if header.BaseFees != nil {
		expectedBaseFees := CalcBaseFees(parentExcessGas, parentGasLimits)
		if !header.BaseFees.VectorAllEq(expectedBaseFees) {
			return errors.New("invalid baseFee")
		}
	}

	return nil
}

// CalcExecGas calculates excess gas given parent block gas usage and target parent target gas usage
func CalcExecGas(parentGasUsed, parentExecGas, parentGasLimits types.VectorGasLimit) types.VectorGasLimit {
	excessGas := parentExecGas.VectorAdd(parentGasUsed)

	return excessGas.VectorSubtractClampAtZero(getBlockTargets(parentGasLimits))
}

// CalcBaseFees  calculates vector of the base fees for current block header given parent excess gas and targets
func CalcBaseFees(parentExecGas, parentGasLimits types.VectorGasLimit) types.VectorFeeBigint {
	baseFees := types.NewVectorFeeBigInt()
	targets := getBlockTargets(parentGasLimits)

	for i, execGas := range parentExecGas {
		target := big.NewInt(int64(targets[i]))
		target = target.Mul(target, gaspriceUpdateFraction)
		baseFees[i] = fakeExponential(minTxGasPrice, big.NewInt(int64(execGas)), target)
	}

	return baseFees
}

func getBlockTargets(parentGasLimits types.VectorGasLimit) types.VectorGasLimit {
	var targets types.VectorGasLimit
	for i, limit := range parentGasLimits {
		if params.LimitTargetRatios[i] == 0 {
			targets[i] = 0
			continue
		}
		targets[i] = limit / params.LimitTargetRatios[i]
	}

	return targets
}

// fakeExponential approximates factor * e ** (numerator / denominator) using
// Taylor expansion.
func fakeExponential(factor, numerator, denominator *big.Int) *big.Int {
	var (
		output = new(big.Int)
		accum  = new(big.Int).Mul(factor, denominator)
	)
	for i := 1; accum.Sign() > 0; i++ {
		output.Add(output, accum)

		accum.Mul(accum, numerator)
		accum.Div(accum, denominator)
		accum.Div(accum, big.NewInt(int64(i)))
	}
	return output.Div(output, denominator)
}
