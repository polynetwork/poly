package fisco

import (
	"fmt"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	"github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

func PutFiscoLatestHeightInProcessing(ns *native.NativeService, chainID uint64, fromContract []byte, height uint32) {
	last, _ := GetFiscoLatestHeightInProcessing(ns, chainID, fromContract)
	if height <= last {
		return
	}
	ns.GetCacheDB().Put(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(common.LATEST_HEIGHT_IN_PROCESSING), utils.GetUint64Bytes(chainID), fromContract),
		utils.GetUint32Bytes(height))
}

func GetFiscoLatestHeightInProcessing(ns *native.NativeService, chainID uint64, fromContract []byte) (uint32, error) {
	store, err := ns.GetCacheDB().Get(utils.ConcatKey(utils.CrossChainManagerContractAddress, []byte(common.LATEST_HEIGHT_IN_PROCESSING), utils.GetUint64Bytes(chainID), fromContract))
	if err != nil {
		return 0, fmt.Errorf("GetFiscoRoot, get root error: %v", err)
	}
	if store == nil {
		return 0, fmt.Errorf("GetFiscoRoot, can not find any records")
	}
	raw, err := states.GetValueFromRawStorageItem(store)
	if err != nil {
		return 0, fmt.Errorf("GetFiscoRoot, deserialize from raw storage item err: %v", err)
	}
	return utils.GetBytesUint32(raw), nil
}
