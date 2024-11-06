package eip7706

import (
	"math/big"
	"regexp"
	"slices"
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

		if !slices.Equal(tt.want, got) {
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

		if !slices.Equal(tt.want, got) {
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
			parentExecGas:   types.VectorGasLimit{0, 0, 0},
			parentGasLimits: types.VectorGasLimit{0, 0, 0},
			expectedBaseFees: types.VectorFeeBigint{
				big.NewInt(1),
				big.NewInt(1),
				big.NewInt(1),
			},
			description: "All zero",
		},
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

// TestVerifyEIP7706Header tests the VerifyEIP7706Header function.
func TestVerifyEIP7706Header(t *testing.T) {
	// Helper function to create a default header.
	defaultHeader := func() *types.Header {
		header := &types.Header{
			GasLimits:     types.VectorGasLimit{params.GenesisGasLimit, params.MaxBlobGasPerBlock, params.GenesisGasLimit / params.CallDataGasLimitRatio},
			GasUsedVector: types.VectorGasLimit{0, 0, 0},
			ExcessGas:     types.VectorGasLimit{0, 0, 0},
			// BaseFees are optional and can be set for testing.
			BaseFees: nil,
		}
		return header
	}

	// Helper function to create a default parent header.
	defaultParent := func() *types.Header {
		blobGasUsed := uint64(0)
		excessBlobGas := uint64(0)
		parent := &types.Header{
			GasLimit:      params.GenesisGasLimit,
			GasUsed:       0,
			BlobGasUsed:   &blobGasUsed,
			ExcessBlobGas: &excessBlobGas,
			GasLimits:     nil, // Pre-fork parent; GasLimits are nil.
			GasUsedVector: nil,
			ExcessGas:     nil,
			BaseFees:      nil,
			Number:        big.NewInt(0),
		}
		return parent
	}

	// Helper function to create a parent header that is an EIP-7706 block.
	eip7706Parent := func() *types.Header {
		parent := &types.Header{
			GasLimits:     types.VectorGasLimit{params.GenesisGasLimit, params.MaxBlobGasPerBlock, params.GenesisGasLimit / params.CallDataGasLimitRatio},
			GasUsedVector: types.VectorGasLimit{0, 0, 0},
			ExcessGas:     types.VectorGasLimit{0, 0, 0},
			BaseFees:      nil,
			Number:        big.NewInt(1),
		}
		return parent
	}

	// Define test cases.
	tests := []struct {
		name         string
		parent       *types.Header
		header       *types.Header
		expectError  bool
		errorMessage string
	}{
		{
			name:        "Valid header with EIP-7706 parent",
			parent:      eip7706Parent(),
			header:      defaultHeader(),
			expectError: false,
		},
		{
			name:        "Valid header with pre-fork parent",
			parent:      defaultParent(),
			header:      defaultHeader(),
			expectError: false,
		},
		{
			name:   "Header missing ExcessGas",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.ExcessGas = nil
				return h
			}(),
			expectError:  true,
			errorMessage: "header is missing excessGas",
		},
		{
			name:   "Header missing GasUsedVector",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.GasUsedVector = nil
				return h
			}(),
			expectError:  true,
			errorMessage: "header is missing gasUsedVector",
		},
		{
			name:   "Header missing GasLimits",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.GasLimits = nil
				return h
			}(),
			expectError:  true,
			errorMessage: "header is missing gasLimits",
		},
		{
			name:   "Blob gas used exceeds MaxBlobGasPerBlock",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.GasUsedVector[1] = params.MaxBlobGasPerBlock + params.BlobTxBlobGasPerBlob
				return h
			}(),
			expectError:  true,
			errorMessage: "blob gas used \\d+ exceeds maximum allowance \\d+",
		},
		{
			name:   "Blob gas used not multiple of BlobTxBlobGasPerBlob",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.GasUsedVector[1] = params.BlobTxBlobGasPerBlob + 1
				return h
			}(),
			expectError:  true,
			errorMessage: "blob gas used \\d+ not a multiple of blob gas per blob \\d+",
		},
		{
			name:   "Calldata gas used not multiple of CalldataGasPerToken",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.GasUsedVector[2] = params.CalldataGasPerToken + 1
				return h
			}(),
			expectError:  true,
			errorMessage: "calldata gas used \\d+ not a multiple of calldata gas per token \\d+",
		},
		{
			name:   "Invalid excessGas",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				h.ExcessGas[0] = 1 // Set an invalid excessGas value.
				return h
			}(),
			expectError:  true,
			errorMessage: "invalid excessGas",
		},
		{
			name:   "Invalid baseFee",
			parent: eip7706Parent(),
			header: func() *types.Header {
				h := defaultHeader()
				// Set incorrect BaseFees for testing.
				h.BaseFees = types.VectorFeeBigint{big.NewInt(0), big.NewInt(1), big.NewInt(1)}
				return h
			}(),
			expectError:  true,
			errorMessage: "invalid baseFee",
		},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyEIP7706Header(tt.parent, tt.header)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMessage != "" {
					if matched := matchErrorMessage(err, tt.errorMessage); !matched {
						t.Errorf("unexpected error message: got %v, expected %v", err.Error(), tt.errorMessage)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to match error messages using regex.
func matchErrorMessage(err error, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, err.Error())
	return matched
}
