package main

import (
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/urfave/cli/v2"
)

// TODO: when we have clearer picture of how do we want rollup "features" (EIPs/RIPs) to be activated
// make this "rule" activated (ie. if not "rule activated" then L1 client can simply be nil)
func activateL1RPCEndpoint(ctx *cli.Context, stack *node.Node) {
	if !ctx.IsSet(utils.L1NodeRPCEndpointFlag.Name) {
		log.Error("L1 node RPC endpoint URL not set", "flag", utils.L1NodeRPCEndpointFlag.Name)
		return
	}

	l1RPCEndpoint := ctx.String(utils.L1NodeRPCEndpointFlag.Name)
	stack.RegisterEthClient(l1RPCEndpoint)
}
