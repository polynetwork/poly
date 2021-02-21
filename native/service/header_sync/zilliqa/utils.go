package zilliqa

import (
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func IsHeaderExist(native *native.NativeService, hash []byte, chainID uint64) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return false, fmt.Errorf("IsHeaderExist, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func GetHeaderByHash(native *native.NativeService, hash []byte, chainID uint64) (*core.TxBlockAndDsComm, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize headerBytes from raw storage item err:%v", err)
	}
	var txBlockHeaderAndDsComm core.TxBlockAndDsComm
	if err := json.Unmarshal(storeBytes, &txBlockHeaderAndDsComm); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return &txBlockHeaderAndDsComm, nil
}

func putBlockHeader(native *native.NativeService, txBlockAndComm *core.TxBlockAndDsComm, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	storeBytes, _ := json.Marshal(txBlockAndComm)
	hash := txBlockAndComm.Block.BlockHash[:]
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash),
		cstates.GenRawStorageItem(storeBytes))
	scom.NotifyPutHeader(native, chainID, txBlockAndComm.Block.BlockHeader.BlockNum, util.EncodeHex(hash))
	return nil
}

func putGenesisBlockHeader(native *native.NativeService, txBlockAndDsComm core.TxBlockAndDsComm, chainID uint64) error {
	blockHash := txBlockAndDsComm.Block.BlockHash[:]
	blockNum := txBlockAndDsComm.Block.BlockHeader.BlockNum
	contract := utils.HeaderSyncContractAddress
	storeBytes, _ := json.Marshal(&txBlockAndDsComm)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), blockHash),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockNum)),
		cstates.GenRawStorageItem(blockHash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(blockNum)))
	scom.NotifyPutHeader(native, chainID, blockNum, util.EncodeHex(blockHash))
	return nil
}
