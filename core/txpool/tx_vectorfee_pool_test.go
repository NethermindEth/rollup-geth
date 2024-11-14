package txpool

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestVectorFeePool_Add(t *testing.T) {
	_, pool, key, addr := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	tests := []struct {
		name string
		tx   *types.Transaction
	}{
		{
			name: "valid transaction",
			tx: createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1),
				21000, key),
		},
		{
			name: "duplicate transaction",
			tx: createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1),
				21000, key),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := pool.Add([]*types.Transaction{tt.tx}, false, false)
			assert.Len(t, errs, 0)
			assert.Len(t, pool.txs, 1)
			assert.Len(t, pool.txsByAddress, 1)
			assert.Len(t, pool.txsByAddress[addr], 1)

			assert.True(t, pool.Has(tt.tx.Hash()))
		})
	}
}

func TestVectorFeePool_Pending(t *testing.T) {
	_, pool, key, addr := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Add multiple transactions
	txs := []*types.Transaction{
		createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1), 21000, key),
		createSignedVectorFeeTx(t, 1, recipient, big.NewInt(1), 21000, key),
		createSignedVectorFeeTx(t, 2, recipient, big.NewInt(1), 21000, key),

		// duplicate should not be added
		createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1), 21000, key),
	}

	errs := pool.Add(txs, false, false)
	assert.Empty(t, errs)

	pending := pool.Pending(PendingFilter{})
	assert.Len(t, pending[addr], 3)
}

func TestVectorFeePool_Reset(t *testing.T) {
	blockchain, pool, key, _ := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Add a transaction
	tx := createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1), 21000, key)
	errs := pool.Add([]*types.Transaction{tx}, false, false)
	assert.Empty(t, errs)

	// Verify it's in the pool
	assert.True(t, pool.Has(tx.Hash()))

	// Create a new block that doesn't include our TX
	newHeader := &types.Header{
		Number:     big.NewInt(1),
		GasLimit:   8000000,
		ParentHash: pool.head.Hash(),
	}

	pool.Reset(pool.head, newHeader)
	// We didn't receive block with this transaction yet, so it should still be in the pool
	assert.True(t, pool.Has(tx.Hash()))

	// Create a new block that includes our TX
	newHeader = &types.Header{
		Number:     big.NewInt(2),
		GasLimit:   8000000,
		ParentHash: pool.head.Hash(),
	}

	newBlock := types.NewBlockWithHeader(newHeader).WithBody(types.Body{
		Transactions: types.Transactions{tx},
	})
	blockchain.blocks[newHeader.Number.Uint64()] = newBlock

	pool.Reset(pool.head, newHeader)

	// Pool should be empty
	assert.False(t, pool.Has(tx.Hash()))
	assert.Len(t, pool.txs, 0)
	assert.Len(t, pool.txsByAddress, 0)
}

func TestVectorFeePool_Get(t *testing.T) {
	_, pool, key, _ := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Add a transaction
	tx := createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1000), 21000, key)
	errs := pool.Add([]*types.Transaction{tx}, false, false)
	assert.Empty(t, errs)

	// Test Get
	retrieved := pool.Get(tx.Hash())
	assert.NotNil(t, retrieved)
	assert.Equal(t, tx.Hash(), retrieved.Hash())

	// Test Get with non-existent transaction
	retrieved = pool.Get(common.Hash{})
	assert.Nil(t, retrieved)
}

func TestVectorFeePool_Nonce(t *testing.T) {
	_, pool, key, addr := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Check initial nonce
	assert.Equal(t, uint64(0), pool.Nonce(addr))

	// Add transactions with different nonces
	txs := []*types.Transaction{
		createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1000), 21000, key),
		createSignedVectorFeeTx(t, 1, recipient, big.NewInt(1000), 21000, key),
		createSignedVectorFeeTx(t, 2, recipient, big.NewInt(1000), 21000, key),
	}

	errs := pool.Add(txs, false, false)
	assert.Empty(t, errs)

	// Check nonce after adding transactions
	assert.Equal(t, uint64(3), pool.Nonce(addr))
}

