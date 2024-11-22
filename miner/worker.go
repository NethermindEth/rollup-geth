// Copyright 2015 The go-ethereum Authors
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

package miner

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/consensus/misc/eip4844"
	"github.com/ethereum/go-ethereum/consensus/misc/eip7706"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/types/interoptypes"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

const (
	// minRecommitInterruptInterval is the minimum time interval used to interrupt filling a
	// sealing block with pending transactions from the mempool
	minRecommitInterruptInterval = 2 * time.Second
)

var (
	errTxConditionalInvalid = errors.New("transaction conditional failed")

	errBlockInterruptedByNewHead  = errors.New("new head arrived while building block")
	errBlockInterruptedByRecommit = errors.New("recommit interrupt while building block")
	errBlockInterruptedByTimeout  = errors.New("timeout while building block")
	errBlockInterruptedByResolve  = errors.New("payload resolution while building block")

	txConditionalRejectedCounter = metrics.NewRegisteredCounter("miner/transactionConditional/rejected", nil)
	txConditionalMinedTimer      = metrics.NewRegisteredTimer("miner/transactionConditional/elapsedtime", nil)

	txInteropRejectedCounter = metrics.NewRegisteredCounter("miner/transactionInterop/rejected", nil)
)

// environment is the worker's current environment and holds all
// information of the sealing block generation.
type environment struct {
	signer   types.Signer
	state    *state.StateDB // apply state changes here
	tcount   int            // tx count in cycle
	gasPool  *core.GasPool  // available gas used to pack transactions
	coinbase common.Address

	header   *types.Header
	txs      []*types.Transaction
	receipts []*types.Receipt
	sidecars []*types.BlobTxSidecar
	blobs    int

	witness *stateless.Witness

	noTxs  bool            // true if we are reproducing a block, and do not have to check interop txs
	rpcCtx context.Context // context to control block-building RPC work. No RPC allowed if nil.
}

const (
	commitInterruptNone int32 = iota
	commitInterruptNewHead
	commitInterruptResubmit
	commitInterruptTimeout
	commitInterruptResolve
)

// newPayloadResult is the result of payload generation.
type newPayloadResult struct {
	err      error
	block    *types.Block
	fees     *big.Int               // total block fees
	sidecars []*types.BlobTxSidecar // collected blobs of blob transactions
	stateDB  *state.StateDB         // StateDB after executing the transactions
	receipts []*types.Receipt       // Receipts collected during construction
	witness  *stateless.Witness     // Witness is an optional stateless proof
}

// generateParams wraps various settings for generating sealing task.
type generateParams struct {
	timestamp   uint64            // The timestamp for sealing task
	forceTime   bool              // Flag whether the given timestamp is immutable or not
	parentHash  common.Hash       // Parent block hash, empty means the latest chain head
	coinbase    common.Address    // The fee recipient address for including transaction
	random      common.Hash       // The randomness generated by beacon chain, empty before the merge
	withdrawals types.Withdrawals // List of withdrawals to include in block (shanghai field)
	beaconRoot  *common.Hash      // The beacon root (cancun field).
	noTxs       bool              // Flag whether an empty block without any transaction is expected

	txs           types.Transactions // Deposit transactions to include at the start of the block
	gasLimit      *uint64            // Optional gas limit override
	eip1559Params []byte             // Optional EIP-1559 parameters
	interrupt     *atomic.Int32      // Optional interruption signal to pass down to worker.generateWork
	isUpdate      bool               // Optional flag indicating that this is building a discardable update

	rpcCtx context.Context // context to control block-building RPC work. No RPC allowed if nil.
}

