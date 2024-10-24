package eip7706

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func CalcBaseFee(config *params.ChainConfig, parent *types.Header) types.VectorFeeBigint {
	return types.NewVectorFeeBigInt()
}

func VerifyEIP7706Header(config *params.ChainConfig, parent, header *types.Header) error {
	return nil
}
