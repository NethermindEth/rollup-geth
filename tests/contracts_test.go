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

package tests

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
)

func TestTXINDEXPrecompile(t *testing.T) {
	testKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	testAddr := crypto.PubkeyToAddress(testKey.PublicKey)

	testFunds := big.NewInt(1e18)

	gspec := &core.Genesis{
		Config: params.TestChainConfig,
		Alloc: core.GenesisAlloc{
			testAddr: {Balance: testFunds},
		},
		BaseFee: big.NewInt(params.InitialBaseFee),
	}

	engine := ethash.NewFaker()

	numTxs := 5
	nonce := uint64(0)
	recipient := common.HexToAddress("0xpleasework")

	txPrecompileAddr := common.HexToAddress("0x0b")

	// Generate the chain with transactions and get the database
	db, blocks, _ := core.GenerateChainWithGenesis(gspec, engine, 1, func(i int, gen *core.BlockGen) {
		for j := 0; j < numTxs; j++ {
			tx := types.NewTransaction(
				nonce,
				recipient,
				big.NewInt(1000),
				params.TxGas,
				big.NewInt(params.InitialBaseFee),
				nil,
			)
			signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)
			if err != nil {
				t.Fatalf("failed to sign tx: %v", err)
			}
			gen.AddTx(signedTx)
			nonce++
		}
	})

	// Create a state database using the database returned from GenerateChainWithGenesis
	trieDB := triedb.NewDatabase(db, triedb.HashDefaults)
	statedb, err := state.New(blocks[0].Root(), state.NewDatabase(trieDB, nil))
	if err != nil {
		t.Fatalf("failed to create state database: %v", err)
	}

	chainRules := params.MergedTestChainConfig.Rules(blocks[0].Number(), false, blocks[0].Time())

	for i, tx := range blocks[0].Transactions() {
		t.Run(fmt.Sprintf("Transaction_%d", i), func(t *testing.T) {
			// Set the transaction context for the specific transaction we're testing
			statedb.SetTxContext(tx.Hash(), i)

			blockCtx := vm.BlockContext{
				CanTransfer: core.CanTransfer,
				Transfer:    core.Transfer,
				BlockNumber: blocks[0].Number(),
				Time:        blocks[0].Time(),
				Difficulty:  blocks[0].Difficulty(),
				GasLimit:    blocks[0].GasLimit(),
				BaseFee:     blocks[0].BaseFee(),
			}

			evm := vm.NewEVM(blockCtx, statedb, params.MergedTestChainConfig, vm.Config{})
			_ = evm

			precompiles := vm.ActivePrecompiledContracts(chainRules)
			txIndexPrecompile, exists := precompiles[txPrecompileAddr]
			if !exists {
				t.Fatalf("txIndex precompile not found at address %s", txPrecompileAddr.Hex())
			}

			input := []byte{}
			gas := txIndexPrecompile.RequiredGas(input)

			result, _, err := vm.RunPrecompiledContract(txIndexPrecompile, input, gas, nil)
			if err != nil {
				t.Fatalf("Failed to run txIndex precompile: %v", err)
			}

			expected := make([]byte, 4)
			binary.BigEndian.PutUint32(expected, uint32(i))

			if !bytes.Equal(result, expected) {
				t.Errorf("Expected output %x, got %x", expected, result)
			}
		})
	}
}
