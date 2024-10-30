package eip7706

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	minTxGasPrice          = big.NewInt(params.TxMinGasPrice)
	gaspriceUpdateFraction = big.NewInt(params.BaseFeeUpdateFraction)
)

// TODO: [rollup-geth] implement this
// VerifyEIP7706Header Verifies EIP-7706 header
func VerifyEIP7706Header(parent, header *types.Header) error {
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
