package ethapi

import "github.com/ethereum/go-ethereum/core/vm"

func (b *backendMock) GetL1RpcClient() vm.L1RpcClient {
	return nil
}
