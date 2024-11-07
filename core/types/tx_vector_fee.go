package types

import (
	"bytes"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

type VectorFeeTx struct {
	ChainID    *uint256.Int
	Nonce      uint64
	Gas        uint64
	To         common.Address
	Value      *uint256.Int
	Data       []byte
	AccessList AccessList

	GasTipCaps VectorFeeUint // a.k.a. maxPriorityFeePerGas for all 3 types: Execution, Blob and Calldata
	GasFeeCaps VectorFeeUint // a.k.a. maxFeePerGas for all 3 types: Execution, Blob and Calldata

	BlobHashes []common.Hash
	// A blob transaction can optionally contain blobs. This field must be set when BlobTx
	// is used to create a transaction for signing.
	Sidecar *BlobTxSidecar `rlp:"-"`

	// Signature values
	V *uint256.Int `json:"v" gencodec:"required"`
	R *uint256.Int `json:"r" gencodec:"required"`
	S *uint256.Int `json:"s" gencodec:"required"`
}

// accessors for innerTx. (satisfies TxData Interface)
func (tx *VectorFeeTx) txType() byte           { return DynamicFeeTxType }
func (tx *VectorFeeTx) chainID() *big.Int      { return tx.ChainID.ToBig() }
func (tx *VectorFeeTx) accessList() AccessList { return tx.AccessList }
func (tx *VectorFeeTx) data() []byte           { return tx.Data }
func (tx *VectorFeeTx) gas() uint64            { return tx.Gas }

func (tx *VectorFeeTx) gasLimits() VectorGasLimit {
	return VectorGasLimit{tx.Gas, tx.blobGas(), tx.calldataGas()}
}

func (tx *VectorFeeTx) gasTipCaps() VectorFeeBigint {
	feesAsBigInt := make(VectorFeeBigint, len(tx.GasTipCaps))
	for i, f := range tx.GasTipCaps {
		feesAsBigInt[i] = f.ToBig()
	}

	return feesAsBigInt
}

func (tx *VectorFeeTx) gasFeeCaps() VectorFeeBigint {
	feesAsBigInt := make(VectorFeeBigint, len(tx.GasFeeCaps))
	for i, f := range tx.GasFeeCaps {
		feesAsBigInt[i] = f.ToBig()
	}

	return feesAsBigInt
}

func (tx *VectorFeeTx) value() *big.Int     { return tx.Value.ToBig() }
func (tx *VectorFeeTx) nonce() uint64       { return tx.Nonce }
func (tx *VectorFeeTx) to() *common.Address { tmp := tx.To; return &tmp }

func (tx *VectorFeeTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V.ToBig(), tx.R.ToBig(), tx.S.ToBig()
}

func (tx *VectorFeeTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID.SetFromBig(chainID)
	tx.V.SetFromBig(v)
	tx.R.SetFromBig(r)
	tx.S.SetFromBig(s)
}

func (tx *VectorFeeTx) encode(b *bytes.Buffer) error {
	return rlp.Encode(b, tx)
}

func (tx *VectorFeeTx) decode(input []byte) error {
	return rlp.DecodeBytes(input, tx)
}

func (tx *VectorFeeTx) calldataGas() uint64 {
	zeroBytes := bytes.Count(tx.Data, []byte{0x00})
	nonZeroBytes := len(tx.Data) - zeroBytes
	tokens := uint64(zeroBytes) + uint64(nonZeroBytes)*params.CalldataTokensPerNonZeroByte

	return tokens * params.CalldataGasPerToken
}

func (tx *VectorFeeTx) blobGas() uint64 {
	return params.BlobTxBlobGasPerBlob * uint64(len(tx.BlobHashes))
}

// TODO: check if this is indeed proper implemenation
// NOTE: These methods are needed to satisfy TxData Interface
func (tx *VectorFeeTx) gasFeeCap() *big.Int                                       { return nil }
func (tx *VectorFeeTx) gasTipCap() *big.Int                                       { return nil }
func (tx *VectorFeeTx) gasPrice() *big.Int                                        { return nil }
func (tx *VectorFeeTx) effectiveGasPrice(dst *big.Int, baseFee *big.Int) *big.Int { return nil }

func (tx *VectorFeeTx) copy() TxData {
	cpy := &VectorFeeTx{
		Nonce: tx.Nonce,
		To:    tx.To,
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList: make(AccessList, len(tx.AccessList)),
		BlobHashes: make([]common.Hash, len(tx.BlobHashes)),
		Value:      new(uint256.Int),
		ChainID:    new(uint256.Int),
		GasTipCaps: VectorFeeUint{},
		GasFeeCaps: VectorFeeUint{},
		V:          new(uint256.Int),
		R:          new(uint256.Int),
		S:          new(uint256.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	copy(cpy.BlobHashes, tx.BlobHashes)

	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}

	for i, v := range tx.GasTipCaps {
		cpy.GasTipCaps[i].Set(v)
	}

	for i, v := range tx.GasFeeCaps {
		cpy.GasFeeCaps[i].Set(v)
	}

	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	if tx.Sidecar != nil {
		cpy.Sidecar = &BlobTxSidecar{
			Blobs:       append([]kzg4844.Blob(nil), tx.Sidecar.Blobs...),
			Commitments: append([]kzg4844.Commitment(nil), tx.Sidecar.Commitments...),
			Proofs:      append([]kzg4844.Proof(nil), tx.Sidecar.Proofs...),
		}
	}

	return cpy
}