// generateWork generates a sealing block based on the given parameters.
func (miner *Miner) generateWork(params *generateParams, witness bool) *newPayloadResult {
	work, err := miner.prepareWork(params, witness)
	if err != nil {
		return &newPayloadResult{err: err}
	}
	if work.gasPool == nil {
		gasLimit := work.header.GasLimit

		// If we're building blocks with mempool transactions, we need to ensure that the
		// gas limit is not higher than the effective gas limit. We must still accept any
		// explicitly selected transactions with gas usage up to the block header's limit.
		if !params.noTxs {
			effectiveGasLimit := miner.config.EffectiveGasCeil
			if effectiveGasLimit != 0 && effectiveGasLimit < gasLimit {
				gasLimit = effectiveGasLimit
			}
		}
		work.gasPool = new(core.GasPool).AddGas(gasLimit)
	}

	misc.EnsureCreate2Deployer(miner.chainConfig, work.header.Time, work.state)

	for _, tx := range params.txs {
		from, _ := types.Sender(work.signer, tx)
		work.state.SetTxContext(tx.Hash(), work.tcount)
		err = miner.commitTransaction(work, tx)
		if err != nil {
			return &newPayloadResult{err: fmt.Errorf("failed to force-include tx: %s type: %d sender: %s nonce: %d, err: %w", tx.Hash(), tx.Type(), from, tx.Nonce(), err)}
		}
	}
	if !params.noTxs {
		// use shared interrupt if present
		interrupt := params.interrupt
		if interrupt == nil {
			interrupt = new(atomic.Int32)
		}
		timer := time.AfterFunc(max(minRecommitInterruptInterval, miner.config.Recommit), func() {
			interrupt.Store(commitInterruptTimeout)
		})

		err := miner.fillTransactions(interrupt, work)
		timer.Stop() // don't need timeout interruption any more
		if errors.Is(err, errBlockInterruptedByTimeout) {
			log.Warn("Block building is interrupted", "allowance", common.PrettyDuration(miner.config.Recommit))
		} else if errors.Is(err, errBlockInterruptedByResolve) {
			log.Info("Block building got interrupted by payload resolution")
		}
	}
	if intr := params.interrupt; intr != nil && params.isUpdate && intr.Load() != commitInterruptNone {
		return &newPayloadResult{err: errInterruptedUpdate}
	}

	body := types.Body{Transactions: work.txs, Withdrawals: params.withdrawals}
	allLogs := make([]*types.Log, 0)
	for _, r := range work.receipts {
		allLogs = append(allLogs, r.Logs...)
	}
	// Read requests if Prague is enabled.
	if miner.chainConfig.IsPrague(work.header.Number, work.header.Time) {
		requests, err := core.ParseDepositLogs(allLogs, miner.chainConfig)
		if err != nil {
			return &newPayloadResult{err: err}
		}
		body.Requests = requests
	}
	block, err := miner.engine.FinalizeAndAssemble(miner.chain, work.header, work.state, &body, work.receipts)
	if err != nil {
		return &newPayloadResult{err: err}
	}
	return &newPayloadResult{
		block:    block,
		fees:     totalFees(block, work.receipts),
		sidecars: work.sidecars,
		stateDB:  work.state,
		receipts: work.receipts,
		witness:  work.witness,
	}
}

