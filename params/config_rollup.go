package params

import "math/big"

var AllRollupDevChainProtocolChanges = &ChainConfig{
	ChainID:                       big.NewInt(1337),
	HomesteadBlock:                big.NewInt(0),
	EIP150Block:                   big.NewInt(0),
	EIP155Block:                   big.NewInt(0),
	EIP158Block:                   big.NewInt(0),
	ByzantiumBlock:                big.NewInt(0),
	ConstantinopleBlock:           big.NewInt(0),
	PetersburgBlock:               big.NewInt(0),
	IstanbulBlock:                 big.NewInt(0),
	MuirGlacierBlock:              big.NewInt(0),
	BerlinBlock:                   big.NewInt(0),
	LondonBlock:                   big.NewInt(0),
	ArrowGlacierBlock:             big.NewInt(0),
	GrayGlacierBlock:              big.NewInt(0),
	ShanghaiTime:                  newUint64(0),
	CancunTime:                    newUint64(0),
	TerminalTotalDifficulty:       big.NewInt(0),
	TerminalTotalDifficultyPassed: true,

	//[rollup-geth] EIP-7706
	EIP7706Time: newUint64(0),
}

// IsEIP4762 returns whether eip 4762 has been activated at given block.
func (c *ChainConfig) IsEIP7706(num *big.Int, time uint64) bool {
	return isTimestampForked(c.EIP7706Time, time)
}
