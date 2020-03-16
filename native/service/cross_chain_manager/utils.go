package cross_chain_manager

import (
	"fmt"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

func PutBlackChain(native *native.NativeService, chainID uint64) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes),
		cstates.GenRawStorageItem(chainIDBytes))
}

func CheckIfChainBlacked(native *native.NativeService, chainID uint64) (bool, error) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	chainIDStore, err := native.GetCacheDB().Get(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes))
	if err != nil {
		return true, fmt.Errorf("CheckBlackChain, get black chainIDStore error: %v", err)
	}
	if chainIDStore == nil {
		return false, nil
	}
	return true, nil
}

func RemoveBlackChain(native *native.NativeService, chainID uint64) {
	contract := utils.CrossChainManagerContractAddress
	chainIDBytes := utils.GetUint64Bytes(chainID)
	native.GetCacheDB().Delete(utils.ConcatKey(contract, []byte(BLACKED_CHAIN), chainIDBytes))
}
