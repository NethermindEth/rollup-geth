package node

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

func (n *Node) RegisterEthClient(endpoint string) {
	ethClient, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Error("Unable to connect to ETH RPC endpoint at", "URL", ethClient, "error", err)
		return
	}

	n.ethClient = ethClient
	log.Info("Initialized ETH RPC client", "endpoint", ethClient)
}