func TestVectorFeePool_Filter(t *testing.T) {
	_, pool, key, _ := setupTestPool(t)
	defer pool.Close()

	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create different types of transactions
	vectorFeeTx := createSignedVectorFeeTx(t, 0, recipient, big.NewInt(1000), 21000, key)
	legacyTx := types.NewTransaction(0, recipient, big.NewInt(1000), 21000, big.NewInt(1), nil)

	// Test filter
	assert.True(t, pool.Filter(vectorFeeTx))
	assert.False(t, pool.Filter(legacyTx))
}

type testBlockChain struct {
	statedb       *state.StateDB
	config        *params.ChainConfig
	gasLimit      uint64
	chainHeadFeed *event.Feed

	blocks map[uint64]*types.Block
}

func (bc *testBlockChain) CurrentBlock() *types.Header {
	return &types.Header{
		Number:   new(big.Int),
		GasLimit: bc.gasLimit,
	}
}

func (bc *testBlockChain) StateAt(root common.Hash) (*state.StateDB, error) {
	return bc.statedb, nil
}

func (bc *testBlockChain) Config() *params.ChainConfig {
	return bc.config
}

func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return bc.chainHeadFeed.Subscribe(ch)
}

func (bc *testBlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	block := bc.blocks[number]
	return block
}

func setupTestPool(t *testing.T) (*testBlockChain, *VectorFeePoolDummy, *ecdsa.PrivateKey, common.Address) {
	var (
		db  = rawdb.NewMemoryDatabase()
		tdb = triedb.NewDatabase(db, nil)
		sdb = state.NewDatabase(tdb, nil)
	)
	statedb, _ := state.New(types.EmptyRootHash, sdb)

	blockchain := &testBlockChain{
		statedb:  statedb,
		config:   getBlockChianConfig(),
		gasLimit: 8000000,
		blocks:   make(map[uint64]*types.Block),
	}

	pool := NewVectorFeePoolDummy(blockchain)

	// Create a funded account
	key, addr := generateAccount()
	statedb.AddBalance(addr, uint256.NewInt(1000000000000000000), tracing.BalanceChangeUnspecified) // 1 ETH

	err := pool.Init(1, blockchain.CurrentBlock(), func(addr common.Address, reserve bool) error {
		return nil
	})
	assert.NoError(t, err)

	return blockchain, pool, key, addr
}

func generateAccount() (*ecdsa.PrivateKey, common.Address) {
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

func createSignedVectorFeeTx(t *testing.T, nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, key *ecdsa.PrivateKey) *types.Transaction {
	tx := types.NewTx(&types.VectorFeeTx{
		ChainID:    uint256.NewInt(1),
		Nonce:      nonce,
		Gas:        gasLimit,
		GasTipCaps: types.VectorFeeUint{uint256.NewInt(1), uint256.NewInt(3), uint256.NewInt(3)},
		GasFeeCaps: types.VectorFeeUint{uint256.NewInt(4), uint256.NewInt(5), uint256.NewInt(6)},
		To:         to,
		Value:      uint256.MustFromBig(amount),
		Data:       nil,
	})

	signedTx, err := types.SignTx(tx, types.LatestSigner(getBlockChianConfig()), key)
	if err != nil {
		t.Errorf("Could not sign tx: %v", err)
		t.FailNow()
	}

	return signedTx
}

func getBlockChianConfig() *params.ChainConfig {
	EIP7706Time := uint64(0)
	var blockChainConfig params.ChainConfig = params.ChainConfig{
		ChainID:     big.NewInt(1),
		Ethash:      new(params.EthashConfig),
		EIP7706Time: &EIP7706Time,
	}

	return &blockChainConfig
}
