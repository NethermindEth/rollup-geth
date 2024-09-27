package params

const (
	L1SLoadBaseGas            uint64 = 2000 // Base price for L1Sload
	L1SLoadPerLoadGas         uint64 = 2000 // Per-load price for loading one storage slot
	L1SLoadMaxNumStorageSlots        = 5    // Max number of storage slots requested in L1Sload precompile
)

var L1SLoadRPCTimeoutInSec = MainnetChainConfig.Clique.Period // After how many ms will RPC call timeout