// prepareWork constructs the sealing task according to the given parameters,
// either based on the last chain head or specified parent. In this function
// the pending transactions are not filled yet, only the empty task returned.
func (miner *Miner) prepareWork(genParams *generateParams, witness bool) (*environment, error) {
	miner.confMu.RLock()
	defer miner.confMu.RUnlock()

	// Find the parent block for sealing task
	parent := miner.chain.CurrentBlock()
	if genParams.parentHash != (common.Hash{}) {
		block := miner.chain.GetBlockByHash(genParams.parentHash)
		if block == nil {
			return nil, errors.New("missing parent")
		}
		parent = block.Header()
	}
	// Sanity check the timestamp correctness, recap the timestamp
	// to parent+1 if the mutation is allowed.
	timestamp := genParams.timestamp
	if parent.Time >= timestamp {
		if genParams.forceTime {
			return nil, fmt.Errorf("invalid timestamp, parent %d given %d", parent.Time, timestamp)
		}
		timestamp = parent.Time + 1
	}
	// Construct the sealing block header.
	header := &types.Header{
		ParentHash: parent.Hash(),
		Number:     new(big.Int).Add(parent.Number, common.Big1),
		GasLimit:   core.CalcGasLimit(parent.GasLimit, miner.config.GasCeil),
		Time:       timestamp,
		Coinbase:   genParams.coinbase,
	}
	// Set the extra field.
	if len(miner.config.ExtraData) != 0 && miner.chainConfig.Optimism == nil {
		// Optimism chains have their own ExtraData handling rules
		header.Extra = miner.config.ExtraData
	}
	// Set the randomness field from the beacon chain if it's available.
	if genParams.random != (common.Hash{}) {
		header.MixDigest = genParams.random
	}
	// Set baseFee and GasLimit if we are on an EIP-1559 chain
	if miner.chainConfig.IsLondon(header.Number) {
		header.BaseFee = eip1559.CalcBaseFee(miner.chainConfig, parent, header.Time)
		if !miner.chainConfig.IsLondon(parent.Number) {
			parentGasLimit := parent.GasLimit * miner.chainConfig.ElasticityMultiplier()
			header.GasLimit = core.CalcGasLimit(parentGasLimit, miner.config.GasCeil)
		}
	}
	if genParams.gasLimit != nil { // override gas limit if specified
		header.GasLimit = *genParams.gasLimit
	} else if miner.chain.Config().Optimism != nil && miner.config.GasCeil != 0 {
		// configure the gas limit of pending blocks with the miner gas limit config when using optimism
		header.GasLimit = miner.config.GasCeil
	}
	if miner.chainConfig.IsHolocene(header.Time) {
		if err := eip1559.ValidateHolocene1559Params(genParams.eip1559Params); err != nil {
			return nil, err
		}
		// If this is a holocene block and the params are 0, we must convert them to their previous
		// constants in the header.
		d, e := eip1559.DecodeHolocene1559Params(genParams.eip1559Params)
		if d == 0 {
			d = miner.chainConfig.BaseFeeChangeDenominator(header.Time)
			e = miner.chainConfig.ElasticityMultiplier()
		}
		header.Extra = eip1559.EncodeHoloceneExtraData(d, e)
	} else if genParams.eip1559Params != nil {
		return nil, errors.New("got eip1559 params, expected none")
	}
	// Run the consensus preparation with the default or customized consensus engine.
	// Note that the `header.Time` may be changed.
	if err := miner.engine.Prepare(miner.chain, header); err != nil {
		log.Error("Failed to prepare header for sealing", "err", err)
		return nil, err
	}
	// Apply EIP-4844, EIP-4788.
	if miner.chainConfig.IsCancun(header.Number, header.Time) {
		var excessBlobGas uint64
		if miner.chainConfig.IsCancun(parent.Number, parent.Time) {
			excessBlobGas = eip4844.CalcExcessBlobGas(*parent.ExcessBlobGas, *parent.BlobGasUsed)
		} else {
			// For the first post-fork block, both parent.data_gas_used and parent.excess_data_gas are evaluated as 0
			excessBlobGas = eip4844.CalcExcessBlobGas(0, 0)
		}
		header.BlobGasUsed = new(uint64)
		header.ExcessBlobGas = &excessBlobGas
		header.ParentBeaconRoot = genParams.beaconRoot
	}

	if miner.chainConfig.IsEIP7706(header.Number, header.Time) {
		parentGasUsed, parentExcessGas, parentGasLimits := eip7706.SanitizeEIP7706Fields(parent)

		header.ExcessGas = eip7706.CalcExecGas(parentGasUsed, parentExcessGas, parentGasLimits)
		header.GasLimits = core.CalcGasLimits(parent.GasLimit, miner.config.GasCeil)
		header.BaseFees = eip7706.CalcBaseFees(parentExcessGas, parentGasLimits)
	}

	// Could potentially happen if starting to mine in an odd state.
	// Note genParams.coinbase can be different with header.Coinbase
	// since clique algorithm can modify the coinbase field in header.
	env, err := miner.makeEnv(parent, header, genParams.coinbase, witness, genParams.rpcCtx)
	if err != nil {
		log.Error("Failed to create sealing context", "err", err)
		return nil, err
	}
	env.noTxs = genParams.noTxs
	if header.ParentBeaconRoot != nil {
		context := core.NewEVMBlockContext(header, miner.chain, nil, miner.chainConfig, env.state)
		vmenv := vm.NewEVM(context, vm.TxContext{}, env.state, miner.chainConfig, vm.Config{})
		core.ProcessBeaconBlockRoot(*header.ParentBeaconRoot, vmenv, env.state)
	}
	if miner.chainConfig.IsPrague(header.Number, header.Time) {
		context := core.NewEVMBlockContext(header, miner.chain, nil, miner.chainConfig, env.state)
		vmenv := vm.NewEVM(context, vm.TxContext{}, env.state, miner.chainConfig, vm.Config{})
		core.ProcessParentBlockHash(header.ParentHash, vmenv, env.state)
	}
	return env, nil
}

