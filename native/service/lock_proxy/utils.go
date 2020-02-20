package lock_proxy

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	ontcommon "github.com/ontio/ontology/common"
	ontccm "github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"math/big"
)

const (
	LOCK_NAME               = "lock"
	UNLOCK_NAME             = "unlock"
	BIND_PROXY_NAME         = "bindProxy"
	BIND_ASSET_NAME         = "bindAsset"
	GET_PROXY_HASH_NAME     = "getProxyHash"
	GET_ASSET_HASH_NAME     = "getAssetHash"
	GET_CROSSED_LIMIT_NAME  = "getCrossedLimit"
	GET_CROSSED_AMOUNT_NAME = "getCrossedAmount"

	TARGET_ASSET_HASH_PEFIX = "TargetAssetHash"
	CROSS_LIMIT_PREFIX      = "AssetCrossLimit"
	CROSS_AMOUNT_PREFIX     = "AssetCrossedAmount"
)

func AddLockNotifications(native *native.NativeService, contract common.Address, toContract []byte, targetAssetContract []byte, param *LockParam) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{LOCK_NAME, param.FromAddress.ToBase58(), param.ToChainID, hex.EncodeToString(toContract), hex.EncodeToString(targetAssetContract), hex.EncodeToString(param.ToAddress), param.Value},
		})
}

func AddUnLockNotifications(native *native.NativeService, contract common.Address, fromChainId uint64, fromProxyContract []byte, targetAssetHash common.Address, toAddress common.Address, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{UNLOCK_NAME, fromChainId, hex.EncodeToString(fromProxyContract), hex.EncodeToString(targetAssetHash[:]), toAddress.ToBase58(), amount},
		})
}

func getCreateTxArgs(toChainID uint64, contractHashBytes []byte, method string, argsBytes []byte) []byte {
	createCrossChainTxParam := &ontccm.CreateCrossChainTxParam{
		ToChainID:         toChainID,
		ToContractAddress: contractHashBytes,
		Method:            method,
		Args:              argsBytes,
	}
	sink := ontcommon.NewZeroCopySink(nil)
	createCrossChainTxParam.Serialization(sink)
	return sink.Bytes()
}

func getTransferInput(state ont.State) []byte {
	var transfers ont.Transfers
	transfers.States = []ont.State{state}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)
	return sink.Bytes()
}

func GenBindProxyKey(contract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_PROXY_NAME)...)
	return append(temp, chainIdBytes...)
}

func GenBindAssetHashKey(contract, assetContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(TARGET_ASSET_HASH_PEFIX)...)
	temp = append(temp, assetContract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedLimitKey(contract, assetContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(BIND_ASSET_NAME)...)
	temp = append(temp, []byte(CROSS_LIMIT_PREFIX)...)
	temp = append(temp, assetContract[:]...)
	return append(temp, chainIdBytes...)
}

func GenCrossedAmountKey(contract, sourceContract common.Address, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], []byte(CROSS_AMOUNT_PREFIX)...)
	temp = append(temp, sourceContract[:]...)
	return append(temp, chainIdBytes...)
}

func getAmount(native *native.NativeService, storgedKey []byte) (*big.Int, error) {
	valueBs, err := utils.GetStorageVarBytes(native, storgedKey)
	if err != nil {
		return nil, fmt.Errorf("getAmount, error:%s", err)
	}
	value := big.NewInt(0).SetBytes(valueBs)
	return value, nil
}
