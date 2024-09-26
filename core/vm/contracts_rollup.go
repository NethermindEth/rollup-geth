//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

var rollupL1SloadAddress = common.BytesToAddress([]byte{0x10, 0x01})

var PrecompiledContractsRollupR0 = PrecompiledContracts{
	rollupL1SloadAddress: &L1SLoad{},
}

type RollupPrecompileActivationConfig struct {
	L1SLoad
}

func activeRollupPrecompiledContracts(rules params.Rules) PrecompiledContracts {
	switch rules.IsR0 {
	case rules.IsR0:
		return PrecompiledContractsRollupR0
	default:
		return nil
	}
}

func (pc *PrecompiledContracts) ActivateRollupPrecompiledContracts(config RollupPrecompileActivationConfig) {
	// NOTE: if L1SLoad was not activated via chain rules this is no-op
	pc.activateL1SLoad(config.L1RpcClient, config.GetLatestL1BlockNumber)
}

func (evm *EVM) activateRollupPrecompiledContracts() {
	evm.precompiles.ActivateRollupPrecompiledContracts(RollupPrecompileActivationConfig{
		L1SLoad{L1RpcClient: evm.Config.L1RpcClient, GetLatestL1BlockNumber: evm.rollupPrecompileOverrides.l1SLoadGetLatestL1Block},
	})
}

type L1SLoad struct {
	L1RpcClient            L1RpcClient
	GetLatestL1BlockNumber func() *big.Int
}

func (c *L1SLoad) RequiredGas(input []byte) uint64 { return 0 }

func (c *L1SLoad) Run(input []byte) ([]byte, error) {
	if !c.isL1SLoadActive() {
		return nil, errors.New("L1SLoad precompile not active")
	}

	return nil, nil
}

func (c *L1SLoad) isL1SLoadActive() bool {
	return c.GetLatestL1BlockNumber != nil && c.L1RpcClient != nil
}

func (pc *PrecompiledContracts) activateL1SLoad(l1RpcClient L1RpcClient, getLatestL1BlockNumber func() *big.Int) {
	rulesSayContractShouldBeActive := (*pc)[rollupL1SloadAddress] != nil
	paramsNotNil := l1RpcClient != nil && getLatestL1BlockNumber != nil

	if shouldActivateL1SLoad := rulesSayContractShouldBeActive && paramsNotNil; shouldActivateL1SLoad {
		(*pc)[rollupL1SloadAddress] = &L1SLoad{
			L1RpcClient:            l1RpcClient,
			GetLatestL1BlockNumber: getLatestL1BlockNumber,
		}
	}
}
