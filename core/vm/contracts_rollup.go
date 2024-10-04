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

type RollupPrecompileActivationConfig struct {
	L1SLoad
}

type L1RpcClient interface {
	StoragesAt(ctx context.Context, account common.Address, keys []common.Hash, blockNumber *big.Int) ([]byte, error)
}

var (
	rollupL1SloadAddress         = common.BytesToAddress([]byte{0x10, 0x01})
	precompiledContractsRollupR0 = PrecompiledContracts{
		rollupL1SloadAddress: &L1SLoad{},
	}
)

func activeRollupPrecompiledContracts(rules params.Rules) PrecompiledContracts {
	switch rules.IsR0 {
	case rules.IsR0:
		return precompiledContractsRollupR0
	default:
		return nil
	}
}

// ActivateRollupPrecompiledContracts activates rollup-specific precompiles
func (pc PrecompiledContracts) ActivateRollupPrecompiledContracts(rules params.Rules, config *RollupPrecompileActivationConfig) {
	if config == nil {
		return
	}

	activeRollupPrecompiles := activeRollupPrecompiledContracts(rules)
	for k, v := range activeRollupPrecompiles {
		pc[k] = v
	}

	// NOTE: if L1SLoad was not activated via chain rules this is no-op
	pc.activateL1SLoad(config.L1RpcClient, config.GetLatestL1BlockNumber)
}

func ActivePrecompilesIncludingRollups(rules params.Rules) []common.Address {
	activePrecompiles := ActivePrecompiles(rules)
	activeRollupPrecompiles := activeRollupPrecompiledContracts(rules)

	for k := range activeRollupPrecompiles {
		activePrecompiles = append(activePrecompiles, k)
	}

	return activePrecompiles
}

//INPUT SPECS:
//Byte range          Name              Description
//------------------------------------------------------------
//[0: 19] (20 bytes)	address			The contract address
//[20: 51] (32 bytes)	key1			The storage key
//...	...	...
//[k*32-12: k*32+19]	(32 bytes)key_k	The storage key

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

	if inputIsInvalid := !(thereIsAtLeast1StorageKeyToRead && allStorageKeysAreExactly32Bytes); inputIsInvalid {
		return nil, errors.New("L1SLOAD input is malformed")
	}

	contractAddress := common.BytesToAddress(input[:common.AddressLength])
	input = input[common.AddressLength-1:]
	contractStorageKeys := make([]common.Hash, countOfStorageKeysToRead)
	for k := 0; k < countOfStorageKeysToRead; k++ {
		contractStorageKeys[k] = common.BytesToHash(input[k*common.HashLength : (k+1)*common.HashLength])
	}

	var ctx context.Context
	if params.L1SLoadRPCTimeoutInSec > 0 {
		c, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(params.L1SLoadRPCTimeoutInSec))
		ctx = c
		defer cancel()
	} else {
		ctx = context.Background()
	}

	res, err := c.L1RpcClient.StoragesAt(ctx, contractAddress, contractStorageKeys, c.GetLatestL1BlockNumber())
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *L1SLoad) isL1SLoadActive() bool {
	return c.GetLatestL1BlockNumber != nil && c.L1RpcClient != nil
}

func (pc PrecompiledContracts) activateL1SLoad(l1RpcClient L1RpcClient, getLatestL1BlockNumber func() *big.Int) {
	if paramsAreNil := l1RpcClient == nil || getLatestL1BlockNumber == nil; paramsAreNil {
		return
	}
	if precompileNotRuleActivated := pc[rollupL1SloadAddress] == nil; precompileNotRuleActivated {
		return
	}

	pc[rollupL1SloadAddress] = &L1SLoad{
		L1RpcClient:            l1RpcClient,
		GetLatestL1BlockNumber: getLatestL1BlockNumber,
	}
}