// makeEnv creates a new environment for the sealing block.
func (miner *Miner) makeEnv(parent *types.Header, header *types.Header, coinbase common.Address, witness bool, rpcCtx context.Context) (*environment, error) {
	// Retrieve the parent state to execute on top.
	state, err := miner.chain.StateAt(parent.Root)
	if err != nil {
		return nil, err
	}
	if miner.chainConfig.Optimism != nil { // Allow the miner to reorg its own chain arbitrarily deep
		if historicalBackend, ok := miner.backend.(BackendWithHistoricalState); ok {
			var release tracers.StateReleaseFunc
			parentBlock := miner.backend.BlockChain().GetBlockByHash(parent.Hash())
			state, release, err = historicalBackend.StateAtBlock(context.Background(), parentBlock, ^uint64(0), nil, false, false)
			if err != nil {
				return nil, err
			}
			state = state.Copy()
			release()
		}
	}

	if witness {
		bundle, err := stateless.NewWitness(header, miner.chain)
		if err != nil {
			return nil, err
		}
		state.StartPrefetcher("miner", bundle)
	}
	// Note the passed coinbase may be different with header.Coinbase.
	return &environment{
		signer:   types.MakeSigner(miner.chainConfig, header.Number, header.Time),
		state:    state,
		coinbase: coinbase,
		header:   header,
		witness:  state.Witness(),
		rpcCtx:   rpcCtx,
	}, nil
}

func (miner *Miner) commitTransaction(env *environment, tx *types.Transaction) error {
	if tx.Type() == types.BlobTxType {
		return miner.commitBlobTransaction(env, tx)
	}

	// If a conditional is set, check prior to applying
	if conditional := tx.Conditional(); conditional != nil {
		txConditionalMinedTimer.UpdateSince(tx.Time())

		// check the conditional
		if err := env.header.CheckTransactionConditional(conditional); err != nil {
			return fmt.Errorf("failed header check: %s: %w", err, errTxConditionalInvalid)
		}
		if err := env.state.CheckTransactionConditional(conditional); err != nil {
			return fmt.Errorf("failed state check: %s: %w", err, errTxConditionalInvalid)
		}
	}

	receipt, err := miner.applyTransaction(env, tx)
	if err != nil {
		return err
	}
	env.txs = append(env.txs, tx)
	env.receipts = append(env.receipts, receipt)
	env.tcount++
	return nil
}

func (miner *Miner) commitBlobTransaction(env *environment, tx *types.Transaction) error {
	sc := tx.BlobTxSidecar()
	if sc == nil {
		panic("blob transaction without blobs in miner")
	}
	// Checking against blob gas limit: It's kind of ugly to perform this check here, but there
	// isn't really a better place right now. The blob gas limit is checked at block validation time
	// and not during execution. This means core.ApplyTransaction will not return an error if the
	// tx has too many blobs. So we have to explicitly check it here.
	if (env.blobs+len(sc.Blobs))*params.BlobTxBlobGasPerBlob > params.MaxBlobGasPerBlock {
		return errors.New("max data blobs reached")
	}
	receipt, err := miner.applyTransaction(env, tx)
	if err != nil {
		return err
	}
	env.txs = append(env.txs, tx.WithoutBlobTxSidecar())
	env.receipts = append(env.receipts, receipt)
	env.sidecars = append(env.sidecars, sc)
	env.blobs += len(sc.Blobs)
	*env.header.BlobGasUsed += receipt.BlobGasUsed
	env.tcount++
	return nil
}

type LogInspector interface {
	GetLogs(hash common.Hash, blockNumber uint64, blockHash common.Hash) []*types.Log
}

