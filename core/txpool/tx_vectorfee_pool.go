//NOTE: THIS IS DUMMY POOL; It is far from prod ready!!!
//it is intended as helper for local testing

package txpool

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

type VectorFeePoolDummy struct {
	lock sync.RWMutex

	reserve AddressReserver // Address reserver to ensure exclusivity across subpools

	txs          map[common.Hash]*types.Transaction
	txsByAddress map[common.Address]types.Transactions // Transactions grouped by account

	chain  BlockChain   // Chain object to access the state through
	signer types.Signer // Transaction signer to use for sender recovery

	head  *types.Header  // Current head of the chain
	state *state.StateDB // Current state at the head of the chain

	discoverFeed event.Feed // Event feed to send out new tx events on pool discovery (reorg excluded)
	insertFeed   event.Feed // Event feed to send out new tx events on pool inclusion (reorg included)

}

func NewVectorFeePoolDummy(chain BlockChain) *VectorFeePoolDummy {
	return &VectorFeePoolDummy{
		chain:        chain,
		signer:       types.LatestSigner(chain.Config()),
		txs:          make(map[common.Hash]*types.Transaction),
		txsByAddress: make(map[common.Address]types.Transactions),
	}
}

// Filter is a selector used to decide whether a transaction would be added
// to VectorFeePool.
func (pool *VectorFeePoolDummy) Filter(tx *types.Transaction) bool {
	return tx.Type() == types.VectorFeeTxType
}

// Init sets the base parameters of the subpool, allowing it to load any saved
// transactions from disk and also permitting internal maintenance routines to
// start up.
//
// These should not be passed as a constructor argument - nor should the pools
// start by themselves - in order to keep multiple subpools in lockstep with
// one another.
func (pool *VectorFeePoolDummy) Init(gasTip uint64, head *types.Header, reserve AddressReserver) error {
	// Initialize the state with head block, or fallback to empty one in
	// case the head state is not available (might occur when node is not
	// fully synced).
	state, err := pool.chain.StateAt(head.Root)
	if err != nil {
		state, err = pool.chain.StateAt(types.EmptyRootHash)
	}
	if err != nil {
		return err
	}
	pool.head, pool.state = head, state

	return nil
}

// Close terminates any background processing threads and releases any held
// resources.
func (pool *VectorFeePoolDummy) Close() error {
	return nil
}

// Reset retrieves the current state of the blockchain and ensures the content
// of the transaction pool is valid with regard to the chain state.
// For VectorFeePool, since it is dummy pool this not even naive implementation, but stupid
// We just go through all transactions received by the block and remove them from the pool
func (pool *VectorFeePoolDummy) Reset(oldHead, newHead *types.Header) {
	statedb, err := pool.chain.StateAt(newHead.Root)
	if err != nil {
		log.Error("Failed to reset vector fee tx pool state", "err", err)
		return
	}

	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.head, pool.state = newHead, statedb

	latestBlock := pool.chain.GetBlock(newHead.Hash(), newHead.Number.Uint64())
	if latestBlock == nil {
		return
	}

	for _, tx := range latestBlock.Transactions() {
		from, err := types.Sender(pool.signer, tx)
		if err != nil {
			continue
		}

		delete(pool.txs, tx.Hash())

		//TODO: remove all TXs which nonce is less than the latest nonce for the given address
		// Handle deleting transactions grouped by sender
		if txsByAddress, ok := pool.txsByAddress[from]; ok {
			// We don't order txs by nonce so we don't care about persevering he ordering
			for i, txFromAddress := range txsByAddress {
				if txToDelFound := txFromAddress.Hash() == tx.Hash(); txToDelFound {
					txsByAddress[i] = txsByAddress[len(txsByAddress)-1]
					pool.txsByAddress[from] = txsByAddress[:len(txsByAddress)-1]
					break
				}
			}

			if addressHasNoTxsLeftInPool := len(pool.txsByAddress[from]) == 0; addressHasNoTxsLeftInPool {
				delete(pool.txsByAddress, from)
			}
		}
	}
}

// SetGasTip updates the minimum price required by the subpool for a new
// transaction, and drops all transactions below this threshold.
func (pool *VectorFeePoolDummy) SetGasTip(tip *big.Int) {}

// Has returns an indicator whether subpool has a transaction cached with the
// given hash.
func (pool *VectorFeePoolDummy) Has(hash common.Hash) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	tx, has := pool.txs[hash]
	return has && tx != nil
}

// Get returns a transaction if it is contained in the pool, or nil otherwise.
func (pool *VectorFeePoolDummy) Get(hash common.Hash) *types.Transaction {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	tx := pool.txs[hash]
	return tx
}

