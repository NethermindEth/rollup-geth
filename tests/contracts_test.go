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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// TestTxIndexer tests the functionalities for managing transaction indexes.
func TestTxIndexer(t *testing.T) {
	// Generate a test key and account.
	testKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	testAddr := crypto.PubkeyToAddress(testKey.PublicKey)
	// Fund the test account.
	testFunds := big.NewInt(1e18) // 1 ETH

	// Build a genesis specification.
	gspec := &core.Genesis{
		Config: params.TestChainConfig,
		Alloc: core.GenesisAlloc{
			testAddr: {Balance: testFunds},
		},
		// BaseFee is used in EIP-1559 blocks; for testing we use an initial value.
		BaseFee: big.NewInt(params.InitialBaseFee),
	}

	// Use the ethash engine in faker mode for deterministic mining.
	engine := ethash.NewFaker()

	// We will create a single block with a 3 transactions.
	numTxs := 3
	nonce := uint64(0)

	txPrecompileAddr := common.HexToAddress("0x0b")

	_, _, receipts := core.GenerateChainWithGenesis(gspec, engine, numTxs, func(i int, gen *core.BlockGen) {
		// Create a transaction that calls the txIndex precompile.
		tx := types.NewTransaction(nonce, txPrecompileAddr, big.NewInt(0), params.TxGas, big.NewInt(10*params.InitialBaseFee), nil)
		signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)
		if err != nil {
			t.Fatalf("failed to sign tx: %v", err)
		}
		gen.AddTx(signedTx)
		nonce++
	})

	if len(receipts) != numTxs {
		t.Fatalf("expected %d receipts, got %d", numTxs, len(receipts[0]))
	}

	for i, receipt := range receipts[0] {
		expected := make([]byte, 4)
		binary.BigEndian.PutUint32(expected, uint32(i))
		if !bytes.Equal(receipt.output, expected) {
			t.Errorf("transaction %d: expected output %x, got %x", i, expected, receipt)
		}
	}
	// for i, receipt := range receipts[0] {
	// 	if receipt.Status != types.ReceiptStatusSuccessful {
	// 		t.Errorf("transaction %d: expected successful status", i)
	// 	}
	// }
}
