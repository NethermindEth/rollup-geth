package vm

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type MockL1RPCClient struct{}

func (m MockL1RPCClient) StoragesAt(ctx context.Context, account common.Address, keys []common.Hash, blockNumber *big.Int) ([]byte, error) {
	// testcase is in format "abab", this makes output lenght 2 bytes
	const mockedRespValueSize = 2
	mockResp := make([]byte, mockedRespValueSize*len(keys))
	for i := range keys {
		copy(mockResp[mockedRespValueSize*i:], common.Hex2Bytes("abab"))
	}

	return mockResp, nil
}

func TestPrecompiledL1SLOAD(t *testing.T) {
	mockL1RPCClient := MockL1RPCClient{}

	allPrecompiles[rollupL1SloadAddress] = &L1SLoad{}
	allPrecompiles.activateL1SLoad(mockL1RPCClient, func() *big.Int { return big1 })

	testJson("l1sload", rollupL1SloadAddress.Hex(), t)
	testJsonFail("l1sload", rollupL1SloadAddress.Hex(), t)
}