package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

// Verify verifies the given signature (r, s) for the given hash and public key (x, y).
// It returns true if the signature is valid, false otherwise.
func Verify(hash []byte, r, s, x, y *big.Int) bool {
	curve := elliptic.P256()

	if !curve.IsOnCurve(x, y) || x == nil || y == nil {
		return false
	}

	publicKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	// Verify the signature with the public key,
	// then return true if it's valid, false otherwise
	return ecdsa.Verify(publicKey, hash, r, s)
}