// applyTransaction runs the transaction. If execution fails, state and gas pool are reverted.
func (miner *Miner) applyTransaction(env *environment, tx *types.Transaction) (*types.Receipt, error) {
	var (
		snap = env.state.Snapshot()
		gp   = env.gasPool.Gas()
	)
	var extraOpts *core.ApplyTransactionOpts
	// If not just reproducing the block, check the interop executing messages.
	if !env.noTxs && miner.chain.Config().IsInterop(env.header.Time) {
		// Whenever there are `noTxs` it means we are building a block from pre-determined txs. There are two cases:
		//	(1) it's derived from L1, and will be verified asynchronously by the op-node.
		//	(2) it is a deposits-only empty-block by the sequencer, in which case there are no interop-txs to verify (as deposits do not emit any).

		// We have to insert as call-back, since we cannot revert the snapshot
		// after the tx is deemed successful and the journal has been cleared already.
		extraOpts = &core.ApplyTransactionOpts{
			PostValidation: func(evm *vm.EVM, result *core.ExecutionResult) error {
				logInspector, ok := evm.StateDB.(LogInspector)
				if !ok {
					return fmt.Errorf("cannot get logs from StateDB type %T", evm.StateDB)
				}
				logs := logInspector.GetLogs(tx.Hash(), env.header.Number.Uint64(), common.Hash{})
				return miner.checkInterop(env.rpcCtx, tx, result.Failed(), logs)
			},
		}
	}
	receipt, err := core.ApplyTransactionExtended(miner.chainConfig, miner.chain, &env.coinbase, env.gasPool, env.state, env.header, tx, &env.header.GasUsed, vm.Config{}, extraOpts)
	if err != nil {
		env.state.RevertToSnapshot(snap)
		env.gasPool.SetGas(gp)

		log.Error("Miner apply transaction failed", "err", err)
		log.Error("Miner apply transaction failed", "tx", tx.Type())
		return receipt, err
	}

	//[rollup-geth] EIP-7706
	if miner.chainConfig.IsEIP7706(env.header.Number, env.header.Time) {
		log.Info("Miner", "setting gs used vector", receipt.GasUsedVector)
		env.header.GasUsedVector = receipt.GasUsedVector
	}

	return receipt, err
}

func (miner *Miner) checkInterop(ctx context.Context, tx *types.Transaction, failed bool, logs []*types.Log) error {
	if tx.Type() == types.DepositTxType {
		return nil // deposit-txs are always safe
	}
	if failed {
		return nil // failed txs don't persist any logs
	}
	if tx.Rejected() {
		return errors.New("transaction was previously rejected")
	}
	b, ok := miner.backend.(BackendWithInterop)
	if !ok {
		return fmt.Errorf("cannot mine interop txs without interop backend, got backend type %T", miner.backend)
	}
	if ctx == nil { // check if the miner was set up correctly to interact with an RPC
		return errors.New("need RPC context to check executing messages")
	}
	executingMessages, err := interoptypes.ExecutingMessagesFromLogs(logs)
	if err != nil {
		return fmt.Errorf("cannot parse interop messages from receipt of %s: %w", tx.Hash(), err)
	}
	if len(executingMessages) == 0 {
		return nil // avoid an RPC check if there are no executing messages to verify.
	}
	if err := b.CheckMessages(ctx, executingMessages, interoptypes.CrossUnsafe); err != nil {
		if ctx.Err() != nil { // don't reject transactions permanently on RPC timeouts etc.
			log.Debug("CheckMessages timed out", "err", ctx.Err())
			return err
		}
		txInteropRejectedCounter.Inc(1)
		tx.SetRejected() // Mark the tx as rejected: it will not be welcome in the tx-pool anymore.
		return err
	}
	return nil
}

