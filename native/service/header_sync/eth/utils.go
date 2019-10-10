package eth

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/multi-chain/common/config"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

func PutBlockHeader(native *native.NativeService, blockHeader types.Header, headerbytes []byte) error {
	contract := utils.HeaderSyncContractAddress
	blockHash := blockHeader.Hash().Bytes()

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.BLOCK_HEADER), utils.ETH_CHAIN_ID_BYTE, blockHash),
		cstates.GenRawStorageItem(headerbytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.ETH_CHAIN_ID_BYTE, blockHeader.Number.Bytes()),
		cstates.GenRawStorageItem(blockHash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEIGHT), utils.ETH_CHAIN_ID_BYTE), cstates.GenRawStorageItem(blockHeader.Number.Bytes()))
	notifyPutHeader(native, utils.ETH_CHAIN_ID, uint32(blockHeader.Number.Int64()), hex.EncodeToString(blockHash))
	return nil
}

func notifyPutHeader(native *native.NativeService, chainID uint64, height uint32, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.GetHeight()},
		})
}
