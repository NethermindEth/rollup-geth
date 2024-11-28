package types

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// TestHeaderRLPEncoding tests basic header RLP encoding/decoding
func TestHeaderRLPEncoding(t *testing.T) {
	header := &Header{
		ParentHash:  common.HexToHash("0x1234567890"),
		UncleHash:   common.HexToHash("0x9876543210"),
		Coinbase:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Root:        common.HexToHash("0xabcdef0123"),
		TxHash:      common.HexToHash("0x4567890abc"),
		ReceiptHash: common.HexToHash("0x7890abcdef"),
		Difficulty:  big.NewInt(12345),
		Number:      big.NewInt(67890),
		GasLimit:    1000000,
		GasUsed:     500000,
		Time:        1234567890,
		Extra:       []byte("test extra data"),
		MixDigest:   common.HexToHash("0x1234567890abcdef"),
		Nonce:       [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}

	// Encode and decode
	encoded, err := rlp.EncodeToBytes(header)
	if err != nil {
		t.Fatalf("Failed to encode header: %v", err)
	}

	var decoded Header
	if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
		t.Fatalf("Failed to decode header: %v", err)
	}

	if !reflect.DeepEqual(*header, decoded) {
		t.Fatalf("Decoded header not equal to original. Expected %v, got %v", *header, decoded)
	}
}

// TestHeaderRLPPreEIP7706 tests encoding/decoding of headers with pre-EIP7706 fields
func TestHeaderRLPPreEIP7706(t *testing.T) {
	header := &Header{
		ParentHash:  common.HexToHash("0x1234567890"),
		UncleHash:   common.HexToHash("0x9876543210"),
		Coinbase:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Root:        common.HexToHash("0xabcdef0123"),
		TxHash:      common.HexToHash("0x4567890abc"),
		ReceiptHash: common.HexToHash("0x7890abcdef"),
		Difficulty:  big.NewInt(12345),
		Number:      big.NewInt(67890),
		GasLimit:    1000000,
		GasUsed:     500000,
		Time:        1234567890,
		Extra:       []byte("test extra data"),
		MixDigest:   common.HexToHash("0x1234567890abcdef"),
		Nonce:       [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	}

	testCases := []struct {
		name   string
		header *Header
	}{
		{
			name:   "WithBaseFee",
			header: (func() *Header { h := CopyHeader(header); h.BaseFee = big.NewInt(10); return h })(),
		},
		{
			name: "WithBlobGasFields",
			header: (func() *Header {
				h := CopyHeader(header)
				//NOTE: not strictly necessary, but if nil, this will be decoded as 0
				// Since I'm using reflect.DeepEqual, it will fail (nil != 0), so this is easier
				h.BaseFee = big.NewInt(100)

				//NOTE: if the h.WithdrawalsHash is nil, the test will fail
				//because the rlp decoding fails, because
				//nil WithdrawalsHash will be encoded as empty string, and common.Hash would fail to decode
				h.WithdrawalsHash = &common.Hash{1, 2, 3}

				blobGasUsed := uint64(2000)
				excessBlobGas := uint64(3000)
				h.BlobGasUsed = &blobGasUsed
				h.ExcessBlobGas = &excessBlobGas
				return h
			})(),
		},
		{
			name: "WithAllPreEIP7706Fields",
			header: (func() *Header {
				h := CopyHeader(header)

				h.BaseFee = big.NewInt(100)
				h.WithdrawalsHash = &common.Hash{1, 2, 3}

				blobGasUsed := uint64(2000)
				excessBlobGas := uint64(3000)
				h.BlobGasUsed = &blobGasUsed
				h.ExcessBlobGas = &excessBlobGas

				h.ParentBeaconRoot = &common.Hash{4, 5, 6}
				h.RequestsHash = &common.Hash{7, 8, 9}

				return h
			})(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := rlp.EncodeToBytes(tc.header)
			if err != nil {
				t.Fatalf("Failed to encode header: %v", err)
			}

			var decoded Header
			if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
				t.Fatalf("Failed to decode header: %v", err)
			}

			if !reflect.DeepEqual(*tc.header, decoded) {
				t.Fatalf("Decoded header not equal to original. \n Expected %v, \n got %v", *tc.header, decoded)
			}
		})
	}
}

// TestHeaderRLPEIP7706 tests encoding/decoding of EIP-7706 headers
func TestHeaderRLPEIP7706(t *testing.T) {
	header := &Header{
		ParentHash:  common.HexToHash("0x1234567890"),
		UncleHash:   common.HexToHash("0x9876543210"),
		Coinbase:    common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Root:        common.HexToHash("0xabcdef0123"),
		TxHash:      common.HexToHash("0x4567890abc"),
		ReceiptHash: common.HexToHash("0x7890abcdef"),
		Difficulty:  big.NewInt(12345),
		Number:      big.NewInt(67890),
		Time:        1234567890,
		Extra:       []byte("test extra data"),
		MixDigest:   common.HexToHash("0x1234567890abcdef"),
		Nonce:       [8]byte{1, 2, 3, 4, 5, 6, 7, 8},

		WithdrawalsHash:  &common.Hash{1, 2, 3},
		ParentBeaconRoot: &common.Hash{4, 5, 6},
		RequestsHash:     &common.Hash{7, 8, 9},

		GasUsedVector: VectorGasLimit{1, 2, 3},
		GasLimits:     VectorGasLimit{1, 2, 3},
		ExcessGas:     VectorGasLimit{1, 2, 3},
	}

	// Encode and decode
	encoded, err := rlp.EncodeToBytes(header)
	if err != nil {
		t.Fatalf("Failed to encode EIP-7706 header: %v", err)
	}

	var decoded Header
	if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
		t.Fatalf("Failed to decode EIP-7706 header: %v", err)
	}

	if !reflect.DeepEqual(*header, decoded) {
		t.Fatalf("Decoded header not equal to original. \n Expected %v, \n got %v", *header, decoded)
	}

	//	Verify that GasLimit and GasUsed are not encoded in EIP-7706 mode
	if decoded.GasLimit != 0 || decoded.GasUsed != 0 {
		t.Error("GasLimit or GasUsed should not be present in EIP-7706 header")
	}
}

// TestHeaderRLPEdgeCases tests various edge cases and error conditions
func TestHeaderRLPEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		header      *Header
		expectError bool
	}{
		{
			name:        "EmptyHeader",
			header:      &Header{},
			expectError: false,
		},
		{
			name: "NegativeDifficulty",
			header: &Header{
				Difficulty: big.NewInt(-1),
			},
			expectError: true,
		},
		{
			name: "NegativeNumber",
			header: &Header{
				Number: big.NewInt(-1),
			},
			expectError: true,
		},
		{
			name: "InconsistentEIP7706",
			header: &Header{
				GasLimits: []uint64{1000},
				// Missing GasUsedVector and ExcessGas
			},
			expectError: false, // Should fall back to pre-EIP7706 encoding
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := rlp.EncodeToBytes(tc.header)
			if tc.expectError {
				if err == nil {
					t.Fatal("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			var decoded Header
			if err := rlp.DecodeBytes(encoded, &decoded); err != nil {
				t.Fatalf("Failed to decode header: %v", err)
			}
		})
	}
}
