//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

type RollupPrecompiledContractsOverrides struct {
	l1SLoadGetLatestL1Block func() *big.Int
}

func GenerateRollupPrecompiledContractsOverrides(evm *EVM) RollupPrecompiledContractsOverrides {
	return RollupPrecompiledContractsOverrides{
		l1SLoadGetLatestL1Block: getLatestL1BlockNumber(evm),
	}
}

var rollupL1SloadAddress = common.BytesToAddress([]byte{0x10, 0x01})

var PrecompiledContractsRollupR0 = PrecompiledContracts{
	rollupL1SloadAddress: &l1SLoad{},
}

func activeRollupPrecompiledContracts(rules params.Rules) PrecompiledContracts {
	switch rules.IsR0 {
	case rules.IsR0:
		return PrecompiledContractsRollupR0
	default:
		return nil
	}
}

func (evm *EVM) activateRollupPrecompiledContracts() {
	activeRollupPrecompiles := activeRollupPrecompiledContracts(evm.chainRules)
	for k, v := range activeRollupPrecompiles {
		evm.precompiles[k] = v
	}

	// NOTE: if L1SLoad was not activated via chain rules this is no-op
	evm.precompiles.activateL1SLoad(evm.Config.L1RpcClient, evm.rollupPrecompileOverrides.l1SLoadGetLatestL1Block)
}

type l1SLoad struct {
	l1RpcClient            L1Client
	getLatestL1BlockNumber func() *big.Int
}

func (c *l1SLoad) RequiredGas(input []byte) uint64 { return 0 }

func (c *l1SLoad) Run(input []byte) ([]byte, error) {
	if !c.isL1SLoadActive() {
		return nil, errors.New("L1SLoad precompile not active")
	}

	return nil, nil
}

func (c *l1SLoad) isL1SLoadActive() bool {
	return c.getLatestL1BlockNumber != nil && c.l1RpcClient != nil
}

func (pc *PrecompiledContracts) activateL1SLoad(l1RpcClient L1Client, getLatestL1BlockNumber func() *big.Int) {
	if (*pc)[rollupL1SloadAddress] != nil {
		(*pc)[rollupL1SloadAddress] = &l1SLoad{
			l1RpcClient:            l1RpcClient,
			getLatestL1BlockNumber: getLatestL1BlockNumber,
		}
	}
}
