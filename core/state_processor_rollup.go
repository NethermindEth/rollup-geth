// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// Data for the L1OriginSource system  contract (RIP-7859).
type L1OriginSource struct {
	blockHash        common.Hash
	parentBeaconRoot common.Hash
	stateRoot        common.Hash
	receiptRoot      common.Hash
	transactionRoot  common.Hash
	blockHeight      *big.Int
}

// Encodes the system contract function call to update the L1OriginSource data.
func (l *L1OriginSource) UpdateL1OriginSourceCallData() []byte {
	methodID := crypto.Keccak256([]byte("updateL1BlockData(uint256,bytes32,bytes32,bytes32,bytes32,bytes32)"))[0:4]

	data := make([]byte, 4+32*6) // 4 bytes for method ID + 6 parameters of 32 bytes each
	copy(data[0:4], methodID)

	heightBytes := common.LeftPadBytes(l.blockHeight.Bytes(), 32)
	copy(data[4:36], heightBytes)
	copy(data[36:68], l.blockHash[:])
	copy(data[68:100], l.parentBeaconRoot[:])
	copy(data[100:132], l.stateRoot[:])
	copy(data[132:164], l.receiptRoot[:])
	copy(data[164:196], l.transactionRoot[:])

	return data
}

// ProcessL1OriginBlockInfo stores the L1 block info in the L1OriginSource contract
// as defined in RIP-7859.
func ProcessL1OriginBlockInfo(l1OriginSource *L1OriginSource, evm *vm.EVM) {
	if tracer := evm.Config.Tracer; tracer != nil {
		onSystemCallStart(tracer, evm.GetVMContext())
		if tracer.OnSystemCallEnd != nil {
			defer tracer.OnSystemCallEnd()
		}
	}

	msg := &Message{
		From:      params.SystemAddress,
		GasLimit:  30_000_000,
		GasPrice:  common.Big0,
		GasFeeCap: common.Big0,
		GasTipCap: common.Big0,
		To:        &params.L1OriginContractAddress,
		Data:      l1OriginSource.UpdateL1OriginSourceCallData(),
	}

	evm.SetTxContext(NewEVMTxContext(msg))
	evm.StateDB.AddAddressToAccessList(params.L1OriginContractAddress)
	_, _, err := evm.Call(msg.From, *msg.To, msg.Data, msg.GasLimit, common.U2560)

	if err != nil {
		panic(fmt.Errorf("failed to process L1 block info: %v", err))
	}

	if evm.StateDB.AccessEvents() != nil {
		evm.StateDB.AccessEvents().Merge(evm.AccessEvents)
	}
	evm.StateDB.Finalise(true)
}
