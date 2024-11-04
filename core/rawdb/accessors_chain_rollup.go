package rawdb

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
)

// ReadHeaderBaseFees reads the base fees for the given header (hash)
func ReadHeaderBaseFees(db ethdb.KeyValueReader, hash common.Hash) *types.VectorFeeBigint {
	data, _ := db.Get(headerBaseFeesKey(hash))
	if len(data) == 0 {
		return nil
	}
	dec := new(types.VectorFeeBigint)
	err := rlp.DecodeBytes(data, dec)
	if err != nil {
		return nil
	}

	return dec
}

// WriteHeaderBaseFee stores the header hash->base fees mapping.
func WriteHeaderBaseFees(db ethdb.KeyValueWriter, hash common.Hash, baseFees *types.VectorFeeBigint) {
	if baseFees == nil {
		//Pre-EIP-7706 blocks don't have base fees so no need to log errors
		return
	}

	key := headerBaseFeesKey(hash)
	enc, err := rlp.EncodeToBytes(baseFees)
	if err != nil {
		log.Crit("Failed to RLP encode Base Fees", "err", err)
	}

	if err = db.Put(key, enc); err != nil {
		log.Crit("Failed to store header hash to base fees mapping", "err", err)
	}
}

// DeleteHeaderBaseFees removes header hash -> base fees mapping
func DeleteHeaderBaseFees(db ethdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(headerBaseFeesKey(hash)); err != nil {
		log.Crit("Failed to delete header hash to base fees mapping", "err", err)
	}
}
