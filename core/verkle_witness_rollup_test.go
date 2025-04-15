// Copyright 2024 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
)

// addToHash takes a base hash and adds an offset to it (for calculating struct field slots)
func addToHash(hash common.Hash, offset byte) common.Hash {
	// Convert hash to big.Int and then add the offset
	hashInt := new(big.Int).SetBytes(hash[:])
	hashInt.Add(hashInt, big.NewInt(int64(offset)))
	// Return in the hash format
	result := common.Hash{}
	hashInt.FillBytes(result[:])
	return result
}

func verifyStructField(t *testing.T, statedb *state.StateDB, structBaseSlot common.Hash, offset byte, want common.Hash) {
	slot := addToHash(structBaseSlot, offset)
	have := statedb.GetState(params.L1OriginContractAddress, slot)
	if have != want {
		t.Errorf("struct field %d mismatch: got %v, want %v", offset, have, want)
	}
}

func TestL1OriginSource(t *testing.T) {
	// This test stores the L1 block information in the L1OriginSource contract
	// and reads it back from the statedb to verify the values.
	checkL1OriginSource := func(statedb *state.StateDB, isVerkle bool) {
		const maxStoredBlocks = 8192
		const totalBlocks = 9000

		statedb.SetNonce(params.L1OriginContractAddress, 1, tracing.NonceChangeUnspecified)
		statedb.SetCode(params.L1OriginContractAddress, params.L1OriginContractCode)

		// Store the L1 origin data for blocks 1 to totalBlocks
		for i := 1; i <= totalBlocks; i++ {
			header := &types.Header{
				ParentHash: common.Hash{byte(i % 256)},
				Number:     big.NewInt(int64(i)),
				Difficulty: new(big.Int),
			}

			// Create a mock L1 block info
			heightBytes := make([]byte, 8)
			binary.BigEndian.PutUint64(heightBytes, uint64(i))
			mockBlockHash := crypto.Keccak256Hash(heightBytes)

			// Generate other roots using different prefixes to ensure uniqueness
			stateRoot := crypto.Keccak256Hash(append([]byte("state"), heightBytes...))
			receiptRoot := crypto.Keccak256Hash(append([]byte("receipt"), heightBytes...))
			transactionRoot := crypto.Keccak256Hash(append([]byte("transaction"), heightBytes...))
			parentBeaconRoot := crypto.Keccak256Hash(append([]byte("beacon"), heightBytes...))

			l1OriginData := &L1OriginSource{
				blockHash:        mockBlockHash,
				stateRoot:        stateRoot,
				receiptRoot:      receiptRoot,
				transactionRoot:  transactionRoot,
				blockHeight:      big.NewInt(int64(i)),
				parentBeaconRoot: parentBeaconRoot,
			}

			// Set up chain config
			chainConfig := params.MergedTestChainConfig
			if isVerkle {
				chainConfig = testVerkleChainConfig
			}

			vmContext := NewEVMBlockContext(header, nil, new(common.Address))
			evm := vm.NewEVM(vmContext, statedb, chainConfig, vm.Config{})

			// Process the L1 block info, which should store it in the contract
			ProcessL1OriginBlockInfo(l1OriginData, evm)

			// Calculate storage slot for the buffer entry and verify the values
			bufferIndex := uint64(i) % maxStoredBlocks
			baseSlot := common.Hash{}
			indexBytes := make([]byte, 32)
			binary.BigEndian.PutUint64(indexBytes[24:], bufferIndex)
			structBaseSlot := crypto.Keccak256Hash(append(indexBytes, baseSlot[:]...))

			// Verify the values based on the same order as defined in the L1OriginSource struct
			verifyStructField(t, statedb, structBaseSlot, 0, mockBlockHash)
			verifyStructField(t, statedb, structBaseSlot, 1, parentBeaconRoot)
			verifyStructField(t, statedb, structBaseSlot, 2, stateRoot)
			verifyStructField(t, statedb, structBaseSlot, 3, receiptRoot)
			verifyStructField(t, statedb, structBaseSlot, 4, transactionRoot)
			verifyStructField(t, statedb, structBaseSlot, 5, common.BigToHash(big.NewInt(int64(i))))
		}

		// Calculate the number of blocks that will overwrite older blocks, if there are more than maxStoredBlocks
		overlapBlocks := 0
		if totalBlocks > maxStoredBlocks {
			overlapBlocks = totalBlocks - maxStoredBlocks
		}

		// Verify that blocks [1, overlapBlocks] are overwritten by blocks [maxStoredBlocks+1, totalBlocks]
		for i := 1; i <= overlapBlocks; i++ {
			bufferIndex := uint64(i) % maxStoredBlocks
			baseSlot := common.Hash{}
			indexBytes := make([]byte, 32)
			binary.BigEndian.PutUint64(indexBytes[24:], bufferIndex)
			structBaseSlot := crypto.Keccak256Hash(append(indexBytes, baseSlot[:]...))

			// Check block height to verify if its updated with the new block height
			verifyStructField(t, statedb, structBaseSlot, 5, common.BigToHash(big.NewInt(int64(maxStoredBlocks+i))))
		}

		// Verify the blocks in [overlapBlocks+1, maxStoredBlocks] remain unchanged
		for i := overlapBlocks + 1; i <= maxStoredBlocks && i <= totalBlocks; i++ {
			bufferIndex := uint64(i) % maxStoredBlocks
			baseSlot := common.Hash{}
			indexBytes := make([]byte, 32)
			binary.BigEndian.PutUint64(indexBytes[24:], bufferIndex)
			structBaseSlot := crypto.Keccak256Hash(append(indexBytes, baseSlot[:]...))

			// Check block height to verify if it is unchanged
			verifyStructField(t, statedb, structBaseSlot, 5, common.BigToHash(big.NewInt(int64(i))))
		}
	}

	t.Run("MPT", func(t *testing.T) {
		statedb, _ := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
		checkL1OriginSource(statedb, false)
	})

	t.Run("Verkle", func(t *testing.T) {
		db := rawdb.NewMemoryDatabase()
		cacheConfig := DefaultCacheConfigWithScheme(rawdb.PathScheme)
		cacheConfig.SnapshotLimit = 0
		triedb := triedb.NewDatabase(db, cacheConfig.triedbConfig(true))
		statedb, _ := state.New(types.EmptyVerkleHash, state.NewDatabase(triedb, nil))
		checkL1OriginSource(statedb, true)
	})
}
