package vm

import (
	"bytes"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"math/big"
	"testing"
)

func FuzzOpIsStatic(f *testing.F) {
	opcodes := []byte{
		byte(PUSH1), 0,
		byte(CALLDATALOAD),
		byte(PUSH1), 2,
		byte(MUL),
		byte(PUSH1), 0,
		byte(MSTORE),
		byte(ISSTATIC), // This is the new ISSTATIC opcode (EIP-2970)
		byte(PUSH1), 32,
		byte(MSTORE),
		byte(PUSH1), 64,
		byte(PUSH1), 0,
		byte(RETURN),
	}
	f.Add(opcodes)

	f.Fuzz(func(t *testing.T, bytecode []byte) {
		// Make sure the bytecode contains a single ISSTATIC opcode
		if bytes.Count(bytecode, []byte{byte(ISSTATIC)}) != 1 {
			return
		}
		// Skip if the bytecode contains a STATICCALL opcode
		if bytes.Contains(bytecode, []byte{byte(STATICCALL)}) {
			return
		}

		isStaticIndex := bytes.IndexByte(bytecode, byte(ISSTATIC))
		// Skip if the ISSTATIC opcode is the last opcode
		if isStaticIndex == len(bytecode)-1 {
			return
		}

		contract := NewContract(contractRef{common.Address{}}, AccountRef(common.Address{1}), uint256.NewInt(0), 100000)
		contract.Code = bytecode

		tracer := &isStaticTracer{opcodeAfterIsStatic: bytecode[isStaticIndex+1]}
		statedb, err := state.New(types.EmptyRootHash, state.NewDatabaseForTesting())
		if err != nil {
			t.Fatal(err)
		}
		var (
			env = NewEVM(
				BlockContext{BlockNumber: big.NewInt(1), Random: &common.Hash{}, Time: 1},
				TxContext{},
				statedb,
				params.MergedTestChainConfig,
				Config{
					Tracer: &tracing.Hooks{
						OnOpcode: tracer.OnOpcode,
					},
				},
			)
			evmInterpreter = env.interpreter
		)

		// `evmInterpreter.Run` might panic if the bytecode is invalid
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic due to invalid opcode %v", r)
				return
			}
		}()

		for _, tt := range []struct {
			name                  string
			isReadOnly            bool
			expectedIsStaticValue *uint256.Int
		}{
			{name: "ISSTATIC=true", isReadOnly: true, expectedIsStaticValue: uint256.NewInt(1)},
			{name: "ISSTATIC=false", isReadOnly: false, expectedIsStaticValue: uint256.NewInt(0)},
		} {
			// Reset the tracer's values
			tracer.called = false
			tracer.returnedValue = nil

			_, err = evmInterpreter.Run(contract, nil, tt.isReadOnly)
			if err != nil {
				// Skip if the execution failed due to an invalid bytecode
				return
			}

			// Even though ISSTATIC opcode is present in the bytecode, it might not be executed
			if tracer.called && !tracer.returnedValue.Eq(tt.expectedIsStaticValue) {
				t.Errorf("Testcase %v: ISSTATIC opcode was not called", tt.name)
			}
		}
	})
}
