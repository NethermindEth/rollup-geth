package types

import (
	"math/big"

	"github.com/holiman/uint256"
)

type (
	VectorFeeUint   [3]*uint256.Int
	VectorFeeBigint [3]*big.Int
	VectorGasLimit  [3]uint64
)

func NewVectorFeeBigInt() VectorFeeBigint {
	var result VectorFeeBigint
	for i := range result {
		result[i] = new(big.Int)
	}

	return result
}

// TODO: Add nil checks
func (vec VectorFeeBigint) Sum() *big.Int {
	sum := big.NewInt(0)
	for _, v := range vec {
		sum = sum.Add(sum, v)
	}

	return sum
}

func (vec VectorFeeBigint) VectorAllLessOrEqual(other VectorFeeBigint) bool {
	for i, v := range vec {
		if v.Cmp(other[i]) > 0 {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VectorAdd(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		result[i] = new(big.Int).Add(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorMul(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		result[i] = new(big.Int).Mul(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorSubtract(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		result[i] = new(big.Int).Sub(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorSubtractClampAtZero(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		if subWontProducePositiveValue := v.Cmp(other[i]) <= 0; subWontProducePositiveValue {
			result[i] = big.NewInt(0)
		} else {
			result[i] = new(big.Int).Sub(v, other[i])
		}
	}

	return result
}

func (vec VectorFeeBigint) VecBitLenAllZero() bool {
	for _, v := range vec {
		if v.BitLen() > 0 {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VecBitLenAllLessEqThan256() bool {
	for _, v := range vec {
		if v.BitLen() > 256 {
			return false
		}
	}

	return true
}

func (vec VectorGasLimit) ToVectorBigInt() VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		result[i] = new(big.Int).SetUint64(v)
	}

	return result
}

func (vec VectorGasLimit) VectorSubtract(other VectorGasLimit) VectorGasLimit {
	var result VectorGasLimit
	for i, v := range vec {
		result[i] = v - other[i]
	}

	return result
}
