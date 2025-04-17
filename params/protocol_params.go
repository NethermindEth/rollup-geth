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

package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const (
	GasLimitBoundDivisor uint64 = 1024               // The bound divisor of the gas limit, used in update calculations.
	MinGasLimit          uint64 = 5000               // Minimum the gas limit may ever be.
	MaxGasLimit          uint64 = 0x7fffffffffffffff // Maximum the gas limit (2^63-1).
	GenesisGasLimit      uint64 = 4712388            // Gas limit of the Genesis block.

	MaximumExtraDataSize  uint64 = 32    // Maximum size extra data may be after Genesis.
	ExpByteGas            uint64 = 10    // Times ceil(log256(exponent)) for the EXP instruction.
	SloadGas              uint64 = 50    // Multiplied by the number of 32-byte words that are copied (round up) for any *COPY operation and added.
	CallValueTransferGas  uint64 = 9000  // Paid for CALL when the value transfer is non-zero.
	CallNewAccountGas     uint64 = 25000 // Paid for CALL when the destination address didn't exist prior.
	TxGas                 uint64 = 21000 // Per transaction not creating a contract. NOTE: Not payable on data of calls between transactions.
	TxGasContractCreation uint64 = 53000 // Per transaction that creates a contract. NOTE: Not payable on data of calls between transactions.
	TxDataZeroGas         uint64 = 4     // Per byte of data attached to a transaction that equals zero. NOTE: Not payable on data of calls between transactions.
	QuadCoeffDiv          uint64 = 512   // Divisor for the quadratic particle of the memory cost equation.
	LogDataGas            uint64 = 8     // Per byte in a LOG* operation's data.
	CallStipend           uint64 = 2300  // Free gas given at beginning of call.

	Keccak256Gas     uint64 = 30 // Once per KECCAK256 operation.
	Keccak256WordGas uint64 = 6  // Once per word of the KECCAK256 operation's data.
	InitCodeWordGas  uint64 = 2  // Once per word of the init code when creating a contract.

	SstoreSetGas    uint64 = 20000 // Once per SSTORE operation.
	SstoreResetGas  uint64 = 5000  // Once per SSTORE operation if the zeroness changes from zero.
	SstoreClearGas  uint64 = 5000  // Once per SSTORE operation if the zeroness doesn't change.
	SstoreRefundGas uint64 = 15000 // Once per SSTORE operation if the zeroness changes to zero.

	NetSstoreNoopGas  uint64 = 200   // Once per SSTORE operation if the value doesn't change.
	NetSstoreInitGas  uint64 = 20000 // Once per SSTORE operation from clean zero.
	NetSstoreCleanGas uint64 = 5000  // Once per SSTORE operation from clean non-zero.
	NetSstoreDirtyGas uint64 = 200   // Once per SSTORE operation from dirty.

	NetSstoreClearRefund      uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot
	NetSstoreResetRefund      uint64 = 4800  // Once per SSTORE operation for resetting to the original non-zero value
	NetSstoreResetClearRefund uint64 = 19800 // Once per SSTORE operation for resetting to the original zero value

	SstoreSentryGasEIP2200            uint64 = 2300  // Minimum gas required to be present for an SSTORE call, not consumed
	SstoreSetGasEIP2200               uint64 = 20000 // Once per SSTORE operation from clean zero to non-zero
	SstoreResetGasEIP2200             uint64 = 5000  // Once per SSTORE operation from clean non-zero to something else
	SstoreClearsScheduleRefundEIP2200 uint64 = 15000 // Once per SSTORE operation for clearing an originally existing storage slot

	ColdAccountAccessCostEIP2929 = uint64(2600) // COLD_ACCOUNT_ACCESS_COST
	ColdSloadCostEIP2929         = uint64(2100) // COLD_SLOAD_COST
	WarmStorageReadCostEIP2929   = uint64(100)  // WARM_STORAGE_READ_COST

	// In EIP-2200: SstoreResetGas was 5000.
	// In EIP-2929: SstoreResetGas was changed to '5000 - COLD_SLOAD_COST'.
	// In EIP-3529: SSTORE_CLEARS_SCHEDULE is defined as SSTORE_RESET_GAS + ACCESS_LIST_STORAGE_KEY_COST
	// Which becomes: 5000 - 2100 + 1900 = 4800
	SstoreClearsScheduleRefundEIP3529 uint64 = SstoreResetGasEIP2200 - ColdSloadCostEIP2929 + TxAccessListStorageKeyGas

	JumpdestGas   uint64 = 1     // Once per JUMPDEST operation.
	EpochDuration uint64 = 30000 // Duration between proof-of-work epochs.

	CreateDataGas         uint64 = 200   //
	CallCreateDepth       uint64 = 1024  // Maximum depth of call/create stack.
	ExpGas                uint64 = 10    // Once per EXP instruction
	LogGas                uint64 = 375   // Per LOG* operation.
	CopyGas               uint64 = 3     //
	StackLimit            uint64 = 1024  // Maximum size of VM stack allowed.
	TierStepGas           uint64 = 0     // Once per operation, for a selection of them.
	LogTopicGas           uint64 = 375   // Multiplied by the * of the LOG*, per LOG transaction. e.g. LOG0 incurs 0 * c_txLogTopicGas, LOG4 incurs 4 * c_txLogTopicGas.
	CreateGas             uint64 = 32000 // Once per CREATE operation & contract-creation transaction.
	Create2Gas            uint64 = 32000 // Once per CREATE2 operation
	CreateNGasEip4762     uint64 = 1000  // Once per CREATEn operations post-verkle
	SelfdestructRefundGas uint64 = 24000 // Refunded following a selfdestruct operation.
	MemoryGas             uint64 = 3     // Times the address of the (highest referenced byte in memory + 1). NOTE: referencing happens on read, write and in instructions such as RETURN and CALL.

	TxDataNonZeroGasFrontier  uint64 = 68    // Per byte of data attached to a transaction that is not equal to zero. NOTE: Not payable on data of calls between transactions.
	TxDataNonZeroGasEIP2028   uint64 = 16    // Per byte of non zero data attached to a transaction after EIP 2028 (part in Istanbul)
	TxTokenPerNonZeroByte     uint64 = 4     // Token cost per non-zero byte as specified by EIP-7623.
	TxCostFloorPerToken       uint64 = 10    // Cost floor per byte of data as specified by EIP-7623.
	TxAccessListAddressGas    uint64 = 2400  // Per address specified in EIP 2930 access list
	TxAccessListStorageKeyGas uint64 = 1900  // Per storage key specified in EIP 2930 access list
	TxAuthTupleGas            uint64 = 12500 // Per auth tuple code specified in EIP-7702

	// These have been changed during the course of the chain
	CallGasFrontier              uint64 = 40  // Once per CALL operation & message call transaction.
	CallGasEIP150                uint64 = 700 // Static portion of gas for CALL-derivates after EIP 150 (Tangerine)
	BalanceGasFrontier           uint64 = 20  // The cost of a BALANCE operation
	BalanceGasEIP150             uint64 = 400 // The cost of a BALANCE operation after Tangerine
	BalanceGasEIP1884            uint64 = 700 // The cost of a BALANCE operation after EIP 1884 (part of Istanbul)
	ExtcodeSizeGasFrontier       uint64 = 20  // Cost of EXTCODESIZE before EIP 150 (Tangerine)
	ExtcodeSizeGasEIP150         uint64 = 700 // Cost of EXTCODESIZE after EIP 150 (Tangerine)
	SloadGasFrontier             uint64 = 50
	SloadGasEIP150               uint64 = 200
	SloadGasEIP1884              uint64 = 800  // Cost of SLOAD after EIP 1884 (part of Istanbul)
	SloadGasEIP2200              uint64 = 800  // Cost of SLOAD after EIP 2200 (part of Istanbul)
	ExtcodeHashGasConstantinople uint64 = 400  // Cost of EXTCODEHASH (introduced in Constantinople)
	ExtcodeHashGasEIP1884        uint64 = 700  // Cost of EXTCODEHASH after EIP 1884 (part in Istanbul)
	SelfdestructGasEIP150        uint64 = 5000 // Cost of SELFDESTRUCT post EIP 150 (Tangerine)

	// EXP has a dynamic portion depending on the size of the exponent
	ExpByteFrontier uint64 = 10 // was set to 10 in Frontier
	ExpByteEIP158   uint64 = 50 // was raised to 50 during Eip158 (Spurious Dragon)

	// Extcodecopy has a dynamic AND a static cost. This represents only the
	// static portion of the gas. It was changed during EIP 150 (Tangerine)
	ExtcodeCopyBaseFrontier uint64 = 20
	ExtcodeCopyBaseEIP150   uint64 = 700

	// CreateBySelfdestructGas is used when the refunded account is one that does
	// not exist. This logic is similar to call.
	// Introduced in Tangerine Whistle (Eip 150)
	CreateBySelfdestructGas uint64 = 25000

	DefaultBaseFeeChangeDenominator = 8          // Bounds the amount the base fee can change between blocks.
	DefaultElasticityMultiplier     = 2          // Bounds the maximum gas limit an EIP-1559 block may have.
	InitialBaseFee                  = 1000000000 // Initial base fee for EIP-1559 blocks.

	MaxCodeSize     = 24576           // Maximum bytecode to permit for a contract
	MaxInitCodeSize = 2 * MaxCodeSize // Maximum initcode to permit in a creation transaction and create instructions

	// Precompiled contract gas prices

	EcrecoverGas        uint64 = 3000 // Elliptic curve sender recovery gas price
	Sha256BaseGas       uint64 = 60   // Base price for a SHA256 operation
	Sha256PerWordGas    uint64 = 12   // Per-word price for a SHA256 operation
	Ripemd160BaseGas    uint64 = 600  // Base price for a RIPEMD160 operation
	Ripemd160PerWordGas uint64 = 120  // Per-word price for a RIPEMD160 operation
	IdentityBaseGas     uint64 = 15   // Base price for a data copy operation
	IdentityPerWordGas  uint64 = 3    // Per-work price for a data copy operation

	Bn256AddGasByzantium             uint64 = 500    // Byzantium gas needed for an elliptic curve addition
	Bn256AddGasIstanbul              uint64 = 150    // Gas needed for an elliptic curve addition
	Bn256ScalarMulGasByzantium       uint64 = 40000  // Byzantium gas needed for an elliptic curve scalar multiplication
	Bn256ScalarMulGasIstanbul        uint64 = 6000   // Gas needed for an elliptic curve scalar multiplication
	Bn256PairingBaseGasByzantium     uint64 = 100000 // Byzantium base price for an elliptic curve pairing check
	Bn256PairingBaseGasIstanbul      uint64 = 45000  // Base price for an elliptic curve pairing check
	Bn256PairingPerPointGasByzantium uint64 = 80000  // Byzantium per-point price for an elliptic curve pairing check
	Bn256PairingPerPointGasIstanbul  uint64 = 34000  // Per-point price for an elliptic curve pairing check

	Bls12381G1AddGas          uint64 = 375   // Price for BLS12-381 elliptic curve G1 point addition
	Bls12381G1MulGas          uint64 = 12000 // Price for BLS12-381 elliptic curve G1 point scalar multiplication
	Bls12381G2AddGas          uint64 = 600   // Price for BLS12-381 elliptic curve G2 point addition
	Bls12381G2MulGas          uint64 = 22500 // Price for BLS12-381 elliptic curve G2 point scalar multiplication
	Bls12381PairingBaseGas    uint64 = 37700 // Base gas price for BLS12-381 elliptic curve pairing check
	Bls12381PairingPerPairGas uint64 = 32600 // Per-point pair gas price for BLS12-381 elliptic curve pairing check
	Bls12381MapG1Gas          uint64 = 5500  // Gas price for BLS12-381 mapping field element to G1 operation
	Bls12381MapG2Gas          uint64 = 23800 // Gas price for BLS12-381 mapping field element to G2 operation

	// The Refund Quotient is the cap on how much of the used gas can be refunded. Before EIP-3529,
	// up to half the consumed gas could be refunded. Redefined as 1/5th in EIP-3529
	RefundQuotient        uint64 = 2
	RefundQuotientEIP3529 uint64 = 5

	BlobTxBytesPerFieldElement         = 32      // Size in bytes of a field element
	BlobTxFieldElementsPerBlob         = 4096    // Number of field elements stored in a single data blob
	BlobTxBlobGasPerBlob               = 1 << 17 // Gas consumption of a single data blob (== blob byte size)
	BlobTxMinBlobGasprice              = 1       // Minimum gas price for data blobs
	BlobTxPointEvaluationPrecompileGas = 50000   // Gas price for the point evaluation precompile.

	HistoryServeWindow = 8192 // Number of blocks to serve historical block hashes for, EIP-2935.
)

