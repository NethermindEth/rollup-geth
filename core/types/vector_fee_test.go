package types

import (
	"math/big"
	"testing"
)

func TestVectorFeeBigint_ContainsNilElement(t *testing.T) {
	tests := []struct {
		name string
		vec  VectorFeeBigint
		want bool
	}{
		{
			name: "no nil elements",
			vec: VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
			want: false,
		},
		{
			name: "one nil element",
			vec: VectorFeeBigint{
				big.NewInt(1),
				nil,
				big.NewInt(3),
			},
			want: true,
		},
		{
			name: "all nil elements",
			vec: VectorFeeBigint{
				nil,
				nil,
				nil,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec.ContainsNilElement(); got != tt.want {
				t.Errorf("ContainsNilElement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVectorFeeBigint_VectorAllNil(t *testing.T) {
	tests := []struct {
		name string
		vec  VectorFeeBigint
		want bool
	}{
		{
			name: "no nil elements",
			vec: VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
			want: false,
		},
		{
			name: "some nil elements",
			vec: VectorFeeBigint{
				big.NewInt(1),
				nil,
				big.NewInt(3),
			},
			want: false,
		},
		{
			name: "all nil elements",
			vec: VectorFeeBigint{
				nil,
				nil,
				nil,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec.VectorAllNil(); got != tt.want {
				t.Errorf("VectorAllNil() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestVectorFeeBigint_VectorCopy(t *testing.T) {
	tests := []struct {
		name string
		vec  VectorFeeBigint
	}{
		{
			name: "all non-nil elements",
			vec: VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
		},
		{
			name: "some nil elements",
			vec: VectorFeeBigint{
				big.NewInt(1),
				nil,
				big.NewInt(3),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copy := tt.vec.VectorCopy()
			if !tt.vec.VectorAllEq(copy) {
				t.Error("Copy not equal to original")
			}

			// Modify copy to ensure deep copy
			for _, v := range copy {
				if v != nil {
					v.Add(v, big.NewInt(1))
				}
			}
			if tt.vec.VectorAllEq(copy) {
				t.Error("Modified copy should not be equal to original")
			}
		})
	}
}

func TestVectorFeeBigint_Sum(t *testing.T) {
	tests := []struct {
		name    string
		vec     VectorFeeBigint
		want    *big.Int
		wantErr bool
	}{
		{
			name: "valid sum",
			vec: VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(2),
				big.NewInt(3),
			},
			want:    big.NewInt(6),
			wantErr: false,
		},
		{
			name: "contains nil",
			vec: VectorFeeBigint{
				big.NewInt(1),
				nil,
				big.NewInt(3),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.vec.Sum()
			if (err != nil) != tt.wantErr {
				t.Errorf("Sum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("Sum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVectorFeeBigint_VectorAllEq(t *testing.T) {
	tests := []struct {
		name   string
		vec1   VectorFeeBigint
		vec2   VectorFeeBigint
		result bool
	}{
		{
			name:   "equal vectors",
			vec1:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			result: true,
		},
		{
			name:   "unequal vectors",
			vec1:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(4)},
			result: false,
		},
		{
			name:   "both nil elements",
			vec1:   VectorFeeBigint{big.NewInt(1), nil, big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(1), nil, big.NewInt(3)},
			result: true,
		},
		{
			name:   "one nil element",
			vec1:   VectorFeeBigint{big.NewInt(1), nil, big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			result: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec1.VectorAllEq(tt.vec2); got != tt.result {
				t.Errorf("VectorAllEq() = %v, want %v", got, tt.result)
			}
		})
	}
}

func TestVectorFeeBigint_VectorAllLessOrEqual(t *testing.T) {
	tests := []struct {
		name   string
		vec1   VectorFeeBigint
		vec2   VectorFeeBigint
		result bool
	}{
		{
			name:   "all less",
			vec1:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(2), big.NewInt(3), big.NewInt(4)},
			result: true,
		},
		{
			name:   "all equal",
			vec1:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
			result: true,
		},
		{
			name:   "one greater",
			vec1:   VectorFeeBigint{big.NewInt(1), big.NewInt(3), big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(2), big.NewInt(2), big.NewInt(4)},
			result: false,
		},
		{
			name:   "with nil elements",
			vec1:   VectorFeeBigint{big.NewInt(1), nil, big.NewInt(3)},
			vec2:   VectorFeeBigint{big.NewInt(2), big.NewInt(2), big.NewInt(4)},
			result: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec1.VectorAllLessOrEqual(tt.vec2); got != tt.result {
				t.Errorf("VectorAllLessOrEqual() = %v, want %v", got, tt.result)
			}
		})
	}
}

func TestVectorFeeBigint_VectorOperations(t *testing.T) {
	tests := []struct {
		name         string
		vec1         VectorFeeBigint
		vec2         VectorFeeBigint
		wantAdd      VectorFeeBigint
		wantMul      VectorFeeBigint
		wantSub      VectorFeeBigint
		wantSubClamp VectorFeeBigint
		wantErr      bool
	}{
		{
			name:         "valid operations",
			vec1:         VectorFeeBigint{big.NewInt(1), big.NewInt(3), big.NewInt(5)},
			vec2:         VectorFeeBigint{big.NewInt(2), big.NewInt(3), big.NewInt(4)},
			wantAdd:      VectorFeeBigint{big.NewInt(3), big.NewInt(6), big.NewInt(9)},
			wantMul:      VectorFeeBigint{big.NewInt(2), big.NewInt(9), big.NewInt(20)},
			wantSub:      VectorFeeBigint{big.NewInt(-1), big.NewInt(0), big.NewInt(1)},
			wantSubClamp: VectorFeeBigint{big.NewInt(0), big.NewInt(0), big.NewInt(1)},
			wantErr:      false,
		},
		{
			name:    "with nil elements",
			vec1:    VectorFeeBigint{big.NewInt(1), nil, big.NewInt(3)},
			vec2:    VectorFeeBigint{big.NewInt(2), big.NewInt(3), big.NewInt(4)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Add
			resultAdd, err := tt.vec1.VectorAdd(tt.vec2)
			if (err != nil) != tt.wantErr {
				t.Errorf("VectorAdd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.wantAdd.VectorAllEq(resultAdd) {
				t.Error("VectorAdd() result not equal to expected")
			}

			// Test Multiply
			resultMul, err := tt.vec1.VectorMul(tt.vec2)
			if (err != nil) != tt.wantErr {
				t.Errorf("VectorMul() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.wantMul.VectorAllEq(resultMul) {
				t.Error("VectorMul() result not equal to expected")
			}

			// Test Subtract
			resultSub, err := tt.vec1.VectorSubtract(tt.vec2)
			if (err != nil) != tt.wantErr {
				t.Errorf("VectorSubtract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.wantSub.VectorAllEq(resultSub) {
				t.Error("VectorSubtract() result not equal to expected")
			}

			// Test Subtract Clamp at zero
			resultSubClamp, err := tt.vec1.VectorSubtractClampAtZero(tt.vec2)
			if (err != nil) != tt.wantErr {
				t.Errorf("VectorSubtractClampAtZero() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.wantSubClamp.VectorAllEq(resultSubClamp) {
				t.Error("VectorSubtractClampAtZero() result not equal to expected")
			}
		})
	}
}

func TestVectorFeeBigint_BitLengthChecks(t *testing.T) {
	bigNum := new(big.Int).Lsh(big.NewInt(1), 257) // 2^257

	tests := []struct {
		name             string
		vec              VectorFeeBigint
		wantAllZero      bool
		wantAllLessEq256 bool
	}{
		{
			name: "all zero",
			vec: VectorFeeBigint{
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
			},
			wantAllZero:      true,
			wantAllLessEq256: true,
		},
		{
			name: "some non-zero, all <= 256",
			vec: VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(0),
				nil,
			},
			wantAllZero:      false,
			wantAllLessEq256: true,
		},
		{
			name: "with large number",
			vec: VectorFeeBigint{
				big.NewInt(1),
				bigNum,
				big.NewInt(0),
			},
			wantAllZero:      false,
			wantAllLessEq256: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec.VecBitLenAllZero(); got != tt.wantAllZero {
				t.Errorf("VecBitLenAllZero() = %v, want %v", got, tt.wantAllZero)
			}
			if got := tt.vec.VecBitLenAllLessEqThan256(); got != tt.wantAllLessEq256 {
				t.Errorf("VecBitLenAllLessEqThan256() = %v, want %v", got, tt.wantAllLessEq256)
			}
		})
	}
}

func TestVectorGasLimit_Operations(t *testing.T) {
	tests := []struct {
		name         string
		vec1         VectorGasLimit
		vec2         VectorGasLimit
		wantAdd      VectorGasLimit
		wantSub      VectorGasLimit
		wantSubClamp VectorGasLimit
	}{
		{
			name:         "basic operations",
			vec1:         VectorGasLimit{100, 200, 300},
			vec2:         VectorGasLimit{50, 100, 150},
			wantAdd:      VectorGasLimit{150, 300, 450},
			wantSub:      VectorGasLimit{50, 100, 150},
			wantSubClamp: VectorGasLimit{50, 100, 150},
		},
		{
			name:         "with clamping",
			vec1:         VectorGasLimit{100, 200, 300},
			vec2:         VectorGasLimit{150, 250, 250},
			wantAdd:      VectorGasLimit{250, 450, 550},
			wantSubClamp: VectorGasLimit{0, 0, 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultAdd := tt.vec1.VectorAdd(tt.vec2)
			if !resultAdd.VectorAllEq(tt.wantAdd) {
				t.Errorf("VectorAdd() = %v, want %v", resultAdd, tt.wantAdd)
			}

			resultSubClamp := tt.vec1.VectorSubtractClampAtZero(tt.vec2)
			if !resultSubClamp.VectorAllEq(tt.wantSubClamp) {
				t.Errorf("VectorSubtractClampAtZero() = %v, want %v", resultSubClamp, tt.wantSubClamp)
			}
		})
	}
}

func TestVectorGasLimit_ToVectorBigInt(t *testing.T) {
	tests := []struct {
		name string
		vec  VectorGasLimit
		want VectorFeeBigint
	}{
		{
			name: "basic conversion",
			vec:  VectorGasLimit{100, 200, 300},
			want: VectorFeeBigint{
				big.NewInt(100),
				big.NewInt(200),
				big.NewInt(300),
			},
		},
		{
			name: "zero values",
			vec:  VectorGasLimit{0, 0, 0},
			want: VectorFeeBigint{
				big.NewInt(0),
				big.NewInt(0),
				big.NewInt(0),
			},
		},
		{
			name: "max uint64",
			vec:  VectorGasLimit{^uint64(0), ^uint64(0), ^uint64(0)},
			want: VectorFeeBigint{
				new(big.Int).SetUint64(^uint64(0)),
				new(big.Int).SetUint64(^uint64(0)),
				new(big.Int).SetUint64(^uint64(0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vec.ToVectorBigInt()
			if !got.VectorAllEq(tt.want) {
				t.Errorf("ToVectorBigInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVectorGasLimit_VectorAllEq(t *testing.T) {
	tests := []struct {
		name string
		vec1 VectorGasLimit
		vec2 VectorGasLimit
		want bool
	}{
		{
			name: "equal vectors",
			vec1: VectorGasLimit{100, 200, 300},
			vec2: VectorGasLimit{100, 200, 300},
			want: true,
		},
		{
			name: "unequal vectors",
			vec1: VectorGasLimit{100, 200, 300},
			vec2: VectorGasLimit{100, 200, 301},
			want: false,
		},
		{
			name: "zero vectors",
			vec1: VectorGasLimit{0, 0, 0},
			vec2: VectorGasLimit{0, 0, 0},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vec1.VectorAllEq(tt.vec2); got != tt.want {
				t.Errorf("VectorAllEq() = %v, want %v", got, tt.want)
			}
		})
	}
}
