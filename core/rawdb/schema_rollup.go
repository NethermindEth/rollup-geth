package rawdb

import "github.com/ethereum/go-ethereum/common"

var headerBaseFeesPrefix = []byte("hb") // headerBaseFeesPrefix + hash -> RLP(types.VectorFeeBigInt)

// headerBaseFeesKey = headerBaseFeesPrefix + hash
func headerBaseFeesKey(hash common.Hash) []byte {
	return append(headerBaseFeesPrefix, hash.Bytes()...)
}
