package params

const (
	//EIP-7706
	TxMinGasPrice                = 1 // Minimum gas price
	CalldataTokensPerNonZeroByte = 4
	CalldataGasPerToken          = 4
	CallDataGasLimitRatio        = 4
	BaseFeeUpdateFraction        = 8
)

// EIP-7706
var LimitTargetRatios = [3]uint64{2, 2, 4}
