package eth

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// TODO: when we have clearer picture of how do we want rollup "features" (EIPs/RIPs) to be activated
// make this "rule" activated (ie. if not "rule activated" then L1 client can simply be nil)
func activateL1RPCEndpoint(l1RPCEndpoint string, vmConfig *vm.Config) {
	l1Client, err := ethclient.Dial(l1RPCEndpoint)
	if err != nil {
		log.Crit("Unable to connect to L1 RPC endpoint at", "URL", l1RPCEndpoint, "error", err)
	} else {
		vmConfig.L1RpcClient = l1Client
		log.Info("Initialized L1 RPC client", "endpoint", l1RPCEndpoint)
	}
}
