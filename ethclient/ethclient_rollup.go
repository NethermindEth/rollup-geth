package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// StoragesAt returns the values of keys in the contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) StoragesAt(ctx context.Context, account common.Address, keys []common.Hash, blockNumber *big.Int) ([]byte, error) {
	results := make([]hexutil.Bytes, len(keys))
	reqs := make([]rpc.BatchElem, len(keys))

	for i := range reqs {
		reqs[i] = rpc.BatchElem{
			Method: "eth_getStorageAt",
			Args:   []interface{}{account, keys[i], toBlockNumArg(blockNumber)},
			Result: &results[i],
		}
	}
	if err := ec.c.BatchCallContext(ctx, reqs); err != nil {
		return nil, err
	}

	output := make([]byte, common.HashLength*len(keys))
	for i := range reqs {
		if reqs[i].Error != nil {
			return nil, reqs[i].Error
		}
		copy(output[i*common.HashLength:], results[i])
	}

	return output, nil
}