// Add enqueues a batch of transactions into the pool if they are valid. Due
// to the large transaction churn, add may postpone fully integrating the tx
// to a later point to batch multiple ones together.
func (pool *VectorFeePoolDummy) Add(txs []*types.Transaction, local bool, sync bool) []error {
	if len(txs) == 0 {
		return nil
	}

	//TODO: validate tx
	pool.lock.Lock()
	defer pool.lock.Unlock()

	errors := make([]error, len(txs))
	adds := make(types.Transactions, 0, len(txs))

	for i, tx := range txs {
		h := tx.Hash()
		if _, alreadyInThePool := pool.txs[h]; alreadyInThePool {
			continue
		}

		from, err := types.Sender(pool.signer, tx)
		if err != nil {
			errors[i] = err
			continue
		}

		pool.txs[tx.Hash()] = tx
		pool.txsByAddress[from] = append(pool.txsByAddress[from], tx)

		adds = append(adds, tx)
		log.Trace("Pooled new future transaction", "hash", tx.Hash(), "from", from, "to", tx.To())
	}

	if len(adds) > 0 {
		pool.insertFeed.Send(core.NewTxsEvent{Txs: adds})
		pool.discoverFeed.Send(core.NewTxsEvent{Txs: adds})
	}

	return errors
}

// Pending retrieves all currently processable transactions, grouped by origin
// account and sorted by nonce.
//
// The transactions can also be pre-filtered by the dynamic fee components to
// reduce allocations and load on downstream subsystems.
func (pool *VectorFeePoolDummy) Pending(filter PendingFilter) map[common.Address][]*LazyTransaction {
	if filter.OnlyBlobTxs || filter.OnlyPlainTxs {
		return nil
	}

	pool.lock.RLock()
	defer pool.lock.RUnlock()

	execStart := time.Now()
	result := make(map[common.Address][]*LazyTransaction, len(pool.txsByAddress))

	for address, txs := range pool.txsByAddress {
		lazyTxs := make([]*LazyTransaction, len(txs))

		for i, tx := range txs {
			lazyTx := &LazyTransaction{
				Pool:      pool,
				Hash:      tx.Hash(),
				Time:      execStart,
				GasFeeCap: uint256.MustFromBig(tx.GasFeeCap()),
				GasTipCap: uint256.MustFromBig(tx.GasTipCap()),
				Gas:       tx.Gas(),
				BlobGas:   tx.BlobGas(),
			}

			lazyTxs[i] = lazyTx
		}

		result[address] = lazyTxs
	}

	return result
}

// SubscribeTransactions subscribes to new transaction events. The subscriber
// can decide whether to receive notifications only for newly seen transactions
// or also for reorged out ones.
func (pool *VectorFeePoolDummy) SubscribeTransactions(ch chan<- core.NewTxsEvent, reorgs bool) event.Subscription {
	if reorgs {
		return pool.insertFeed.Subscribe(ch)
	} else {
		return pool.discoverFeed.Subscribe(ch)
	}
}

// Nonce returns the next nonce of an account, with all transactions executable
// by the pool already applied on top.
func (pool *VectorFeePoolDummy) Nonce(addr common.Address) uint64 {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if txs, ok := pool.txsByAddress[addr]; ok && len(txs) > 0 {
		// this is needed because txs are not ordered by nonce
		maxNonce := txs[0].Nonce()
		for _, tx := range txs {
			if tx.Nonce() > maxNonce {
				maxNonce = tx.Nonce()
			}
		}

		return maxNonce + 1
	}

	return pool.state.GetNonce(addr)
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *VectorFeePoolDummy) Stats() (int, int) {
	return 0, 0
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *VectorFeePoolDummy) Content() (map[common.Address][]*types.Transaction, map[common.Address][]*types.Transaction) {
	return make(map[common.Address][]*types.Transaction), make(map[common.Address][]*types.Transaction)
}

// ContentFrom retrieves the data content of the transaction pool, returning the
// pending as well as queued transactions of this address, grouped by nonce.
func (pool *VectorFeePoolDummy) ContentFrom(addr common.Address) ([]*types.Transaction, []*types.Transaction) {
	return []*types.Transaction{}, []*types.Transaction{}
}

// Locals retrieves the accounts currently considered local by the pool.
func (pool *VectorFeePoolDummy) Locals() []common.Address {
	return []common.Address{}
}

// Status returns the known status (unknown/pending/queued) of a transaction
// identified by their hashes.
func (pool *VectorFeePoolDummy) Status(hash common.Hash) TxStatus {
	if pool.Has(hash) {
		return TxStatusPending
	}
	return TxStatusUnknown
}
