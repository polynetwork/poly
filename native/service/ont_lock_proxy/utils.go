package ont_lock_proxy

import (
	"encoding/hex"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
	ontcommon "github.com/ontio/ontology/common"
	ontccm "github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
)

const (
	BALANCEOF_NAME = "balanceOf"
	BIND_NAME      = "bind"
	LOCK_NAME      = "lock"
	UNLOCK_NAME    = "unlock"
)

func GenBindKey(contract common.Address, chainId uint64) []byte {
	chainIdBytes := utils.GetUint64Bytes(chainId)
	temp := append(contract[:], []byte(BIND_NAME)...)
	return append(temp, chainIdBytes...)
}

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
