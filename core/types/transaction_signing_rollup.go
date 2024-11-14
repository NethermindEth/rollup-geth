package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type eip7706Signer struct{ cancunSigner }

// NewCancunSigner returns a signer that accepts
// - EIP-77706 vector fee transactions
// - EIP-4844 blob transactions
// - EIP-1559 dynamic fee transactions
// - EIP-2930 access list transactions,
// - EIP-155 replay protected transactions, and
// - legacy Homestead transactions.
func NewEIP7706Signer(chainId *big.Int) Signer {
	return eip7706Signer{cancunSigner{londonSigner{eip2930Signer{NewEIP155Signer(chainId)}}}}
}

// Sender returns the sender address of the transaction.
func (s eip7706Signer) Sender(tx *Transaction) (common.Address, error) {
	if tx.Type() != VectorFeeTxType {
		return s.cancunSigner.Sender(tx)
	}

	V, R, S := tx.RawSignatureValues()
	// Blob txs are defined to use 0 and 1 as their recovery
	// id, add 27 to become equivalent to unprotected Homestead signatures.
	V = new(big.Int).Add(V, big.NewInt(27))
	if tx.ChainId().Cmp(s.chainId) != 0 {
		return common.Address{}, fmt.Errorf("%w: have %d want %d", ErrInvalidChainId, tx.ChainId(), s.chainId)
	}
	return recoverPlain(s.Hash(tx), R, S, V, true)
}

// SignatureValues returns the raw R, S, V values corresponding to the
// given signature.
func (signer eip7706Signer) SignatureValues(tx *Transaction, sig []byte) (R, S, V *big.Int, err error) {
	txData, ok := tx.inner.(*VectorFeeTx)
	if !ok {
		return signer.cancunSigner.SignatureValues(tx, sig)
	}
	// Check that chain ID of tx matches the signer. We also accept ID zero here,
	// because it indicates that the chain ID was not specified in the tx.
	if txData.ChainID.Sign() != 0 && txData.ChainID.ToBig().Cmp(signer.chainId) != 0 {
		return nil, nil, nil, fmt.Errorf("%w: have %d want %d", ErrInvalidChainId, txData.ChainID, signer.chainId)
	}

	R, S, _ = decodeSignature(sig)
	V = big.NewInt(int64(sig[64]))
	return R, S, V, nil
}

// Returns chain id used by signer
func (s eip7706Signer) ChainID() *big.Int {
	return s.chainId
}

// TODO: [rollup-geth] Question: I guess all clients must agree to follow this exact structure otherwise signatures will be off?

// Hash returns 'signature hash', i.e. the transaction hash that is signed by the
// private key. This hash does not uniquely identify the transaction.
func (s eip7706Signer) Hash(tx *Transaction) common.Hash {
	if tx.Type() != VectorFeeTxType {
		return s.cancunSigner.Hash(tx)
	}
	return prefixedRlpHash(
		tx.Type(),
		[]interface{}{
			s.chainId,
			tx.Nonce(),
			tx.GasTipCaps(),
			tx.GasFeeCaps(),
			tx.Gas(),
			tx.To(),
			tx.Value(),
			tx.Data(),
			tx.AccessList(),
			tx.BlobGasFeeCap(),
			tx.BlobHashes(),
		})
}

// Equal returns true if the given signer is the same as the receiver.
func (s eip7706Signer) Equal(s2 Signer) bool {
	x, ok := s2.(eip7706Signer)
	return ok && x.chainId.Cmp(s.chainId) == 0
}
