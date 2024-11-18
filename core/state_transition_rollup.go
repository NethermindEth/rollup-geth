package core

import (
	"errors"
	"fmt"
	"math/big"

	// cmath "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var OverflowError = errors.New("value overflowed")

func TransactionToMessage(tx *types.Transaction, s types.Signer, h *types.Header, chainConfig *params.ChainConfig) (*Message, error) {
	if chainConfig.IsEIP7706(h.Number, h.Time) {
		if h.BaseFees == nil {
			return nil, fmt.Errorf("base fees not set")
		}
		return TransactionToMessageEIP7706(tx, s, h.BaseFees)
	}

	return TransactionToMessageEIP4844(tx, s, h.GetBaseFee())
}

// TransactionToMessage converts a transaction into a Message post EIP-7706
func TransactionToMessageEIP7706(tx *types.Transaction, s types.Signer, baseFees types.VectorFeeBigint) (*Message, error) {
	msg := &Message{
		Nonce:            tx.Nonce(),
		GasLimit:         tx.Gas(),
		GasFeeCap:        new(big.Int).Set(tx.GasFeeCap()),
		GasTipCap:        new(big.Int).Set(tx.GasTipCap()),
		To:               tx.To(),
		Value:            tx.Value(),
		Data:             tx.Data(),
		AccessList:       tx.AccessList(),
		SkipNonceChecks:  false,
		SkipFromEOACheck: false,
		BlobHashes:       tx.BlobHashes(),
		BlobGasFeeCap:    tx.BlobGasFeeCap(),

		GasFeeCaps:         tx.GasFeeCaps(),
		GasTipCaps:         tx.GasTipCaps(),
		GasLimits:          tx.GasLimits(),
		EffectiveGasTips:   tx.EffectiveGasTips(baseFees),
		EffectiveGasPrices: tx.EffectiveGasPrices(baseFees),
	}

	// TODO: [rollup-geth] EIP-7706 check this
	//Do we indeed set msg.GasPrice to execution price? or?
	msg.GasPrice = tx.EffectiveGasExecutionPrice(baseFees)

	var err error
	msg.From, err = types.Sender(s, tx)
	return msg, err
}

func (st *StateTransition) preCheckGas() error {
	if st.evm.ChainConfig().IsEIP7706(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		return st.preCheckGasEIP7706()
	}

	return st.preCheckGasEIP4484()
}

func (st *StateTransition) buyGas() error {
	if st.evm.ChainConfig().IsEIP7706(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		return st.buyGasEIP7706()
	}

	return st.buyGasEIP4844()
}

func (st *StateTransition) refundGasToAddress() error {
	if st.evm.ChainConfig().IsEIP7706(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		return st.refundGasToAddressEIP7706()
	}

	return st.refundGasToAddressEIP4844()
}

func (st *StateTransition) payTheTip(rules params.Rules, msg *Message) error {
	// Check that we are post bedrock to enable op-geth to be able to create pseudo pre-bedrock blocks (these are pre-bedrock, but don't follow l2 geth rules)
	// Note optimismConfig will not be nil if rules.IsOptimismBedrock is true
	if optimismConfig := st.evm.ChainConfig().Optimism; optimismConfig != nil && rules.IsOptimismBedrock && !st.msg.IsDepositTx {
		gasCost := new(big.Int).Mul(new(big.Int).SetUint64(st.gasUsed()), st.evm.Context.BaseFee)
		amtU256, overflow := uint256.FromBig(gasCost)
		if overflow {
			return fmt.Errorf("optimism gas cost overflows U256: %d", gasCost)
		}
		st.state.AddBalance(params.OptimismBaseFeeRecipient, amtU256, tracing.BalanceIncreaseRewardTransactionFee)
		if l1Cost := st.evm.Context.L1CostFunc(st.msg.RollupCostData, st.evm.Context.Time); l1Cost != nil {
			amtU256, overflow = uint256.FromBig(l1Cost)
			if overflow {
				return fmt.Errorf("optimism l1 cost overflows U256: %d", l1Cost)
			}
			st.state.AddBalance(params.OptimismL1FeeRecipient, amtU256, tracing.BalanceIncreaseRewardTransactionFee)
		}
	}

	if st.evm.ChainConfig().IsEIP7706(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		return st.payTheTipEIP7706(rules, msg)
	}

	return st.payTheTipEIP4844(rules, msg)
}

func (st *StateTransition) buyGasEIP4844() error {
	mgval := new(big.Int).SetUint64(st.msg.GasLimit)
	mgval.Mul(mgval, st.msg.GasPrice)

	var l1Cost *big.Int
	if st.evm.Context.L1CostFunc != nil && !st.msg.SkipNonceChecks && !st.msg.SkipFromEOACheck {
		l1Cost = st.evm.Context.L1CostFunc(st.msg.RollupCostData, st.evm.Context.Time)
		if l1Cost != nil {
			mgval = mgval.Add(mgval, l1Cost)
		}
	}

	balanceCheck := new(big.Int).Set(mgval)
	if st.msg.GasFeeCap != nil {
		balanceCheck.SetUint64(st.msg.GasLimit)
		balanceCheck = balanceCheck.Mul(balanceCheck, st.msg.GasFeeCap)
		if l1Cost != nil {
			balanceCheck.Add(balanceCheck, l1Cost)
		}
	}
	balanceCheck.Add(balanceCheck, st.msg.Value)

	if st.evm.ChainConfig().IsCancun(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		if blobGas := st.blobGasUsed(); blobGas > 0 {
			// Check that the user has enough funds to cover blobGasUsed * tx.BlobGasFeeCap
			blobBalanceCheck := new(big.Int).SetUint64(blobGas)
			blobBalanceCheck.Mul(blobBalanceCheck, st.msg.BlobGasFeeCap)
			balanceCheck.Add(balanceCheck, blobBalanceCheck)
			// Pay for blobGasUsed * actual blob fee
			blobFee := new(big.Int).SetUint64(blobGas)
			blobFee.Mul(blobFee, st.evm.Context.BlobBaseFee)
			mgval.Add(mgval, blobFee)
		}
	}
	balanceCheckU256, overflow := uint256.FromBig(balanceCheck)
	if overflow {
		return fmt.Errorf("%w: address %v required balance exceeds 256 bits", ErrInsufficientFunds, st.msg.From.Hex())
	}
	if have, want := st.state.GetBalance(st.msg.From), balanceCheckU256; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From.Hex(), have, want)
	}
	if err := st.gp.SubGas(st.msg.GasLimit); err != nil {
		return err
	}

	if st.evm.Config.Tracer != nil && st.evm.Config.Tracer.OnGasChange != nil {
		st.evm.Config.Tracer.OnGasChange(0, st.msg.GasLimit, tracing.GasChangeTxInitialBalance)
	}
	st.gasRemaining = st.msg.GasLimit

	st.initialGas = st.msg.GasLimit
	mgvalU256, overflowed := uint256.FromBig(mgval)
	if overflowed {
		return fmt.Errorf("EIP-4844 %w", OverflowError)
	}
	st.state.SubBalance(st.msg.From, mgvalU256, tracing.BalanceDecreaseGasBuy)
	return nil
}

