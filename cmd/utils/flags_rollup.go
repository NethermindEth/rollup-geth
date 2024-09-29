package utils

import (
	"github.com/ethereum/go-ethereum/internal/flags"
	"github.com/urfave/cli/v2"
)

var (
	L1NodeRPCEndpointFlag = &cli.StringFlag{
		Name:     "rollup.l1.rpc_endpoint",
		Usage:    "L1 node RPC endpoint eg. http://0.0.0.0:8545",
		Category: flags.RollupCategory,
		Required: true,
	}
)

var (
	RollupFlags = []cli.Flag{
		L1NodeRPCEndpointFlag,
	}
)
