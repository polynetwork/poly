package header_sync

import (
	"fmt"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native"
	hscommon "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/header_sync/neo"
	"github.com/ontio/multi-chain/native/service/header_sync/ont"
	"github.com/ontio/multi-chain/native/service/side_chain_manager"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	SYNC_GENESIS_HEADER = "syncGenesisHeader"
	SYNC_BLOCK_HEADER   = "syncBlockHeader"
)

//Register methods of governance contract
func RegisterHeaderSyncContract(native *native.NativeService) {
	native.Register(SYNC_GENESIS_HEADER, SyncGenesisHeader)
	native.Register(SYNC_BLOCK_HEADER, SyncBlockHeader)
}

func GetChainHandler(chainid uint64) (hscommon.HeaderSyncHandler, error) {
	switch chainid {
	case 2:
		return ont.NewONTHandler(), nil
	case 3:
		return neo.NewNEOHandler(), nil
	default:
		return nil, fmt.Errorf("not a supported chainid:%d", chainid)
	}
}

func SyncGenesisHeader(native *native.NativeService) ([]byte, error) {
	params := new(hscommon.SyncGenesisHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, contract params deserialize error: %v", err)
	}
	chainID := params.ChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.ChainId != chainID {
		return utils.BYTE_FALSE, fmt.Errorf("SyncGenesisHeader, side chain is not registered")
	}

	handler, err := GetChainHandler(chainID)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = handler.SyncGenesisHeader(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}

func SyncBlockHeader(native *native.NativeService) ([]byte, error) {
	params := new(hscommon.SyncBlockHeaderParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, contract params deserialize error: %v", err)
	}
	chainID := params.ChainID

	//check if chainid exist
	sideChain, err := side_chain_manager.GetSideChain(native, chainID)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, side_chain_manager.GetSideChain error: %v", err)
	}
	if sideChain.ChainId != chainID {
		return utils.BYTE_FALSE, fmt.Errorf("SyncBlockHeader, side chain is not registered")
	}

	handler, err := GetChainHandler(chainID)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	err = handler.SyncBlockHeader(native)
	if err != nil {
		return utils.BYTE_FALSE, err
	}
	return utils.BYTE_TRUE, nil
}
