//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import "math/big"

// generateRollupPrecompiledContractsOverrides generates rollup precompile config inlucing L2 specific overrides
func generateRollupPrecompiledContractsOverrides(evm *EVM) *RollupPrecompileActivationConfig {
	return &RollupPrecompileActivationConfig{
		L1SLoad{
			L1RpcClient:            evm.Config.L1RpcClient,
			GetLatestL1BlockNumber: getLatestL1BlockNumber(evm),
		},
	}
}

// [OVERRIDE]  getLatestL1BlockNumber
// Each rollup should override this function so that it returns
// correct latest L1 block number
func getLatestL1BlockNumber(evm *EVM) func() *big.Int {
	return func() *big.Int {
		return evm.Context.BlockNumber
	}
}
