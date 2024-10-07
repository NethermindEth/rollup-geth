package ethapi

import "github.com/ethereum/go-ethereum/core/vm"

func (b *testBackend) GetL1RpcClient() vm.L1RpcClient {
	return nil
}
