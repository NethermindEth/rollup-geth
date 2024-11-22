// Package types provides vector operations for new EIP-7706 TX type which introduces multi-dimensional fee structure
package types

import (
	"errors"
	"math/big"
	"slices"

	"github.com/holiman/uint256"
)

type (
	VectorFeeUint   []*uint256.Int
	VectorFeeBigint []*big.Int
	VectorGasLimit  []uint64
)

const (
	// ExecutionGasIndex represents the index for execution gas in fee vectors
	ExecutionGasIndex = iota

	// BlobGasIndex represents the index for blob gas in fee vectors
	BlobGasIndex

	// CalldataGasIndex represents the index for calldata gas in fee vectors
	CalldataGasIndex

	// VectorFeeTypesCount defines the total number of fee types supported
	VectorFeeTypesCount = 3
)

// ElementNilError is returned when a vector contains nil element(s)
var ElementNilError = errors.New("Vector contains nil element(s)")

// VectorCopy creates a deep copy of the VectorFeeBigint, nils are copied as nils
func (vec VectorFeeBigint) VectorCopy() VectorFeeBigint {
	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		if v == nil {
			result[i] = nil
		} else {
			result[i] = new(big.Int).Set(v)
		}
	}

	return result
}

// Sum calculates the sum of all elements in the vector
// Returns error if vector contains nil element(s)
func (vec VectorFeeBigint) Sum() (*big.Int, error) {
	if vec.ContainsNilElement() {
		return nil, ElementNilError
	}

	sum := big.NewInt(0)
	for _, v := range vec {
		sum = sum.Add(sum, v)
	}

	return sum, nil
}

// VectorAllEq compares two VectorFeeBigint for equality.
// Returns true if all corresponding elements are either both nil or equal in value.
func (vec VectorFeeBigint) VectorAllEq(vecOther VectorFeeBigint) bool {
	return slices.EqualFunc(vec, vecOther, func(val, other *big.Int) bool {
		if bothValuesNil := val == nil && other == nil; bothValuesNil {
			return true
		}

		if onlyOneOfTheValuesNil := val == nil || other == nil; onlyOneOfTheValuesNil {
			return false
		}

		return val.Cmp(other) == 0
	})
}

// VectorAllLessOrEqual checks if all elements in vec are less than or equal to
// corresponding elements in other vector.
// If any of the provider values contains nil element(s) returns false
func (vec VectorFeeBigint) VectorAllLessOrEqual(other VectorFeeBigint) bool {
	if vec.ContainsNilElement() || other.ContainsNilElement() {
		return false
	}

	for i, v := range vec {
		if v.Cmp(other[i]) > 0 {
			return false
		}
	}

	return true
}

// VectorAdd adds corresponding elements of two vectors.
// Returns error if any of the provided vectors contains nil element(s)
func (vec VectorFeeBigint) VectorAdd(other VectorFeeBigint) (VectorFeeBigint, error) {
	if vec.ContainsNilElement() || other.ContainsNilElement() {
		return nil, ElementNilError
	}

	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = new(big.Int).Add(v, other[i])
	}

	return result, nil
}

// VectorMul multiplies corresponding elements of two vectors.
// Returns error if any of the provided vectors contains nil element(s)
func (vec VectorFeeBigint) VectorMul(other VectorFeeBigint) (VectorFeeBigint, error) {
	if vec.ContainsNilElement() || other.ContainsNilElement() {
		return nil, ElementNilError
	}

	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = new(big.Int).Mul(v, other[i])
	}

	return result, nil
}

// VectorSubtract subtracts corresponding elements of other from vec.
// Returns error if any of the provided vectors contains nil element(s)
func (vec VectorFeeBigint) VectorSubtract(other VectorFeeBigint) (VectorFeeBigint, error) {
	if vec.ContainsNilElement() || other.ContainsNilElement() {
		return nil, ElementNilError
	}

	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = new(big.Int).Sub(v, other[i])
	}

	return result, nil
}

// VectorSubtractClampAtZero subtracts corresponding elements of other from vec,
// clamping results at zero if they would be negative.
// Returns error if any of the provided vectors contains nil element(s)
func (vec VectorFeeBigint) VectorSubtractClampAtZero(other VectorFeeBigint) (VectorFeeBigint, error) {
	if vec.ContainsNilElement() || other.ContainsNilElement() {
		return nil, ElementNilError
	}

	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		if subWontProducePositiveValue := v.Cmp(other[i]) <= 0; subWontProducePositiveValue {
			result[i] = big.NewInt(0)
		} else {
			result[i] = new(big.Int).Sub(v, other[i])
		}
	}

	return result, nil
}

// VecNoNilElements checks if vector contains nil element
func (vec VectorFeeBigint) ContainsNilElement() bool {
	return slices.Contains(vec, nil)
}

// VectorAllNil checks if all elements in the vector are nil.
func (vec VectorFeeBigint) VectorAllNil() bool {
	for _, v := range vec {
		if v != nil {
			return false
		}
	}
	return true
}

// VecBitLenAllZero checks if all non-nil elements have a bit length of zero.
// Nil is treated as having a bit length of zero.
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

// VecBitLenAllLessEqThan256 checks if all non-nil elements have a bit length <= 256.
// Nil is treated as having a bit length of zero.
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

func (vec VectorFeeBigint) ToVectorUint() VectorFeeUint {
	result := make(VectorFeeUint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i], _ = uint256.FromBig(v)
	}
	return result
}

// ToVectorBigInt converts a VectorGasLimit to VectorFeeBigint.
func (vec VectorGasLimit) ToVectorBigInt() VectorFeeBigint {
	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = new(big.Int).SetUint64(v)
	}
	return result
}

// VectorAllEq compares two VectorGasLimit for equality.
// Returns true if all corresponding elements are equal.
func (vec VectorGasLimit) VectorAllEq(other VectorGasLimit) bool {
	for i, v := range vec {
		if v != other[i] {
			return false
		}
	}
	return true
}

// VectorAdd adds corresponding elements of two VectorGasLimit.
// Note: Does not check for overflow.
func (vec VectorGasLimit) VectorAdd(other VectorGasLimit) VectorGasLimit {
	result := make(VectorGasLimit, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = v + other[i]
	}
	return result
}

// VectorSubtractClampAtZero subtracts corresponding elements of other from vec,
// clamping results at zero if they would be negative.
func (vec VectorGasLimit) VectorSubtractClampAtZero(other VectorGasLimit) VectorGasLimit {
	result := make(VectorGasLimit, VectorFeeTypesCount)
	for i, v := range vec {
		if v <= other[i] {
			result[i] = 0
		} else {
			result[i] = v - other[i]
		}
	}
	return result
}

// ToVectorBigInt converts a VectorFeeUint to VectorFeeBigint.
func (vec VectorFeeUint) ToVectorBigInt() VectorFeeBigint {
	result := make(VectorFeeBigint, VectorFeeTypesCount)
	for i, v := range vec {
		result[i] = v.ToBig()
	}
	return result
}
