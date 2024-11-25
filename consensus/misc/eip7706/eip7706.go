package eip7706

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

var (
	minTxGasPrice          = big.NewInt(params.TxMinGasPrice)
	gaspriceUpdateFraction = big.NewInt(params.BaseFeeUpdateFraction)
)

// VerifyEIP7706Header Verifies EIP-7706 header
func VerifyEIP7706Header(parent, header *types.Header) error {
	// Verify the header is not malformed
	if fieldsNilErr := MakeSureEIP7706FieldsAreNonNil(header); fieldsNilErr != nil {
		return fieldsNilErr
	}

	// Verify blob gas usage
	blobGasUsed := header.GasUsedVector[types.BlobGasIndex]
	if blobGasUsed > params.MaxBlobGasPerBlock {
		return fmt.Errorf("blob gas used %d exceeds maximum allowance %d", blobGasUsed, params.MaxBlobGasPerBlock)
	}

	if blobGasUsed%params.BlobTxBlobGasPerBlob != 0 {
		return fmt.Errorf("blob gas used %d not a multiple of blob gas per blob %d", blobGasUsed, params.BlobTxBlobGasPerBlob)
	}

	// Verify calldata gas usage
	calldataGasUsed := header.GasUsedVector[types.CalldataGasIndex]
	if calldataGasUsed%params.CalldataGasPerToken != 0 {
		return fmt.Errorf("calldata gas used %d not a multiple of calldata gas per token %d", calldataGasUsed, params.CalldataGasPerToken)
	}

	parentGasUsed, parentExcessGas, parentGasLimits := SanitizeEIP7706Fields(parent)

	// Verify the excessGas is correct based on the parent header
	expectedExcessGas := CalcExecGas(parentGasUsed, parentExcessGas, parentGasLimits)
	if !header.ExcessGas.VectorAllEq(expectedExcessGas) {
		return errors.New("invalid excessGas")
	}

	// Verify base fees are correct based on the parent header
	// NOTE: EIP-7706 doesn't specify base fees as part of header (they are not part of header protocol spec)
	// But for convenience we do calculate them and store as part of the header
	// So, if they do exist, make sure they are properly calculated
	if header.BaseFees != nil {
		expectedBaseFees := CalcBaseFees(parentExcessGas, parentGasLimits)
		if !header.BaseFees.VectorAllEq(expectedBaseFees) {
			return errors.New("invalid baseFee")
		}
	}

	return nil
}

// MakeSureEIP7706FieldsAreNonNil makes sure EIP-7706 header fields are non-nil
func MakeSureEIP7706FieldsAreNonNil(header *types.Header) error {
	if header == nil {
		return errors.New("header is nil")
	}
	if header.ExcessGas == nil {
		return errors.New("header is missing excessGas")
	}
	if header.GasUsedVector == nil {
		return errors.New("header is missing gasUsedVector")
	}
	if header.GasLimits == nil {
		return errors.New("header is missing gasLimits")
	}

	return nil
}

// SanitizeEIP7706Fields either returns the EIP-7706 existing field values or fallbacks to defaults
func SanitizeEIP7706Fields(header *types.Header) (gasUsed, excessGas, gasLimits types.VectorGasLimit) {
	if noEIP7706FieldsInHeaderErr := MakeSureEIP7706FieldsAreNonNil(header); noEIP7706FieldsInHeaderErr != nil {
		//TODO: are these defaults ok?
		gasUsed = types.VectorGasLimit{header.GasUsed, *header.BlobGasUsed, header.GasUsed / params.CallDataGasLimitRatio}
		excessGas = types.VectorGasLimit{0, *header.ExcessBlobGas, 0}
		gasLimits = types.VectorGasLimit{header.GasLimit, params.MaxBlobGasPerBlock, header.GasLimit / params.CallDataGasLimitRatio}

		return gasUsed, excessGas, gasLimits
	}

	return header.GasUsedVector, header.ExcessGas, header.GasLimits
}

// CalcBaseFeesFromParentHeader calculates base fees per EIP-7706 given parent header
func CalcBaseFeesFromParentHeader(config *params.ChainConfig, parent *types.Header) (types.VectorFeeBigint, error) {
	if parent == nil {
		return nil, errors.New("parent header is nil")
	}

	_, excessGas, gasLimits := SanitizeEIP7706Fields(parent)
	return CalcBaseFees(excessGas, gasLimits), nil
}

// CalcBaseFees  calculates vector of the base fees for current block header given parent excess gas and targets
func CalcBaseFees(parentExecGas, parentGasLimits types.VectorGasLimit) types.VectorFeeBigint {
	baseFees := make(types.VectorFeeBigint, types.VectorFeeTypesCount)
	targets := getBlockTargets(parentGasLimits)

	for i, execGas := range parentExecGas {
		target := big.NewInt(int64(targets[i]))
		target = target.Mul(target, gaspriceUpdateFraction)

		if target.Sign() == 0 {
			baseFees[i] = new(big.Int).Set(minTxGasPrice)
			continue
		}

		baseFees[i] = fakeExponential(minTxGasPrice, big.NewInt(int64(execGas)), target)
	}

	return baseFees
}

// CalcExecGas calculates excess gas given parent block gas usage and target parent target gas usage
func CalcExecGas(parentGasUsed, parentExecGas, parentGasLimits types.VectorGasLimit) types.VectorGasLimit {
	excessGas := parentExecGas.VectorAdd(parentGasUsed)

	return excessGas.VectorSubtractClampAtZero(getBlockTargets(parentGasLimits))
}

func getBlockTargets(parentGasLimits types.VectorGasLimit) types.VectorGasLimit {
	targets := make(types.VectorGasLimit, types.VectorFeeTypesCount)
	for i, limit := range parentGasLimits {
		if params.LimitTargetRatios[i] == 0 {
			targets[i] = 0
			continue
		}
		targets[i] = limit / params.LimitTargetRatios[i]
	}

	return targets
}

// fakeExponential approximates factor * e ** (numerator / denominator) using
// Taylor expansion.
func fakeExponential(factor, numerator, denominator *big.Int) *big.Int {
	var (
		output = new(big.Int)
		accum  = new(big.Int).Mul(factor, denominator)
	)

	for i := 1; accum.Sign() > 0; i++ {
		output.Add(output, accum)

		accum.Mul(accum, numerator)
		accum.Div(accum, denominator)
		accum.Div(accum, big.NewInt(int64(i)))
	}
	return output.Div(output, denominator)
}
