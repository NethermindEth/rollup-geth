package rawdb

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

// Tests block header storage and retrieval operations when base fees are set
func TestHeaderStorageIncludingBaseFees(t *testing.T) {
	db := NewMemoryDatabase()

	// Create a test header to move around the database and make sure it's really new
	baseFees := types.VectorFeeBigint{big.NewInt(1), big.NewInt(2), big.NewInt(3)}
	header := &types.Header{Number: big.NewInt(42), Extra: []byte("test header base fees"), BaseFees: &baseFees}
	if entry := ReadHeader(db, header.Hash(), header.Number.Uint64()); entry != nil {
		t.Fatalf("Non existent header returned: %v", entry)
	}
	// Write and verify the header in the database
	WriteHeader(db, header)
	if entry := ReadHeader(db, header.Hash(), header.Number.Uint64()); entry == nil {
		t.Fatalf("Stored header not found")
	} else if entry.Hash() != header.Hash() {
		t.Fatalf("Retrieved header mismatch: have %v, want %v", entry, header)
	} else if entry.BaseFees == nil {
		t.Fatalf("Base fees are nil")
	} else if !entry.BaseFees.VectorAllEq(baseFees) {
		t.Fatalf("Base fees mismatch: have %v, want %v", entry.BaseFees, baseFees)
	}

	if entry := ReadHeaderRLP(db, header.Hash(), header.Number.Uint64()); entry == nil {
		t.Fatalf("Stored header RLP not found")
	} else {
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(entry)

		if hash := common.BytesToHash(hasher.Sum(nil)); hash != header.Hash() {
			t.Fatalf("Retrieved RLP header mismatch: have %v, want %v", entry, header)
		}
	}
	// Delete the header and verify the execution
	DeleteHeader(db, header.Hash(), header.Number.Uint64())
	if entry := ReadHeader(db, header.Hash(), header.Number.Uint64()); entry != nil {
		t.Fatalf("Deleted header returned: %v", entry)
	}

	if baseFeesEntry := ReadHeaderBaseFees(db, header.Hash()); baseFeesEntry != nil {
		t.Fatalf("Deleted header returned base fees: %v", baseFeesEntry)
	}
}

// Tests block header storage and retrieval operations when there are no base fees
func TestHeaderStorageNoBaseFees(t *testing.T) {
	db := NewMemoryDatabase()

	// Create a test header to move around the database and make sure it's really new
	header := &types.Header{Number: big.NewInt(42), Extra: []byte("test header no base fees")}
	if entry := ReadHeader(db, header.Hash(), header.Number.Uint64()); entry != nil {
		t.Fatalf("Non existent header returned: %v", entry)
	}
	// Write and verify the header in the database
	WriteHeader(db, header)
	if entry := ReadHeader(db, header.Hash(), header.Number.Uint64()); entry == nil {
		t.Fatalf("Stored header not found")
	} else if entry.Hash() != header.Hash() {
		t.Fatalf("Retrieved header mismatch: have %v, want %v", entry, header)
	} else if entry.BaseFees != nil {
		t.Fatalf("Base fees are not nil")
	}

	if baseFeesEntry := ReadHeaderBaseFees(db, header.Hash()); baseFeesEntry != nil {
		t.Fatalf("Base fees entry should not be set, returned %v", baseFeesEntry)
	}
}
