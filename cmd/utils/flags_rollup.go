package utils

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/internal/flags"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
)

var L1NodeRPCEndpointFlag = &cli.StringFlag{
	Name:     "rollup.l1.rpc_endpoint",
	Usage:    "L1 node RPC endpoint eg. http://0.0.0.0:8545",
	Category: flags.RollupCategory,
}

var RollupFlags = []cli.Flag{
	L1NodeRPCEndpointFlag,
}

// TODO: when we have clearer picture of how do we want rollup "features" (EIPs/RIPs) to be activated
// make this "rule" activated (ie. if not "rule activated" then L1 client can simply be nil)
func ActivateL1RPCEndpoint(ctx *cli.Context, stack *node.Node) {
	if !ctx.IsSet(L1NodeRPCEndpointFlag.Name) {
		log.Error("L1 node RPC endpoint URL not set", "flag", L1NodeRPCEndpointFlag.Name)
		return
	}

	l1RPCEndpoint := ctx.String(L1NodeRPCEndpointFlag.Name)
	ethClient, err := ethclient.Dial(l1RPCEndpoint)
	if err != nil {
		log.Error("Unable to connect to ETH RPC endpoint at", "URL", ethClient, "error", err)
		return
	}

	vm.SetVmL1RpcClient(ethClient)
	log.Info("Initialized ETH RPC client", "endpoint", ethClient)
}
