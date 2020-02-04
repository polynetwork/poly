package ont_lock_proxy

import (
	"encoding/hex"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/ont"
	ontcommon "github.com/ontio/ontology/common"
	ontccm "github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
)

const (
	LOCK_NAME       = "lock"
	UNLOCK_NAME     = "unlock"
	BIND_PROXY_NAME = "bindProxy"
	BIND_ASSET_NAME = "bindAsset"
)

func AddLockNotifications(native *native.NativeService, contract common.Address, toContract []byte, param *LockParam) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{LOCK_NAME, param.FromAddress.ToBase58(), param.ToChainID, hex.EncodeToString(toContract), hex.EncodeToString(param.Args.ToAddress), param.Args.Value},
		})
}

func AddUnLockNotifications(native *native.NativeService, contract common.Address, fromChainId uint64, fromContract []byte, toAddress common.Address, amount uint64) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{UNLOCK_NAME, fromChainId, hex.EncodeToString(fromContract), toAddress.ToBase58(), amount},
		})
}

func getCreateTxArgs(toChainID uint64, contractHashBytes []byte, fee uint64, method string, argsBytes []byte) []byte {
	createCrossChainTxParam := &ontccm.CreateCrossChainTxParam{
		ToChainID:         toChainID,
		ToContractAddress: contractHashBytes,
		Fee:               fee,
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

func GenBindAssetKey(contract common.Address, assetContract []byte, chainId uint64) []byte {
	sink := common.NewZeroCopySink(nil)
	sink.WriteUint64(chainId)
	chainIdBytes := sink.Bytes()
	temp := append(contract[:], assetContract...)
	temp = append(temp, []byte(BIND_ASSET_NAME)...)
	return append(temp, chainIdBytes...)
}
