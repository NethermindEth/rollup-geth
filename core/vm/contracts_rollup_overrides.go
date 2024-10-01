//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/state"
)

// generateRollupPrecompiledContractsOverrides generates rollup precompile config including L2 specific overrides
func generateRollupPrecompiledContractsOverrides(evm *EVM) *RollupPrecompileActivationConfig {
	return &RollupPrecompileActivationConfig{
		L1SLoad{
			L1RpcClient:            evm.Config.L1RpcClient,
			GetLatestL1BlockNumber: LetRPCDecideLatestL1Number(),
		},
	}
}

// [OVERRIDE]  LetRPCDecideLatestL1Number
// Each rollup should override this function so that it returns
// correct latest L1 block number
func LetRPCDecideLatestL1Number() func() *big.Int {
	return func() *big.Int {
		return nil
	}
}

// [OVERRIDE]  getLatestL1BlockNumber
// Each rollup should override this function so that it returns
// correct latest L1 block number
//
// EXAMPLE 2
func GetLatestL1BlockNumber(state *state.StateDB) func() *big.Int {
	return func() *big.Int {
		addressOfL1BlockContract := common.Address{}
		slotInContractRepresentingL1BlockNumber := common.Hash{}
		return state.GetState(addressOfL1BlockContract, slotInContractRepresentingL1BlockNumber).Big()
	}
}
