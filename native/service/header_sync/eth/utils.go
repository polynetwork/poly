package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	cty "github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/multi-chain/common/config"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	scom "github.com/ontio/multi-chain/native/service/header_sync/common"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	// source from https://github.com/ethereum/go-ethereum/blob/master/consensus/ethash/consensus.go#L45
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

func putBlockHeader(native *native.NativeService, blockHeader types.Header, headerBytes []byte, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	blockHash := blockHeader.Hash().Bytes()

	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockHeader.Number.Uint64())),
		cstates.GenRawStorageItem(headerBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEIGHT), utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(blockHeader.Number.Uint64())))
	notifyPutHeader(native, chainID, blockHeader.Number.Uint64(), hex.EncodeToString(blockHash))
	return nil
}

func GetCurrentHeaderHeight(native *native.NativeService, chainID []byte) (uint64, error) {
	heightStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.CURRENT_HEIGHT), chainID))
	if err != nil {
		return 0, fmt.Errorf("getPrevHeaderHeight error: %v", err)
	}
	if heightStore == nil {
		return 0, fmt.Errorf("getPrevHeaderHeight, heightStore is nil")
	}
	heightBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		return 0, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return utils.GetBytesUint64(heightBytes), err
}

func GetHeaderByHeight(native *native.NativeService, height, chainID uint64) (*cty.Header, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHeight, can not find any header records")
	}
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	var header cty.Header
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize header error: %v", err)
	}
	return &header, nil
}

func notifyPutHeader(native *native.NativeService, chainID uint64, height uint64, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.GetHeight()},
		})
}