func (miner *Miner) commitTransactions(env *environment, plainTxs, blobTxs *transactionsByPriceAndNonce, interrupt *atomic.Int32) error {
	gasLimit := env.header.GasLimit
	if env.gasPool == nil {
		env.gasPool = new(core.GasPool).AddGas(gasLimit)
	}
	blockDABytes := new(big.Int)
	for {
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
		// If we don't have enough blob space for any further blob transactions,
		// skip that list altogether
		if !blobTxs.Empty() && env.blobs*params.BlobTxBlobGasPerBlob >= params.MaxBlobGasPerBlock {
			log.Trace("Not enough blob space for further blob transactions")
			blobTxs.Clear()
			// Fall though to pick up any plain txs
		}
		// Retrieve the next transaction and abort if all done.
		var (
			ltx *txpool.LazyTransaction
			txs *transactionsByPriceAndNonce
		)
		pltx, ptip := plainTxs.Peek()
		bltx, btip := blobTxs.Peek()

		switch {
		case pltx == nil:
			txs, ltx = blobTxs, bltx
		case bltx == nil:
			txs, ltx = plainTxs, pltx
		default:
			if ptip.Lt(btip) {
				txs, ltx = blobTxs, bltx
			} else {
				txs, ltx = plainTxs, pltx
			}
		}
		if ltx == nil {
			break
		}
		// If we don't have enough space for the next transaction, skip the account.
		if env.gasPool.Gas() < ltx.Gas {
			log.Trace("Not enough gas left for transaction", "hash", ltx.Hash, "left", env.gasPool.Gas(), "needed", ltx.Gas)
			txs.Pop()
			continue
		}
		if left := uint64(params.MaxBlobGasPerBlock - env.blobs*params.BlobTxBlobGasPerBlob); left < ltx.BlobGas {
			log.Trace("Not enough blob gas left for transaction", "hash", ltx.Hash, "left", left, "needed", ltx.BlobGas)
			txs.Pop()
			continue
		}
		daBytesAfter := new(big.Int)
		if ltx.DABytes != nil && miner.config.MaxDABlockSize != nil {
			daBytesAfter.Add(blockDABytes, ltx.DABytes)
			if daBytesAfter.Cmp(miner.config.MaxDABlockSize) > 0 {
				log.Debug("adding tx would exceed block DA size limit",
					"hash", ltx.Hash, "txda", ltx.DABytes, "blockda", blockDABytes, "dalimit", miner.config.MaxDABlockSize)
				txs.Pop()
				continue
			}
		}
		// Transaction seems to fit, pull it up from the pool
		tx := ltx.Resolve()
		if tx == nil {
			log.Trace("Ignoring evicted transaction", "hash", ltx.Hash)
			txs.Pop()
			continue
		}
		// Error may be ignored here. The error has already been checked
		// during transaction acceptance in the transaction pool.
		from, _ := types.Sender(env.signer, tx)

		// Check whether the tx is replay protected. If we're not in the EIP155 hf
		// phase, start ignoring the sender until we do.
		if tx.Protected() && !miner.chainConfig.IsEIP155(env.header.Number) {
			log.Trace("Ignoring replay protected transaction", "hash", ltx.Hash, "eip155", miner.chainConfig.EIP155Block)
			txs.Pop()
			continue
		}
		// Start executing the transaction
		env.state.SetTxContext(tx.Hash(), env.tcount)

		err := miner.commitTransaction(env, tx)
		switch {
		case errors.Is(err, core.ErrNonceTooLow):
			// New head notification data race between the transaction pool and miner, shift
			log.Trace("Skipping transaction with low nonce", "hash", ltx.Hash, "sender", from, "nonce", tx.Nonce())
			txs.Shift()

		case errors.Is(err, errTxConditionalInvalid):
			// err contains contextual info on the failed conditional.
			txConditionalRejectedCounter.Inc(1)

			// mark as rejected so that it can be ejected from the mempool
			tx.SetRejected()
			log.Warn("Skipping account, transaction with failed conditional", "sender", from, "hash", ltx.Hash, "err", err)
			txs.Pop()

		case env.rpcCtx != nil && env.rpcCtx.Err() != nil && errors.Is(err, env.rpcCtx.Err()):
			log.Warn("Transaction processing aborted due to RPC context error", "err", err)
			txs.Pop() // RPC timeout. Tx could not be checked, and thus not included, but not rejected yet.

		case err != nil && tx.Rejected():
			log.Warn("Transaction was rejected during block-building", "hash", ltx.Hash, "err", err)
			txs.Pop()

		case errors.Is(err, nil):
			// Everything ok, collect the logs and shift in the next transaction from the same account
			blockDABytes = daBytesAfter
			txs.Shift()

		default:
			// Transaction is regarded as invalid, drop all consecutive transactions from
			// the same sender because of `nonce-too-high` clause.
			log.Debug("Transaction failed, account skipped", "hash", ltx.Hash, "err", err)
			txs.Pop()
		}
	}
	return nil
}