func (st *StateTransition) buyGasEIP7706() error {
	gasLimits := st.msg.GasLimits.ToVectorBigInt()

	// User should be able to cover GAS_LIMIT * MAX_FEE_PER_GAS + tx.value
	maxGasFees, err := gasLimits.VectorMul(st.msg.GasFeeCaps)
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate max gas fees vector: %w", err)
	}

	balanceCheck, err := maxGasFees.Sum()
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate max gas fees sum: %w", err)
	}

	//[op-geth] the address must be able to cover l1 fees
	var l1Cost *big.Int
	if st.evm.Context.L1CostFunc != nil && !st.msg.SkipNonceChecks && !st.msg.SkipFromEOACheck {
		l1Cost = st.evm.Context.L1CostFunc(st.msg.RollupCostData, st.evm.Context.Time)
		if l1Cost != nil {
			balanceCheck = balanceCheck.Add(balanceCheck, l1Cost)
		}
	}

	balanceCheck.Add(balanceCheck, st.msg.Value)
	balanceCheckU256, overflow := uint256.FromBig(balanceCheck)

	if overflow {
		return fmt.Errorf("EIP-7706: %w: address %v required balance exceeds 256 bits", ErrInsufficientFunds, st.msg.From.Hex())
	}
	if have, want := st.state.GetBalance(st.msg.From), balanceCheckU256; have.Cmp(want) < 0 {
		return fmt.Errorf("EIP-7706: %w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From.Hex(), have, want)
	}

	// NOTE: calculations below, which rely on msg.GasLimit, are still valid
	// This is because per EIP-7706 will still have Gas(Limit) as TX field
	// Which is in fact gas execution limit, so msg.GasLimits[0] == msg.GasLimit
	if err := st.gp.SubGas(st.msg.GasLimit); err != nil {
		return fmt.Errorf("EIP-7706 failed to subtract gas: %w", err)
	}

	if st.evm.Config.Tracer != nil && st.evm.Config.Tracer.OnGasChange != nil {
		st.evm.Config.Tracer.OnGasChange(0, st.msg.GasLimit, tracing.GasChangeTxInitialBalance)
	}

	st.gasRemaining = st.msg.GasLimit
	st.initialGas = st.msg.GasLimit

	// GAS_LIMIT * ACTUAL_FEE_PER_GAS
	totalGasFeesVec, err := st.msg.EffectiveGasPrices.VectorMul(gasLimits)
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate total gas fees vector: %w", err)
	}

	totalGasFees, err := totalGasFeesVec.Sum()
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate total gas fees sum: %w", err)
	}
	totalGasFeesU256, overflowed := uint256.FromBig(totalGasFees)
	if overflowed {
		return fmt.Errorf("EIP-7706 %w", OverflowError)
	}

	st.state.SubBalance(st.msg.From, totalGasFeesU256, tracing.BalanceDecreaseGasBuy)
	return nil
}

