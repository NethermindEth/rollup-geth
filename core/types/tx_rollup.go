// Contains rollup-specific implementations for tx_legacy, tx_access_list, tx_dynamic_fee
// and tx_blob
package types

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func (tx *LegacyTx) calldataGas() uint64 {
	zeroBytes := bytes.Count(tx.Data, []byte{0x00})
	nonZeroBytes := len(tx.Data) - zeroBytes
	tokens := uint64(zeroBytes) + uint64(nonZeroBytes)*params.CalldataTokensPerNonZeroByte

	return tokens * params.CalldataGasPerToken
}

func (tx *LegacyTx) gasLimits() VectorGasLimit {
	return VectorGasLimit{tx.Gas, 0, tx.calldataGas()}
}

func (tx *LegacyTx) gasTipCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasPrice, big.NewInt(0), tx.GasPrice}
}

func (tx *LegacyTx) gasFeeCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasPrice, big.NewInt(0), tx.GasPrice}
}

func (tx *AccessListTx) calldataGas() uint64 {
	zeroBytes := bytes.Count(tx.Data, []byte{0x00})
	nonZeroBytes := len(tx.Data) - zeroBytes
	tokens := uint64(zeroBytes) + uint64(nonZeroBytes)*params.CalldataTokensPerNonZeroByte

	return tokens * params.CalldataGasPerToken
}

func (tx *AccessListTx) gasLimits() VectorGasLimit {
	return VectorGasLimit{tx.Gas, 0, tx.calldataGas()}
}

func (tx *AccessListTx) gasTipCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasPrice, big.NewInt(0), tx.GasPrice}
}

func (tx *AccessListTx) gasFeeCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasPrice, big.NewInt(0), tx.GasPrice}
}

func (tx *DynamicFeeTx) calldataGas() uint64 {
	zeroBytes := bytes.Count(tx.Data, []byte{0x00})
	nonZeroBytes := len(tx.Data) - zeroBytes
	tokens := uint64(zeroBytes) + uint64(nonZeroBytes)*params.CalldataTokensPerNonZeroByte

	return tokens * params.CalldataGasPerToken
}

func (tx *DynamicFeeTx) gasLimits() VectorGasLimit {
	return VectorGasLimit{tx.Gas, 0, tx.calldataGas()}
}

func (tx *DynamicFeeTx) gasTipCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasTipCap, big.NewInt(0), tx.GasTipCap}
}

func (tx *DynamicFeeTx) gasFeeCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasFeeCap, big.NewInt(0), tx.GasFeeCap}
}

func (tx *BlobTx) calldataGas() uint64 {
	zeroBytes := bytes.Count(tx.Data, []byte{0x00})
	nonZeroBytes := len(tx.Data) - zeroBytes
	tokens := uint64(zeroBytes) + uint64(nonZeroBytes)*params.CalldataTokensPerNonZeroByte

	return tokens * params.CalldataGasPerToken
}

func (tx *BlobTx) gasLimits() VectorGasLimit {
	return VectorGasLimit{tx.Gas, tx.blobGas(), tx.calldataGas()}
}

func (tx *BlobTx) gasTipCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasTipCap.ToBig(), big.NewInt(0), tx.GasTipCap.ToBig()}
}

func (tx *BlobTx) gasFeeCaps() VectorFeeBigint {
	return VectorFeeBigint{tx.GasFeeCap.ToBig(), tx.BlobFeeCap.ToBig(), tx.GasFeeCap.ToBig()}
}
