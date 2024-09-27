package vm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type MockL1RPCClient struct{}

func (m MockL1RPCClient) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	return common.Hex2Bytes("abab"), nil
}

func TestPrecompiledL1SLOAD(t *testing.T) {
	mockL1RPCClient := MockL1RPCClient{}

	allPrecompiles[rollupL1SloadAddress] = &L1SLoad{}
	allPrecompiles.activateL1SLoad(mockL1RPCClient, func() *big.Int { return big1 })

	l1SLoadTestcase := precompiledTest{
		Name:        "L1SLOAD",
		Input:       "C02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc22d2c7bb6fc06067df8b0223aec460d1ebb51febb9012bc2554141a4dca08e864",
		Expected:    "abab",
		Gas:         4000,
		NoBenchmark: true,
	}

	testPrecompiled(rollupL1SloadAddress.Hex(), l1SLoadTestcase, t)
}
