// NOTE: This is used instead ov `gen_header_rlp.go` because ATM it's easier for me to
// create new RLP header encoding "by hand" than to figure out how to have conditional ordering
// of the fields in output RLP using just RLP gen code
// Specifically, the issue is that EIP-7706 "removes" some of the previously existing fields,
// Like the `GasUsed` and `GasLimit`, so their RLP encoding is dependant on existence of EIP-7706 fields
package types

import (
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeRLP implements rlp.Encoder interface for Header.
func (h *Header) EncodeRLP(writer io.Writer) error {
	w := rlp.NewEncoderBuffer(writer)

	headerStructAsRLPList := w.List()

	// Encode the fixed fields first
	w.WriteBytes(h.ParentHash[:])
	w.WriteBytes(h.UncleHash[:])
	w.WriteBytes(h.Coinbase[:])
	w.WriteBytes(h.Root[:])
	w.WriteBytes(h.TxHash[:])
	w.WriteBytes(h.ReceiptHash[:])
	w.WriteBytes(h.Bloom[:])

	if err := encodeBigIntOrEmpty(&w, h.Difficulty); err != nil {
		return err
	}
	if err := encodeBigIntOrEmpty(&w, h.Number); err != nil {
		return err
	}

	isEIP7706 := len(h.GasLimits) > 0 && len(h.GasUsedVector) > 0 && len(h.ExcessGas) > 0

	//Per EIP-7706 we remove these fields from the header (because they are replaced with the vectorized counterparts)
	if preEIP7706 := !isEIP7706; preEIP7706 {
		w.WriteUint64(h.GasLimit)
		w.WriteUint64(h.GasUsed)
	}

	w.WriteUint64(h.Time)
	w.WriteBytes(h.Extra)
	w.WriteBytes(h.MixDigest[:])
	w.WriteBytes(h.Nonce[:])

	if isEIP7706 {
		return h.encodeEIP7706(&w, headerStructAsRLPList)
	}

	return h.encodePreEIP7706(&w, headerStructAsRLPList)
}

func (h *Header) encodePreEIP7706(w *rlp.EncoderBuffer, headerStructAsRLPList int) error {
	// Check which optional fields are present
	hasBaseFee := h.BaseFee != nil
	hasWithdrawals := h.WithdrawalsHash != nil
	hasBlobGas := h.BlobGasUsed != nil
	hasExcessBlob := h.ExcessBlobGas != nil
	hasBeaconRoot := h.ParentBeaconRoot != nil
	hasRequests := h.RequestsHash != nil

	// Encode optional fields in order, each field depends on the presence of subsequent fields
	if hasAnyOptionalField := hasBaseFee || hasWithdrawals || hasBlobGas || hasExcessBlob ||
		hasBeaconRoot || hasRequests; hasAnyOptionalField {
		if err := encodeBigIntOrEmpty(w, h.BaseFee); err != nil {
			return err
		}
	}

	if hasSubsequentFields := hasWithdrawals || hasBlobGas || hasExcessBlob ||
		hasBeaconRoot || hasRequests; hasSubsequentFields {
		encodeHashOrEmpty(w, h.WithdrawalsHash)
	}

	if hasSubsequentFields := hasBlobGas || hasExcessBlob || hasBeaconRoot ||
		hasRequests; hasSubsequentFields {
		encodeUint64PtrOrEmpty(w, h.BlobGasUsed)
	}

	if hasSubsequentFields := hasExcessBlob || hasBeaconRoot || hasRequests; hasSubsequentFields {
		encodeUint64PtrOrEmpty(w, h.ExcessBlobGas)
	}

	if hasSubsequentFields := hasBeaconRoot || hasRequests; hasSubsequentFields {
		encodeHashOrEmpty(w, h.ParentBeaconRoot)
	}

	if hasSubsequentFields := hasRequests; hasSubsequentFields {
		encodeHashOrEmpty(w, h.RequestsHash)
	}

	// End the outer list encoding
	w.ListEnd(headerStructAsRLPList)
	return w.Flush()
}

func (h *Header) encodeEIP7706(w *rlp.EncoderBuffer, headerStructAsRLPList int) error {
	encodeHashOrEmpty(w, h.WithdrawalsHash)
	encodeHashOrEmpty(w, h.ParentBeaconRoot)
	encodeHashOrEmpty(w, h.RequestsHash)

	encodeUint64Vector(w, h.GasLimits)
	encodeUint64Vector(w, h.GasUsedVector)
	encodeUint64Vector(w, h.ExcessGas)

	// End the outer list encoding
	w.ListEnd(headerStructAsRLPList)
	return w.Flush()
}

func encodeBigIntOrEmpty(w *rlp.EncoderBuffer, value *big.Int) error {
	if value == nil {
		w.Write(rlp.EmptyString)
		return nil
	}
	if value.Sign() == -1 {
		return rlp.ErrNegativeBigInt
	}
	w.WriteBigInt(value)
	return nil
}

func encodeHashOrEmpty(w *rlp.EncoderBuffer, hash *common.Hash) {
	if hash == nil {
		w.Write(rlp.EmptyString)
	} else {
		w.WriteBytes(hash[:])
	}
}

func encodeUint64PtrOrEmpty(w *rlp.EncoderBuffer, value *uint64) {
	if value == nil {
		w.Write(rlp.EmptyString)
	} else {
		w.WriteUint64(*value)
	}
}

func encodeUint64Vector(w *rlp.EncoderBuffer, vector []uint64) {
	list := w.List()
	for _, value := range vector {
		w.WriteUint64(value)
	}
	w.ListEnd(list)
}

// DecodeRLP implements rlp.Decoder interface for Header.
func (h *Header) DecodeRLP(s *rlp.Stream) error {
	// Start decoding the header list
	_, err := s.List()
	if err != nil {
		return err
	}
	defer s.ListEnd()

	// Decode fixed fields first
	if err := s.Decode(&h.ParentHash); err != nil {
		return err
	}
	if err := s.Decode(&h.UncleHash); err != nil {
		return err
	}
	if err := s.Decode(&h.Coinbase); err != nil {
		return err
	}
	if err := s.Decode(&h.Root); err != nil {
		return err
	}
	if err := s.Decode(&h.TxHash); err != nil {
		return err
	}
	if err := s.Decode(&h.ReceiptHash); err != nil {
		return err
	}
	if err := s.Decode(&h.Bloom); err != nil {
		return err
	}

	h.Difficulty = new(big.Int)
	if err := s.Decode(h.Difficulty); err != nil {
		return err
	}
	h.Number = new(big.Int)
	if err := s.Decode(h.Number); err != nil {
		return err
	}

	//NOTE: here is the tricky part
	//If header is pre-EIP-7706, the fields are as follows
	//GasLimit uint64
	//GasUsed uint64
	//Time uint64
	//Extra []byte
	//
	//But form EIP-7706 fields are:
	//Time
	//Extra[]byte
	//So we need to check if the "pattern" is uint64 []byte or uint64 uint64

	var eitherGasLimitOrTime uint64
	if err := s.Decode(&eitherGasLimitOrTime); err != nil {
		return err
	}

	// assume not EIP-7706
	isEIP7706 := false
	if err := s.Decode(&h.GasUsed); err != nil {
		isEIP7706 = true
	}

	if isEIP7706 {
		h.Time = eitherGasLimitOrTime
	} else {
		h.GasLimit = eitherGasLimitOrTime

		if err := s.Decode(&h.Time); err != nil {
			return err
		}
	}

	var extra = make([]byte, 0)
	if err := s.Decode(&extra); err != nil {
		return err
	}

	h.Extra = extra

	if err := s.Decode(&h.MixDigest); err != nil {
		return err
	}
	if err := s.Decode(&h.Nonce); err != nil {
		return err
	}

	// Handle EIP7706 or pre-EIP7706 optional fields
	if isEIP7706 {
		return h.decodeEIP7706(s)
	}
	return h.decodePreEIP7706(s)
}

func (h *Header) decodePreEIP7706(s *rlp.Stream) error {
	baseFee := new(big.Int)
	if err := s.Decode(baseFee); err != nil {
		return handleErr(err)
	}
	h.BaseFee = baseFee

	if err := s.Decode(&h.WithdrawalsHash); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.BlobGasUsed); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.ExcessBlobGas); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.ParentBeaconRoot); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.RequestsHash); err != nil {
		return handleErr(err)
	}

	return nil
}

func (h *Header) decodeEIP7706(s *rlp.Stream) error {
	if err := s.Decode(&h.WithdrawalsHash); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.ParentBeaconRoot); err != nil {
		return handleErr(err)
	}

	if err := s.Decode(&h.RequestsHash); err != nil {
		return handleErr(err)
	}

	// Decode EIP7706 gas fields
	if err := s.Decode(&h.GasLimits); err != nil {
		return handleErr(err)
	}
	if err := s.Decode(&h.GasUsedVector); err != nil {
		return handleErr(err)
	}
	if err := s.Decode(&h.ExcessGas); err != nil {
		return handleErr(err)
	}

	return nil
}

func handleErr(err error) error {
	if err == rlp.EOL {
		return nil
	}

	return err
}
