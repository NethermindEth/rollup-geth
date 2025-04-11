// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

interface L1OriginSource {
    function getL1OriginBlockHash() external view returns (bytes32 blockHash);
    function getL1OriginParentBeaconRoot() external view returns (bytes32 blockHash);
    function getL1OriginStateRoot() external view returns (bytes32 stateRoot);
    function getL1OriginReceiptRoot() external view returns (bytes32 receiptRoot);
    function getL1OriginTransactionRoot() external view returns (bytes32 transactionRoot);
    function getL1OriginBlockHeight() external view returns (uint256 blockHeight);

    function getL1OriginBlockHashAt(uint256 height) external view returns (bytes32 blockHash);
    function getL1OriginParentBeaconRootAt(uint256 height) external view returns (bytes32 blockHash);
    function getL1OriginStateRootAt(uint256 height) external view returns (bytes32 stateRoot);
    function getL1OriginReceiptRootAt(uint256 height) external view returns (bytes32 receiptRoot);
    function getL1OriginTransactionRootAt(uint256 height) external view returns (bytes32 transactionRoot);
}

/**
 * @title L1Origin
 * @dev Implementation of the L1OriginSource interface that stores L1 block data
 * using a circular buffer of 8192 blocks
 */
contract L1Origin is L1OriginSource {
    struct L1BlockData {
        bytes32 blockHash;
        bytes32 parentBeaconRoot;
        bytes32 stateRoot;
        bytes32 receiptRoot;
        bytes32 transactionRoot;
        uint256 blockHeight;
    }

    // Maximum number of L1 blocks to store
    uint256 private constant MAX_STORED_BLOCKS = 8192;

    // Circular buffer of block data
    mapping(uint256 => L1BlockData) private blockData;

    // Current L1 block height
    uint256 private currentL1BlockHeight;

    // Fixed system address that can update the L1 data
    address public constant SYSTEM_ADDRESS = 0xffffFFFfFFffffffffffffffFfFFFfffFFFfFFfE;

    // Events                                                                                                         .. 
    event L1DataUpdated(uint256 indexed height, bytes32 blockHash);     

    /**
     * @dev Restricts function to the system address
     */
    modifier onlySystem() {
        require(msg.sender == SYSTEM_ADDRESS, "L1Origin: caller is not the system address");
        _;
    }

    /**
     * @dev Updates the L1 block data for a specific height
     * @param height The L1 block height
     * @param blockHash The L1 block hash
     * @param parentBeaconRoot The parent beacon root
     * @param stateRoot The state root
     * @param receiptRoot The receipt root
     * @param transactionRoot The transaction root
     */
    function updateL1BlockData(
        uint256 height,
        bytes32 blockHash,
        bytes32 parentBeaconRoot,
        bytes32 stateRoot,
        bytes32 receiptRoot,
        bytes32 transactionRoot
    ) external onlySystem {
        require(height > 0, "L1Origin: height cannot be zero");

        // Store in the buffer using modulo to get the buffer index
        uint256 bufferIndex = height % MAX_STORED_BLOCKS;

        blockData[bufferIndex] = L1BlockData({
            blockHash: blockHash,
            parentBeaconRoot: parentBeaconRoot,
            stateRoot: stateRoot,
            receiptRoot: receiptRoot,
            transactionRoot: transactionRoot,
            blockHeight: height
        });

        // Update the current height if this is a new highest block
        if (height > currentL1BlockHeight) {
            currentL1BlockHeight = height;
        }

        emit L1DataUpdated(height, blockHash);
    }

    /**
     * @dev Get block data by height, internal function
     */
    function _getBlockDataAt(uint256 height) private view returns (L1BlockData storage) {
        require(height > 0, "L1Origin: height cannot be zero");
        require(height <= currentL1BlockHeight, "L1Origin: block height too high");
        uint256 bufferIndex = height % MAX_STORED_BLOCKS;
        // Check that the buffer contains the requested height
        require(blockData[bufferIndex].blockHeight == height, 
                "L1Origin: block data not found or overwritten");
        return blockData[bufferIndex];
    }

    function getL1OriginBlockHash() external view override returns (bytes32 blockHash) {
        require(currentL1BlockHeight > 0, "L1Origin: no L1 blocks available");
        return _getBlockDataAt(currentL1BlockHeight).blockHash;
    }

    function getL1OriginParentBeaconRoot() external view override returns (bytes32 blockHash) {
        require(currentL1BlockHeight > 0, "L1Origin: no L1 blocks available");
        return _getBlockDataAt(currentL1BlockHeight).parentBeaconRoot;
    }

    function getL1OriginStateRoot() external view override returns (bytes32 stateRoot) {
        require(currentL1BlockHeight > 0, "L1Origin: no L1 blocks available");
        return _getBlockDataAt(currentL1BlockHeight).stateRoot;
    }

    function getL1OriginReceiptRoot() external view override returns (bytes32 receiptRoot) {
        require(currentL1BlockHeight > 0, "L1Origin: no L1 blocks available");
        return _getBlockDataAt(currentL1BlockHeight).receiptRoot;
    }

    function getL1OriginTransactionRoot() external view override returns (bytes32 transactionRoot) {
        require(currentL1BlockHeight > 0, "L1Origin: no L1 blocks available");
        return _getBlockDataAt(currentL1BlockHeight).transactionRoot;
    }

    function getL1OriginBlockHeight() external view override returns (uint256 blockHeight) {
        return currentL1BlockHeight;
    }

    function getL1OriginBlockHashAt(uint256 height) external view override returns (bytes32 blockHash) {
        return _getBlockDataAt(height).blockHash;
    }

    function getL1OriginParentBeaconRootAt(uint256 height) external view override returns (bytes32 blockHash) {
        return _getBlockDataAt(height).parentBeaconRoot;
    }

    function getL1OriginStateRootAt(uint256 height) external view override returns (bytes32 stateRoot) {
        return _getBlockDataAt(height).stateRoot;
    }

    function getL1OriginReceiptRootAt(uint256 height) external view override returns (bytes32 receiptRoot) {
        return _getBlockDataAt(height).receiptRoot;
    }

    function getL1OriginTransactionRootAt(uint256 height) external view override returns (bytes32 transactionRoot) {
        return _getBlockDataAt(height).transactionRoot;
    }
}