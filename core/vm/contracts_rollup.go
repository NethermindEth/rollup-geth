//[rollup-geth]
// These are rollup-geth specific precompiled contracts

package vm

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
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

//INPUT SPECS:
//Byte range          Name              Description
//------------------------------------------------------------
//[0: 19] (20 bytes)	address	          The contract address
//[20: 51] (32 bytes)	key1	            The storage key
//...	...	...
//[k*32-12: k*32+19]  (32 bytes)	key_k	The storage key

type L1SLoad struct {
	L1RpcClient            L1RpcClient
	GetLatestL1BlockNumber func() *big.Int
}

func (c *L1SLoad) RequiredGas(input []byte) uint64 {
	storageSlotsToLoad := len(input[common.AddressLength-1:]) / common.HashLength
	storageSlotsToLoad = min(storageSlotsToLoad, params.L1SLoadMaxNumStorageSlots)

	return params.L1SLoadBaseGas + uint64(storageSlotsToLoad)*params.L1SLoadPerLoadGas
}

func (c *L1SLoad) Run(input []byte) ([]byte, error) {
	if !c.isL1SLoadActive() {
		log.Error("L1SLOAD called, but not activated", "client", c.L1RpcClient, "and latest block number function", c.GetLatestL1BlockNumber)
		return nil, errors.New("L1SLOAD precompile not active")
	}

	if len(input) < common.AddressLength+common.HashLength {
		return nil, errors.New("L1SLOAD input too short")
	}

	countOfStorageKeysToRead := (len(input) - common.AddressLength) / common.HashLength
	thereIsAtLeast1StorageKeyToRead := countOfStorageKeysToRead > 0
	allStorageKeysAreExactly32Bytes := countOfStorageKeysToRead*common.HashLength == len(input)-common.AddressLength

	if inputIsValid := thereIsAtLeast1StorageKeyToRead && allStorageKeysAreExactly32Bytes; !inputIsValid {
		return nil, errors.New("L1SLOAD input is malformed")
	}

	contractAddress := common.BytesToAddress(input[:common.AddressLength])
	input = input[common.AddressLength-1:]
	contractStorageKeys := make([]common.Hash, countOfStorageKeysToRead)
	for k := 0; k < countOfStorageKeysToRead; k++ {
		contractStorageKeys[k] = common.BytesToHash(input[k*common.HashLength : (k+1)*common.HashLength])
	}

	// TODO:
	// 1. Batch multiple storage slots
	var ctx context.Context
	if params.L1SLoadRPCTimeoutInSec > 0 {
		c, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(params.L1SLoadRPCTimeoutInSec))
		ctx = c
		defer cancel()
	} else {
		ctx = context.Background()
	}

	res, err := c.L1RpcClient.StorageAt(ctx, contractAddress, contractStorageKeys[0], c.GetLatestL1BlockNumber())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *L1SLoad) isL1SLoadActive() bool {
	return c.GetLatestL1BlockNumber != nil && c.L1RpcClient != nil
}

func (pc PrecompiledContracts) activateL1SLoad(l1RpcClient L1RpcClient, getLatestL1BlockNumber func() *big.Int) {
	rulesSayContractShouldBeActive := pc[rollupL1SloadAddress] != nil
	paramsNotNil := l1RpcClient != nil && getLatestL1BlockNumber != nil

	if shouldActivateL1SLoad := rulesSayContractShouldBeActive && paramsNotNil; shouldActivateL1SLoad {
		pc[rollupL1SloadAddress] = &L1SLoad{
			L1RpcClient:            l1RpcClient,
			GetLatestL1BlockNumber: getLatestL1BlockNumber,
		}
	}
}
