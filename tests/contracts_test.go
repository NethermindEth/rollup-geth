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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func TestSlotPrecompile(t *testing.T) {
	t.Run("Test Manual Slot Precompile", func(t *testing.T) {
		// Test multiple slot values
		slotValues := []uint64{123, 456, 789}

		for _, slotValue := range slotValues {
			mockPrecompile := &vm.SlotPrecompile{SlotNumber: slotValue}

			// Run the precompile
			input := []byte{}
			gas := mockPrecompile.RequiredGas(input)

			result, remainingGas, err := vm.RunPrecompiledContract(mockPrecompile, input, gas, nil)
			if err != nil {
				t.Fatalf("Failed to run slot precompile: %v", err)
			}

			if remainingGas != 0 {
				t.Errorf("Expected all gas to be used, but %d remained", remainingGas)
			}

			// Verify the result
			expected := []byte{0, 0, 0, 0, 0, 0, 0, 0}
			binary.BigEndian.PutUint64(expected, slotValue)

			if !bytes.Equal(result, expected) {
				t.Errorf("Expected slot value %d (0x%x), got %d (0x%x)",
					slotValue, expected, binary.BigEndian.Uint64(result), result)
			}
		}
	})
	t.Run("Slot Precompile with multiple blocks", func(t *testing.T) {
		testKey, err := crypto.GenerateKey()
		if err != nil {
			t.Fatal(err)
		}
		testAddr := crypto.PubkeyToAddress(testKey.PublicKey)
		testFunds := big.NewInt(1e18)
		SlotPrecompileAddr := common.HexToAddress("0x12")

		gspec := &core.Genesis{
			Config: params.MergedTestChainConfig,
			Alloc: types.GenesisAlloc{
				testAddr: {Balance: testFunds},
			},
			BaseFee: big.NewInt(params.InitialBaseFee),
		}

		engine := ethash.NewFaker()
		//engine := beacon.New(ethash.NewFaker())
		numBlocks := 5
		nonce := uint64(0)
		recipient := common.HexToAddress("plzwork")
		_, blocks, _ := core.GenerateChainWithGenesis(gspec, engine, numBlocks, func(i int, gen *core.BlockGen) {
			// Each block simulates a SlotPrecompile with the current slot number as the block number
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
			slotNumber := gen.Number().Uint64() // this returns the block number
			gen.SetSlotNumber(slotNumber)
			chainRules := params.MergedTestChainConfig.Rules(gen.Number(), true, gen.Timestamp())
			precompiles := vm.ActivePrecompiledContracts(chainRules)
			SlotPrecompile, exists := precompiles[SlotPrecompileAddr]
			if !exists {
				t.Fatalf("SlotPrecompile not found at address %s", SlotPrecompileAddr.Hex())
			}

			input := []byte{}
			gas := SlotPrecompile.RequiredGas(input)
			result, remainingGas, err := vm.RunPrecompiledContract(SlotPrecompile, input, gas, nil)
			if err != nil {
				t.Fatalf("Failed to run Slot precompile after block %d: %v", slotNumber, err)
			}

			if remainingGas != 0 {
				t.Errorf("Expected all gas to be used, but %d remained", remainingGas)
			}
			expected := make([]byte, 8)
			binary.BigEndian.PutUint64(expected, slotNumber)

			if !bytes.Equal(result, expected) {
				t.Errorf("Block %d: expected slot output %x, got %x", slotNumber, expected, result)
			}
		})

		if len(blocks) != numBlocks {
			t.Fatalf("Expected %d blocks, got %d", numBlocks, len(blocks))
		}
	})
}
