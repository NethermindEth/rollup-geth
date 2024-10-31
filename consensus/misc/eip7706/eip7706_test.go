package eip7706

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func TestGetBlockTargets(t *testing.T) {
	tests := []struct {
		parentGasLimits   types.VectorGasLimit
		limitTargetRatios [3]uint64
		want              types.VectorGasLimit
	}{
		// Basic case
		{
			parentGasLimits:   types.VectorGasLimit{10000000, 20000000, 40000000},
			limitTargetRatios: [3]uint64{2, 2, 4},
			want:              types.VectorGasLimit{5000000, 10000000, 10000000},
		},
		// Edge case: Zero gas limits
		{
			parentGasLimits:   types.VectorGasLimit{0, 0, 0},
			limitTargetRatios: [3]uint64{2, 2, 4},
			want:              types.VectorGasLimit{0, 0, 0},
		},
		// Edge case: Zero limit target ratios
		{
			parentGasLimits:   types.VectorGasLimit{10000000, 20000000, 40000000},
			limitTargetRatios: [3]uint64{0, 0, 0}, // Should handle division by zero
			want:              types.VectorGasLimit{0, 0, 0},
		},
		// Edge case: limit target ratios and parent gas limits don't have same element count
		{
			parentGasLimits:   types.VectorGasLimit{10000000, 20000000, 40000000},
			limitTargetRatios: [3]uint64{1, 2},
			want:              types.VectorGasLimit{10000000, 10000000, 0},
		},
	}

	defaultParams := params.LimitTargetRatios
	for i, tt := range tests {
		params.LimitTargetRatios = tt.limitTargetRatios // Set the ratios for the test
		got := getBlockTargets(tt.parentGasLimits)

		if got != tt.want {
			t.Errorf("test %d: getBlockTargets(%v) = %v; want %v",
				i, tt.parentGasLimits, got, tt.want)
		}
	}

	params.LimitTargetRatios = defaultParams
}

func TestCalcExecGas(t *testing.T) {
	tests := []struct {
		parentGasUsed   types.VectorGasLimit
		parentExecGas   types.VectorGasLimit
		parentGasLimits types.VectorGasLimit
		want            types.VectorGasLimit
	}{
		// Basic case
		{
			parentGasUsed:   types.VectorGasLimit{10000000, 15000000, 20000000},
			parentExecGas:   types.VectorGasLimit{5000000, 10000000, 15000000},
			parentGasLimits: types.VectorGasLimit{20000000, 40000000, 80000000},
			want:            types.VectorGasLimit{5000000, 5000000, 15000000},
		},

		// Basic case clamp at zero
		{
			parentGasUsed:   types.VectorGasLimit{10000000, 15000000, 2000},
			parentExecGas:   types.VectorGasLimit{5000000, 10000000, 1500},
			parentGasLimits: types.VectorGasLimit{20000000, 40000000, 80000000},
			want:            types.VectorGasLimit{5000000, 5000000, 0},
		},
		// Edge case: Zero gas used and exec gas
		{
			parentGasUsed:   types.VectorGasLimit{0, 0, 0},
			parentExecGas:   types.VectorGasLimit{0, 0, 0},
			parentGasLimits: types.VectorGasLimit{20000000, 40000000, 80000000},
			want:            types.VectorGasLimit{0, 0, 0},
		},
	}

	for i, tt := range tests {
		got := CalcExecGas(tt.parentGasUsed, tt.parentExecGas, tt.parentGasLimits)

		if got != tt.want {
			t.Errorf("test %d: CalcExecGas(%v, %v, %v) = %v; want %v",
				i, tt.parentGasUsed, tt.parentExecGas, tt.parentGasLimits, got, tt.want)
		}
	}
}

// func TestCalcBaseFees(t *te
func TestCalcBaseFees(t *testing.T) {
	tests := []struct {
		parentExecGas    types.VectorGasLimit
		parentGasLimits  types.VectorGasLimit
		expectedBaseFees types.VectorFeeBigint
		description      string
	}{
		{
			parentExecGas:   types.VectorGasLimit{10_000_000, 10_000_000, 10_000_000},
			parentGasLimits: types.VectorGasLimit{20_000_000, 20_000_000, 40_000_000},
			expectedBaseFees: types.VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(1),
				big.NewInt(1),
			},
			description: "Usage equals target",
		},
		{
			parentExecGas:   types.VectorGasLimit{9_000_000, 9_000_000, 9_000_000},
			parentGasLimits: types.VectorGasLimit{20_000_000, 20_000_000, 40_000_000},
			expectedBaseFees: types.VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(1),
				big.NewInt(1),
			},
			description: "Usage below target",
		},
		{
			parentExecGas:   types.VectorGasLimit{60_000_000, 80_000_000, 80_000_000},
			parentGasLimits: types.VectorGasLimit{20_000_000, 20_000_000, 40_000_000},
			expectedBaseFees: types.VectorFeeBigint{
				big.NewInt(2),
				big.NewInt(2),
				big.NewInt(2),
			},
			description: "Usage above target",
		},
	}

	gasType := [3]string{"execution", "blob", "calldata"}
	for i, test := range tests {
		have := CalcBaseFees(test.parentExecGas, test.parentGasLimits)
		for j := 0; j < 3; j++ {
			if have[j].Cmp(test.expectedBaseFees[j]) != 0 {
				t.Errorf("test %d (%s), gas type %s: have %d, want %d",
					i, test.description, gasType[j], have[j], test.expectedBaseFees[j])
			}
		}
	}
}
