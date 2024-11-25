package params

import "math/big"

// IsEIP4762 returns whether eip 4762 has been activated at given block.
func (c *ChainConfig) IsEIP7706(num *big.Int, time uint64) bool {
	return isTimestampForked(c.EIP7706Time, time)
}