// fillTransactions retrieves the pending transactions from the txpool and fills them
// into the given sealing block. The transaction selection and ordering strategy can
// be customized with the plugin in the future.
func (miner *Miner) fillTransactions(interrupt *atomic.Int32, env *environment) error {
	miner.confMu.RLock()
	tip := miner.config.GasPrice
	miner.confMu.RUnlock()

	// Retrieve the pending transactions pre-filtered by the 1559/4844 dynamic fees
	filter := txpool.PendingFilter{
		MinTip:      uint256.MustFromBig(tip),
		MaxDATxSize: miner.config.MaxDATxSize,
	}

	if env.header.BaseFee != nil {
		filter.BaseFee = uint256.MustFromBig(env.header.BaseFee)
	}
	if env.header.ExcessBlobGas != nil {
		filter.BlobFee = uint256.MustFromBig(eip4844.CalcBlobFee(*env.header.ExcessBlobGas))
	}
	filter.OnlyPlainTxs, filter.OnlyBlobTxs = true, false
	pendingPlainTxs := miner.txpool.Pending(filter)

	filter.OnlyPlainTxs, filter.OnlyBlobTxs = false, true
	pendingBlobTxs := miner.txpool.Pending(filter)

	// Split the pending transactions into locals and remotes.
	localPlainTxs, remotePlainTxs := make(map[common.Address][]*txpool.LazyTransaction), pendingPlainTxs
	localBlobTxs, remoteBlobTxs := make(map[common.Address][]*txpool.LazyTransaction), pendingBlobTxs

	for _, account := range miner.txpool.Locals() {
		if txs := remotePlainTxs[account]; len(txs) > 0 {
			delete(remotePlainTxs, account)
			localPlainTxs[account] = txs
		}
		if txs := remoteBlobTxs[account]; len(txs) > 0 {
			delete(remoteBlobTxs, account)
			localBlobTxs[account] = txs
		}
	}
	// Fill the block with all available pending transactions.
	if len(localPlainTxs) > 0 || len(localBlobTxs) > 0 {
		plainTxs := newTransactionsByPriceAndNonce(env.signer, localPlainTxs, env.header.BaseFee)
		blobTxs := newTransactionsByPriceAndNonce(env.signer, localBlobTxs, env.header.BaseFee)

		if err := miner.commitTransactions(env, plainTxs, blobTxs, interrupt); err != nil {
			return err
		}
	}
	if len(remotePlainTxs) > 0 || len(remoteBlobTxs) > 0 {
		plainTxs := newTransactionsByPriceAndNonce(env.signer, remotePlainTxs, env.header.BaseFee)
		blobTxs := newTransactionsByPriceAndNonce(env.signer, remoteBlobTxs, env.header.BaseFee)

		if err := miner.commitTransactions(env, plainTxs, blobTxs, interrupt); err != nil {
			return err
		}
	}

	//[rollup-geth] EIP-7706
	filter.OnlyVectorFeeTxs, filter.OnlyPlainTxs, filter.OnlyBlobTxs = true, false, false
	pendingVectorTxs := miner.txpool.Pending(filter)
	if err := miner.commitVectorFeeTransactions(env, pendingVectorTxs, interrupt); err != nil {
		return err
	}

	return nil
}

// totalFees computes total consumed miner fees in Wei. Block transactions and receipts have to have the same order.
func totalFees(block *types.Block, receipts []*types.Receipt) *big.Int {
	feesWei := new(big.Int)
	for i, tx := range block.Transactions() {
		minerFee, _ := tx.EffectiveGasTip(block.BaseFee())
		feesWei.Add(feesWei, new(big.Int).Mul(new(big.Int).SetUint64(receipts[i].GasUsed), minerFee))
	}
	return feesWei
}

// signalToErr converts the interruption signal to a concrete error type for return.
// The given signal must be a valid interruption signal.
func signalToErr(signal int32) error {
	switch signal {
	case commitInterruptNewHead:
		return errBlockInterruptedByNewHead
	case commitInterruptResubmit:
		return errBlockInterruptedByRecommit
	case commitInterruptTimeout:
		return errBlockInterruptedByTimeout
	case commitInterruptResolve:
		return errBlockInterruptedByResolve
	default:
		panic(fmt.Errorf("undefined signal %d", signal))
	}
}

// validateParams validates the given parameters.
// It currently checks that the parent block is known and that the timestamp is valid,
// i.e., after the parent block's timestamp.
// It returns an upper bound of the payload building duration as computed
// by the difference in block timestamps between the parent and genParams.
func (miner *Miner) validateParams(genParams *generateParams) (time.Duration, error) {
	miner.confMu.RLock()
	defer miner.confMu.RUnlock()

	// Find the parent block for sealing task
	parent := miner.chain.CurrentBlock()
	if genParams.parentHash != (common.Hash{}) {
		block := miner.chain.GetBlockByHash(genParams.parentHash)
		if block == nil {
			return 0, fmt.Errorf("missing parent %v", genParams.parentHash)
		}
		parent = block.Header()
	}

	// Sanity check the timestamp correctness
	blockTime := int64(genParams.timestamp) - int64(parent.Time)
	if blockTime <= 0 && genParams.forceTime {
		return 0, fmt.Errorf("invalid timestamp, parent %d given %d", parent.Time, genParams.timestamp)
	}

	// minimum payload build time of 2s
	if blockTime < 2 {
		blockTime = 2
	}
	return time.Duration(blockTime) * time.Second, nil
}
