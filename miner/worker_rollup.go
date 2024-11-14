package miner

import (
	"errors"
	"slices"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

func (miner *Miner) commitVectorFeeTransactions(env *environment, txsByAddress map[common.Address][]*txpool.LazyTransaction, interrupt *atomic.Int32) error {
	txs := sortTxsByNonces(txsByAddress)
	if len(txs) == 0 {
		return nil
	}

	gasLimit := env.header.GasLimit
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(gasLimit)
	}

	for _, tx := range txs {
		// Check interruption signal and abort building if it's fired.
		if interrupt != nil {
			if signal := interrupt.Load(); signal != commitInterruptNone {
				return signalToErr(signal)
			}
		}
		// If we don't have enough gas for any further transactions then we're done.
		if env.gasPool.Gas() < params.TxGas {
			log.Trace("Not enough gas for further transactions", "have", env.gasPool, "want", params.TxGas)
			break
		}
		// Retrieve the next transaction and abort if all done.

		// If we don't have enough space for the next transaction, skip the account.
		if env.gasPool.Gas() < tx.Gas() {
			log.Trace("Not enough gas left for transaction", "hash", tx.Hash, "left", env.gasPool.Gas(), "needed", tx.Gas())
			continue
		}

		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		if tx.Protected() && !miner.chainConfig.IsEIP155(env.header.Number) {
			log.Trace("Ignoring replay protected transaction", "hash", tx.Hash, "eip155", miner.chainConfig.EIP155Block)
			continue
		}

		// Start executing the transaction
		env.state.SetTxContext(tx.Hash(), env.tcount)

		err := miner.commitTransaction(env, tx)
		if err != nil {
			log.Error("[EIP-7706] Transaction failed, account skipped", "hash", tx.Hash, "err", err)
		}
	}

	return nil
}

func sortTxsByNonces(txs map[common.Address][]*txpool.LazyTransaction) types.Transactions {
	txsSorted := make(types.Transactions, 0, len(txs))
	// flatten and load lazy-loaded tx
	for _, txsFromAddress := range txs {
		for _, lazyTx := range txsFromAddress {
			tx := lazyTx.Resolve()
			if tx != nil {
				txsSorted = append(txsSorted, tx)
			}
		}
	}

	slices.SortFunc(txsSorted, func(a, b *types.Transaction) int {
		return int(a.Nonce()) - int(b.Nonce())
	})

	return txsSorted
}
