package params

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// [RIP-7728](https://github.com/ethereum/RIPs/blob/d75e5bb4cd4a3a642090ba15249c11bfccb064db/RIPS/rip-7728.md)
const (
	L1SLoadBaseGas            uint64 = 2000 // Base price for L1Sload
	L1SLoadPerLoadGas         uint64 = 2000 // Per-load price for loading one storage slot
	L1SLoadMaxNumStorageSlots        = 5    // Max number of storage slots requested in L1Sload precompile
	L1SLoadRPCTimeoutInSec           = 3    // After how many seconds RPC timeouts
)

// [EIP-7708](https://eips.ethereum.org/EIPS/eip-7708) specific params
var (
	// LogNativeTransferTopicMagic proposed here: https://ethereum-magicians.org/t/eip-7708-eth-transfers-emit-a-log/20034/3?u=mralj
	LogNativeTransferTopicMagic = common.BytesToHash(crypto.Keccak256([]byte("0000000000000000000000000000000000000000000000000000000000000000")))
	// LogNativeTransferContractAddress Proposed here: https://ethereum-magicians.org/t/eip-7708-eth-transfers-emit-a-log/20034/15
	LogNativeTransferContractAddress = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
)