func (st *StateTransition) preCheckGasEIP4484() error {
	msg := st.msg
	// Make sure that transaction gasFeeCap is greater than the baseFee (post london)
	if st.evm.ChainConfig().IsLondon(st.evm.Context.BlockNumber) {
		// Skip the checks if gas fields are zero and baseFee was explicitly disabled (eth_call)
		skipCheck := st.evm.Config.NoBaseFee && msg.GasFeeCap.BitLen() == 0 && msg.GasTipCap.BitLen() == 0
		if !skipCheck {
			if l := msg.GasFeeCap.BitLen(); l > 256 {
				return fmt.Errorf("%w: address %v, maxFeePerGas bit length: %d", ErrFeeCapVeryHigh,
					msg.From.Hex(), l)
			}
			if l := msg.GasTipCap.BitLen(); l > 256 {
				return fmt.Errorf("%w: address %v, maxPriorityFeePerGas bit length: %d", ErrTipVeryHigh,
					msg.From.Hex(), l)
			}
			if msg.GasFeeCap.Cmp(msg.GasTipCap) < 0 {
				return fmt.Errorf("%w: address %v, maxPriorityFeePerGas: %s, maxFeePerGas: %s", ErrTipAboveFeeCap,
					msg.From.Hex(), msg.GasTipCap, msg.GasFeeCap)
			}
			// This will panic if baseFee is nil, but base fee presence is verified
			// as part of header validation.
			if msg.GasFeeCap.Cmp(st.evm.Context.BaseFee) < 0 {
				return fmt.Errorf("%w: address %v, maxFeePerGas: %s, baseFee: %s", ErrFeeCapTooLow,
					msg.From.Hex(), msg.GasFeeCap, st.evm.Context.BaseFee)
			}
		}
	}

	// Check that the user is paying at least the current blob fee
	if st.evm.ChainConfig().IsCancun(st.evm.Context.BlockNumber, st.evm.Context.Time) {
		if st.blobGasUsed() > 0 {
			// Skip the checks if gas fields are zero and blobBaseFee was explicitly disabled (eth_call)
			skipCheck := st.evm.Config.NoBaseFee && msg.BlobGasFeeCap.BitLen() == 0
			if !skipCheck {
				// This will panic if blobBaseFee is nil, but blobBaseFee presence
				// is verified as part of header validation.
				if msg.BlobGasFeeCap.Cmp(st.evm.Context.BlobBaseFee) < 0 {
					return fmt.Errorf("%w: address %v blobGasFeeCap: %v, blobBaseFee: %v", ErrBlobFeeCapTooLow,
						msg.From.Hex(), msg.BlobGasFeeCap, st.evm.Context.BlobBaseFee)
				}
			}
		}
	}

	return nil
}

func (st *StateTransition) preCheckGasEIP7706() error {
	msg := st.msg
	// Skip the checks if gas fields are zero and baseFee was explicitly disabled (eth_call)
	skipCheck := st.evm.Config.NoBaseFee && msg.GasFeeCaps.VecBitLenAllZero() && msg.GasTipCaps.VecBitLenAllZero()
	if skipCheck {
		return nil
	}

	if !msg.GasFeeCaps.VecBitLenAllLessEqThan256() {
		return fmt.Errorf("EIP-7706: %w: address %v", ErrFeeCapVeryHigh, msg.From.Hex())
	}
	if !msg.GasTipCaps.VecBitLenAllLessEqThan256() {
		return fmt.Errorf("EIP-7706: %w: address %v", ErrTipVeryHigh, msg.From.Hex())
	}
	if !msg.GasTipCaps.VectorAllLessOrEqual(msg.GasFeeCaps) {
		return fmt.Errorf("EIP-7706: %w: address %v", ErrTipAboveFeeCap, msg.From.Hex())
	}
	if !st.evm.Context.BaseFees.VectorAllLessOrEqual(msg.GasFeeCaps) {
		return fmt.Errorf("EIP-7706: %w: address %v", ErrFeeCapTooLow, msg.From.Hex())
	}

	return nil
}

