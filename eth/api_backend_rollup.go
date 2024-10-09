package eth

import "github.com/ethereum/go-ethereum/core/vm"

func (b *EthAPIBackend) GetL1RpcClient() vm.L1RpcClient {
	return b.eth.BlockChain().GetVMConfig().L1RpcClient
}