// Bls12381G1MultiExpDiscountTable is the gas discount table for BLS12-381 G1 multi exponentiation operation
var Bls12381G1MultiExpDiscountTable = [128]uint64{1000, 949, 848, 797, 764, 750, 738, 728, 719, 712, 705, 698, 692, 687, 682, 677, 673, 669, 665, 661, 658, 654, 651, 648, 645, 642, 640, 637, 635, 632, 630, 627, 625, 623, 621, 619, 617, 615, 613, 611, 609, 608, 606, 604, 603, 601, 599, 598, 596, 595, 593, 592, 591, 589, 588, 586, 585, 584, 582, 581, 580, 579, 577, 576, 575, 574, 573, 572, 570, 569, 568, 567, 566, 565, 564, 563, 562, 561, 560, 559, 558, 557, 556, 555, 554, 553, 552, 551, 550, 549, 548, 547, 547, 546, 545, 544, 543, 542, 541, 540, 540, 539, 538, 537, 536, 536, 535, 534, 533, 532, 532, 531, 530, 529, 528, 528, 527, 526, 525, 525, 524, 523, 522, 522, 521, 520, 520, 519}

// Bls12381G2MultiExpDiscountTable is the gas discount table for BLS12-381 G2 multi exponentiation operation
var Bls12381G2MultiExpDiscountTable = [128]uint64{1000, 1000, 923, 884, 855, 832, 812, 796, 782, 770, 759, 749, 740, 732, 724, 717, 711, 704, 699, 693, 688, 683, 679, 674, 670, 666, 663, 659, 655, 652, 649, 646, 643, 640, 637, 634, 632, 629, 627, 624, 622, 620, 618, 615, 613, 611, 609, 607, 606, 604, 602, 600, 598, 597, 595, 593, 592, 590, 589, 587, 586, 584, 583, 582, 580, 579, 578, 576, 575, 574, 573, 571, 570, 569, 568, 567, 566, 565, 563, 562, 561, 560, 559, 558, 557, 556, 555, 554, 553, 552, 552, 551, 550, 549, 548, 547, 546, 545, 545, 544, 543, 542, 541, 541, 540, 539, 538, 537, 537, 536, 535, 535, 534, 533, 532, 532, 531, 530, 530, 529, 528, 528, 527, 526, 526, 525, 524, 524}