func (st *StateTransition) refundGasToAddressEIP4844() error {
	gasFeeToRefund := uint256.NewInt(st.gasRemaining)
	gasFeeToRefund.Mul(gasFeeToRefund, uint256.MustFromBig(st.msg.GasPrice))
	st.state.AddBalance(st.msg.From, gasFeeToRefund, tracing.BalanceIncreaseGasReturn)

	return nil
}

func (st *StateTransition) refundGasToAddressEIP7706() error {
	gasToRefund := st.vectorGasRemaining().ToVectorBigInt()
	gasFeeToRefundVec, err := gasToRefund.VectorMul(st.msg.EffectiveGasPrices)
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate gas fees refund vector: %w", err)
	}

	gasFeeToRefund, err := gasFeeToRefundVec.Sum()
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate refund gas fee sum: %w", err)
	}

	st.state.AddBalance(st.msg.From, uint256.MustFromBig(gasFeeToRefund), tracing.BalanceIncreaseGasReturn)

	return nil
}

func (st *StateTransition) payTheTipEIP4844(rules params.Rules, msg *Message) error {
	if st.evm.Config.NoBaseFee && msg.GasFeeCap.Sign() == 0 && msg.GasTipCap.Sign() == 0 {
		// Skip fee payment when NoBaseFee is set and the fee fields
		// are 0. This avoids a negative effectiveTip being applied to
		// the coinbase when simulating calls.
		return nil
	}

	effectiveTip := msg.GasPrice
	if rules.IsLondon {
		effectiveTip = math.BigMin(msg.GasTipCap, new(big.Int).Sub(msg.GasFeeCap, st.evm.Context.BaseFee))
	}
	effectiveTipU256, didOVerflow := uint256.FromBig(effectiveTip)

	if didOVerflow {
		return fmt.Errorf("EIP-4844 %w", OverflowError)
	}

	fee := new(uint256.Int).SetUint64(st.gasUsed())
	fee.Mul(fee, effectiveTipU256)
	st.state.AddBalance(st.evm.Context.Coinbase, fee, tracing.BalanceIncreaseRewardTransactionFee)

	// add the coinbase to the witness iff the fee is greater than 0
	if rules.IsEIP4762 && fee.Sign() != 0 {
		st.evm.AccessEvents.AddAccount(st.evm.Context.Coinbase, true)
	}

	return nil
}

func (st *StateTransition) payTheTipEIP7706(rules params.Rules, msg *Message) error {
	if st.evm.Config.NoBaseFee && msg.GasTipCaps.VecBitLenAllZero() && msg.GasFeeCaps.VecBitLenAllZero() {
		// Skip fee payment when NoBaseFee is set and the fee fields
		// are 0. This avoids a negative effectiveTip being applied to
		// the coinbase when simulating calls.
		return nil
	}

	gasUsed := st.vectorGasUsed()
	gasTipsVec, err := msg.EffectiveGasTips.VectorMul(gasUsed.ToVectorBigInt())
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate gas tips vec: %w", err)
	}

	totalGasTip, err := gasTipsVec.Sum()
	if err != nil {
		return fmt.Errorf("EIP-7706 failed to calculate total gas tip: %w", err)
	}

	totalTip, overflowed := uint256.FromBig(totalGasTip)
	if overflowed {
		return fmt.Errorf("EIP-7706 %v", OverflowError)
	}

	st.state.AddBalance(st.evm.Context.Coinbase, totalTip, tracing.BalanceIncreaseRewardTransactionFee)

	// add the coinbase to the witness iff the fee is greater than 0
	if rules.IsEIP4762 && totalTip.Sign() != 0 {
		st.evm.AccessEvents.AddAccount(st.evm.Context.Coinbase, true)
	}

	return nil
}

func (st *StateTransition) vectorGasUsed() types.VectorGasLimit {
	// TODO: think if this should be "precalculated", that is set where we update st.gasRemaining
	// NOTE: Gas used by [execution, blob, calldata]
	// Blob and calldata gas used is actually same as their gas limits (because it is precomputed from tx data and known upfront)
	// Only execution gas is not known upfront and has to be determined while executing the transaction
	gasUsed := st.msg.GasLimits
	gasUsed[types.ExecutionGasIndex] = st.gasUsed()

	return gasUsed
}

func (st *StateTransition) vectorGasRemaining() types.VectorGasLimit {
	//NOTE: Per EIP-7706:
	//In practice, only the first term will be nonzero for now
	//This is because gas used by blob and calldata can be calculated upfront so we don't have any remaining calldata/blob gas

	//NOTE: 2 msg.GasLimits[0] == msg.GasLimit == st.initialGas
	// this is why this holds
	return st.msg.GasLimits.VectorSubtractClampAtZero(st.vectorGasUsed())
}
