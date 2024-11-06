package types

import (
	"github.com/holiman/uint256"
	"math/big"
)

type (
	VectorFeeUint   [3]*uint256.Int
	VectorFeeBigint [3]*big.Int
	VectorGasLimit  [3]uint64
)

const (
	ExecutionGasIndex = iota
	BlobGasIndex
	CalldataGasIndex
)

func NewVectorFeeBigInt() VectorFeeBigint {
	var result VectorFeeBigint
	for i := range result {
		result[i] = new(big.Int)
	}

	return result
}

func (vec VectorFeeBigint) VectorCopy() VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range result {
		if v != nil {
			result[i] = new(big.Int).Set(v)
		}
	}

	return result
}

func (vec VectorFeeBigint) Sum() *big.Int {
	sum := big.NewInt(0)
	for _, v := range vec {
		if v != nil {
			sum = sum.Add(sum, v)
		}
	}

	return sum
}

func (vec VectorFeeBigint) VectorAllEq(other VectorFeeBigint) bool {
	for i, v := range vec {
		if bothValuesNil := v == nil && other[i] == nil; bothValuesNil {
			continue
		}

		if onlyOneOfTheValuesNil := v == nil || other[i] == nil; onlyOneOfTheValuesNil {
			return false
		}

		if v.Cmp(other[i]) != 0 {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VectorAllLessOrEqual(other VectorFeeBigint) bool {
	for i, v := range vec {
		if bothValuesNil := v == nil && other[i] == nil; bothValuesNil {
			continue
		}

		if onlyOneOfTheValuesNil := v == nil || other[i] == nil; onlyOneOfTheValuesNil {
			return false
		}

		if v.Cmp(other[i]) > 0 {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VectorAdd(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		if bothValuesNil := v == nil && other[i] == nil; bothValuesNil {
			continue
		}

		if v == nil {
			result[i] = new(big.Int).Set(other[i])
			continue
		}

		if other[i] == nil {
			result[i] = new(big.Int).Set(v)
			continue
		}

		result[i] = new(big.Int).Add(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorMul(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		if anyValueNil := v == nil || other[i] == nil; anyValueNil {
			continue
		}

		result[i] = new(big.Int).Mul(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorSubtract(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		if bothValuesNil := v == nil && other[i] == nil; bothValuesNil {
			continue
		}

		if v == nil {
			result[i] = new(big.Int).Sub(big.NewInt(0), other[i])
			continue
		}

		if other[i] == nil {
			result[i] = new(big.Int).Set(v)
		}

		result[i] = new(big.Int).Sub(v, other[i])
	}

	return result
}

func (vec VectorFeeBigint) VectorSubtractClampAtZero(other VectorFeeBigint) VectorFeeBigint {
	var result VectorFeeBigint
	for i, v := range vec {
		if anyValueNil := v == nil || other[i] == nil; anyValueNil {
			result[i] = big.NewInt(0)
			continue
		}

		if subWontProducePositiveValue := v.Cmp(other[i]) <= 0; subWontProducePositiveValue {
			result[i] = big.NewInt(0)
		} else {
			result[i] = new(big.Int).Sub(v, other[i])
		}
	}

	return result
}

func (vec VectorFeeBigint) VectorAllNil() bool {
	for _, v := range vec {
		if v != nil {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VecBitLenAllZero() bool {
	for _, v := range vec {
		if v == nil {
			continue
		}

		if v.BitLen() > 0 {
			return false
		}
	}

	return true
}

func (vec VectorFeeBigint) VecBitLenAllLessEqThan256() bool {
	for _, v := range vec {
		if v == nil {
			continue
		}

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

func (vec VectorGasLimit) VectorAllEq(other VectorGasLimit) bool {
	for i, v := range vec {
		if v != other[i] {
			return false
		}
	}

	return true
}

func (vec VectorGasLimit) VectorAdd(other VectorGasLimit) VectorGasLimit {
	var result VectorGasLimit
	for i, v := range vec {
		result[i] = v + other[i]
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

func (vec VectorGasLimit) VectorSubtractClampAtZero(other VectorGasLimit) VectorGasLimit {
	var result VectorGasLimit
	for i, v := range vec {
		if v <= other[i] {
			result[i] = 0
		} else {
			result[i] = v - other[i]
		}
	}

	return result
}
