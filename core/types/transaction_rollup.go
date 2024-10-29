package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
)

// Gas returns the gas limit of the transaction for each of [execution, blob, calldata] gas "type" respectively .
func (tx *Transaction) GasLimits() VectorGasLimit { return tx.inner.gasLimits() }

// GasTipCaps returns the vector of tip caps per gas for each of [execution, blob, calldata] gas "types" respectively
func (tx *Transaction) GasTipCaps() VectorFeeBigint { return tx.inner.gasTipCaps() }

// GasFeeCaps returns the vector of fee caps per gas  for each of [execution, blob, calldata] gas "types" respectively
func (tx *Transaction) GasFeeCaps() VectorFeeBigint { return tx.inner.gasFeeCaps() }

// EffectiveGasTip returns the effective miner gasTipCap for the given base fees, per gas for each of [execution, blob, calldata] gas "type" respectively .
func (tx *Transaction) EffectiveGasTips(baseFees VectorFeeBigint) VectorFeeBigint {
	gasFeeCaps := tx.GasFeeCaps()
	gasTipCaps := tx.GasTipCaps()
	effectiveTips := NewVectorFeeBigInt()
	for i, baseFee := range baseFees {
		if baseFee == nil {
			effectiveTips[i].Set(gasTipCaps[i])
		} else {
			effectiveTips[i] = math.BigMin(gasTipCaps[i], effectiveTips[i].Sub(gasFeeCaps[i], baseFee))
		}
	}

	return effectiveTips
}

// EffectiveGasPrices returns the effective (actual) prices per gas for each of [execution, blob, calldata] gas "type" respectively .
func (tx *Transaction) EffectiveGasPrices(baseFees VectorFeeBigint) VectorFeeBigint {
	gasFeeCaps := tx.GasFeeCaps()
	gasTipCaps := tx.GasTipCaps()
	effectiveFees := NewVectorFeeBigInt()

	for i, baseFee := range baseFees {
		if baseFee == nil {
			effectiveFees[i].Set(gasFeeCaps[i])
		} else {
			effectiveFees[i] = math.BigMin(effectiveFees[i].Add(gasTipCaps[i], baseFee), gasFeeCaps[i])
		}
	}

	return effectiveFees
}

// EffectiveGasPrice returns the effective (actual) price per gas
func (tx *Transaction) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return tx.inner.effectiveGasPrice(new(big.Int), baseFee)
}

func (tx *Transaction) EffectiveGasExecutionPrice(baseFees VectorFeeBigint) *big.Int {
	return tx.EffectiveGasPrices(baseFees)[0]
}
