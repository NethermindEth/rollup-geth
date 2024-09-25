package arbitrum

import "github.com/ethereum/go-ethereum/core/vm"

func (b *APIBackend) GetL1RpcClient() vm.L1RpcClient {
	return b.BlockChain().GetVMConfig().L1RpcClient
}