// Difficulty parameters.
var (
	DifficultyBoundDivisor = big.NewInt(2048)   // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(131072) // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(131072) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(13)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)

// System contracts.
var (
	// SystemAddress is where the system-transaction is sent from as per EIP-4788
	SystemAddress = common.HexToAddress("0xfffffffffffffffffffffffffffffffffffffffe")

	// EIP-4788 - Beacon block root in the EVM
	BeaconRootsAddress = common.HexToAddress("0x000F3df6D732807Ef1319fB7B8bB8522d0Beac02")
	BeaconRootsCode    = common.FromHex("3373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500")

	// EIP-2935 - Serve historical block hashes from state
	HistoryStorageAddress = common.HexToAddress("0x0000F90827F1C53a10cb7A02335B175320002935")
	HistoryStorageCode    = common.FromHex("3373fffffffffffffffffffffffffffffffffffffffe14604657602036036042575f35600143038111604257611fff81430311604257611fff9006545f5260205ff35b5f5ffd5b5f35611fff60014303065500")

	// EIP-7002 - Execution layer triggerable withdrawals
	WithdrawalQueueAddress = common.HexToAddress("0x00000961Ef480Eb55e80D19ad83579A64c007002")
	WithdrawalQueueCode    = common.FromHex("3373fffffffffffffffffffffffffffffffffffffffe1460cb5760115f54807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff146101f457600182026001905f5b5f82111560685781019083028483029004916001019190604d565b909390049250505036603814608857366101f457346101f4575f5260205ff35b34106101f457600154600101600155600354806003026004013381556001015f35815560010160203590553360601b5f5260385f601437604c5fa0600101600355005b6003546002548082038060101160df575060105b5f5b8181146101835782810160030260040181604c02815460601b8152601401816001015481526020019060020154807fffffffffffffffffffffffffffffffff00000000000000000000000000000000168252906010019060401c908160381c81600701538160301c81600601538160281c81600501538160201c81600401538160181c81600301538160101c81600201538160081c81600101535360010160e1565b910180921461019557906002556101a0565b90505f6002555f6003555b5f54807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff14156101cd57505f5b6001546002828201116101e25750505f6101e8565b01600290035b5f555f600155604c025ff35b5f5ffd")

	// EIP-7251 - Increase the MAX_EFFECTIVE_BALANCE
	ConsolidationQueueAddress = common.HexToAddress("0x0000BBdDc7CE488642fb579F8B00f3a590007251")
	ConsolidationQueueCode    = common.FromHex("3373fffffffffffffffffffffffffffffffffffffffe1460d35760115f54807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1461019a57600182026001905f5b5f82111560685781019083028483029004916001019190604d565b9093900492505050366060146088573661019a573461019a575f5260205ff35b341061019a57600154600101600155600354806004026004013381556001015f358155600101602035815560010160403590553360601b5f5260605f60143760745fa0600101600355005b6003546002548082038060021160e7575060025b5f5b8181146101295782810160040260040181607402815460601b815260140181600101548152602001816002015481526020019060030154905260010160e9565b910180921461013b5790600255610146565b90505f6002555f6003555b5f54807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff141561017357505f5b6001546001828201116101885750505f61018e565b01600190035b5f555f6001556074025ff35b5f5ffd")

	// RIP-7859 - Expose L1 origin information inside L2 execution environment
	L1OriginContractAddress = common.HexToAddress("0x0000000000000000000000000000000000007859") // TBD address
	L1OriginContractCode    = common.FromHex("608060405234801561000f575f5ffd5b50600436106100cd575f3560e01c8063776891fe1161008a578063c2a99ab411610064578063c2a99ab414610225578063d9f315fe14610243578063de8373e514610273578063dfe5743f14610291576100cd565b8063776891fe146101bb5780637f1bcccd146101d9578063856d71b914610209576100cd565b8063027d6dbb146100d15780633434735f1461010157806336b00a721461011f57806336b21b031461014f57806364f492d11461016d5780636c78d03c1461019d575b5f5ffd5b6100eb60048036038101906100e69190610927565b6102af565b6040516100f8919061096a565b60405180910390f35b6101096102c4565b60405161011691906109c2565b60405180910390f35b61013960048036038101906101349190610927565b6102dc565b604051610146919061096a565b60405180910390f35b6101576102f1565b604051610164919061096a565b60405180910390f35b61018760048036038101906101829190610927565b61034a565b604051610194919061096a565b60405180910390f35b6101a561035f565b6040516101b2919061096a565b60405180910390f35b6101c36103b8565b6040516101d0919061096a565b60405180910390f35b6101f360048036038101906101ee9190610927565b610410565b604051610200919061096a565b60405180910390f35b610223600480360381019061021e9190610a05565b610424565b005b61022d61071a565b60405161023a919061096a565b60405180910390f35b61025d60048036038101906102589190610927565b610773565b60405161026a919061096a565b60405180910390f35b61027b610788565b6040516102889190610a9d565b60405180910390f35b610299610791565b6040516102a6919061096a565b60405180910390f35b5f6102b9826107ea565b600401549050919050565b73fffffffffffffffffffffffffffffffffffffffe81565b5f6102e6826107ea565b600101549050919050565b5f5f60015411610336576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161032d90610b36565b60405180910390fd5b6103416001546107ea565b60010154905090565b5f610354826107ea565b600301549050919050565b5f5f600154116103a4576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161039b90610b36565b60405180910390fd5b6103af6001546107ea565b60030154905090565b5f5f600154116103fd576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016103f490610b36565b60405180910390fd5b6104086001546107ea565b5f0154905090565b5f61041a826107ea565b5f01549050919050565b73fffffffffffffffffffffffffffffffffffffffe73ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146104a6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161049d90610bc4565b60405180910390fd5b5f86116104e8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104df90610c52565b60405180910390fd5b5f5f1b851161052c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161052390610ce0565b60405180910390fd5b5f5f1b8411610570576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161056790610d6e565b60405180910390fd5b5f5f1b83116105b4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ab90610dfc565b60405180910390fd5b5f5f1b82116105f8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105ef90610e8a565b60405180910390fd5b5f5f1b811161063c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161063390610f18565b60405180910390fd5b5f6120008761064b9190610f63565b90506040518060c00160405280878152602001868152602001858152602001848152602001838152602001888152505f5f8381526020019081526020015f205f820151815f01556020820151816001015560408201518160020155606082015181600301556080820151816004015560a082015181600501559050506001548711156106d957866001819055505b867f82e7f14988d91d60d054a4f1f35143fa0ae61e1721ccc4d10f20da0da37a507e87604051610709919061096a565b60405180910390a250505050505050565b5f5f6001541161075f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161075690610b36565b60405180910390fd5b61076a6001546107ea565b60040154905090565b5f61077d826107ea565b600201549050919050565b5f600154905090565b5f5f600154116107d6576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107cd90610b36565b60405180910390fd5b6107e16001546107ea565b60020154905090565b5f5f821161082d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161082490610c52565b60405180910390fd5b600154821115610872576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161086990611003565b60405180910390fd5b5f612000836108819190610f63565b9050825f5f8381526020019081526020015f2060050154146108d8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016108cf90611091565b60405180910390fd5b5f5f8281526020019081526020015f20915050919050565b5f5ffd5b5f819050919050565b610906816108f4565b8114610910575f5ffd5b50565b5f81359050610921816108fd565b92915050565b5f6020828403121561093c5761093b6108f0565b5b5f61094984828501610913565b91505092915050565b5f819050919050565b61096481610952565b82525050565b5f60208201905061097d5f83018461095b565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f6109ac82610983565b9050919050565b6109bc816109a2565b82525050565b5f6020820190506109d55f8301846109b3565b92915050565b6109e481610952565b81146109ee575f5ffd5b50565b5f813590506109ff816109db565b92915050565b5f5f5f5f5f5f60c08789031215610a1f57610a1e6108f0565b5b5f610a2c89828a01610913565b9650506020610a3d89828a016109f1565b9550506040610a4e89828a016109f1565b9450506060610a5f89828a016109f1565b9350506080610a7089828a016109f1565b92505060a0610a8189828a016109f1565b9150509295509295509295565b610a97816108f4565b82525050565b5f602082019050610ab05f830184610a8e565b92915050565b5f82825260208201905092915050565b7f4c314f726967696e536f75726365496d706c3a206e6f204c3120626c6f636b735f8201527f20617661696c61626c6500000000000000000000000000000000000000000000602082015250565b5f610b20602a83610ab6565b9150610b2b82610ac6565b604082019050919050565b5f6020820190508181035f830152610b4d81610b14565b9050919050565b7f4c314f726967696e536f75726365496d706c3a2063616c6c6572206973206e6f5f8201527f74207468652073797374656d2061646472657373000000000000000000000000602082015250565b5f610bae603483610ab6565b9150610bb982610b54565b604082019050919050565b5f6020820190508181035f830152610bdb81610ba2565b9050919050565b7f4c314f726967696e536f75726365496d706c3a206865696768742063616e6e6f5f8201527f74206265207a65726f0000000000000000000000000000000000000000000000602082015250565b5f610c3c602983610ab6565b9150610c4782610be2565b604082019050919050565b5f6020820190508181035f830152610c6981610c30565b9050919050565b7f4c314f726967696e536f75726365496d706c3a20626c6f636b486173682063615f8201527f6e6e6f74206265207a65726f0000000000000000000000000000000000000000602082015250565b5f610cca602c83610ab6565b9150610cd582610c70565b604082019050919050565b5f6020820190508181035f830152610cf781610cbe565b9050919050565b7f4c314f726967696e536f75726365496d706c3a20706172656e74426561636f6e5f8201527f526f6f742063616e6e6f74206265207a65726f00000000000000000000000000602082015250565b5f610d58603383610ab6565b9150610d6382610cfe565b604082019050919050565b5f6020820190508181035f830152610d8581610d4c565b9050919050565b7f4c314f726967696e536f75726365496d706c3a207374617465526f6f742063615f8201527f6e6e6f74206265207a65726f0000000000000000000000000000000000000000602082015250565b5f610de6602c83610ab6565b9150610df182610d8c565b604082019050919050565b5f6020820190508181035f830152610e1381610dda565b9050919050565b7f4c314f726967696e536f75726365496d706c3a2072656365697074526f6f74205f8201527f63616e6e6f74206265207a65726f000000000000000000000000000000000000602082015250565b5f610e74602e83610ab6565b9150610e7f82610e1a565b604082019050919050565b5f6020820190508181035f830152610ea181610e68565b9050919050565b7f4c314f726967696e536f75726365496d706c3a207472616e73616374696f6e525f8201527f6f6f742063616e6e6f74206265207a65726f0000000000000000000000000000602082015250565b5f610f02603283610ab6565b9150610f0d82610ea8565b604082019050919050565b5f6020820190508181035f830152610f2f81610ef6565b9050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b5f610f6d826108f4565b9150610f78836108f4565b925082610f8857610f87610f36565b5b828206905092915050565b7f4c314f726967696e536f75726365496d706c3a20626c6f636b206865696768745f8201527f20746f6f20686967680000000000000000000000000000000000000000000000602082015250565b5f610fed602983610ab6565b9150610ff882610f93565b604082019050919050565b5f6020820190508181035f83015261101a81610fe1565b9050919050565b7f4c314f726967696e536f75726365496d706c3a20626c6f636b2064617461206e5f8201527f6f7420666f756e64206f72206f7665727772697474656e000000000000000000602082015250565b5f61107b603783610ab6565b915061108682611021565b604082019050919050565b5f6020820190508181035f8301526110a88161106f565b905091905056fea264697066735822122097d117e54ba752323cec0ba2bd8328ff99cd465cf2aac260ec8b69f4c43e32b764736f6c634300081d0033")
)
