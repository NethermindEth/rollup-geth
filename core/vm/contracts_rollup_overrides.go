//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import "math/big"

type RollupPrecompiledContractsOverrides struct {
	l1SLoadGetLatestL1Block func() *big.Int
}

func GenerateRollupPrecompiledContractsOverrides(evm *EVM) RollupPrecompiledContractsOverrides {
	return RollupPrecompiledContractsOverrides{
		l1SLoadGetLatestL1Block: getLatestL1BlockNumber(evm),
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
