package utils

import (
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/internal/flags"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	l1NodeRPCEndpointFlag = &cli.StringFlag{
		Name:     "rollup.l1.rpc_endpoint",
		Usage:    "L1 node RPC endpoint eg. http://0.0.0.0:8545",
		Category: flags.RollupCategory,
		Required: true,
	}
)

var (
	RollupFlags = []cli.Flag{
		l1NodeRPCEndpointFlag,
	}
)

// [rollup-geth]
func setRollupEthConfig(ctx *cli.Context, cfg *ethconfig.Config) {
	if ctx.IsSet(l1NodeRPCEndpointFlag.Name) {
		cfg.L1NodeRPCEndpoint = ctx.String(l1NodeRPCEndpointFlag.Name)
	} else {
		log.Crit("L1 node RPC endpoint URL not set", "flag", l1NodeRPCEndpointFlag.Name)
	}
}
