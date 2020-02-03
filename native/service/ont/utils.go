package ont

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	TOTAL_SUPPLY_NAME = "totalSupply"
	INIT_NAME         = "init"
	TRANSFER_NAME     = "transfer"
	APPROVE_NAME      = "approve"
	TRANSFERFROM_NAME = "transferFrom"
	NAME_NAME         = "name"
	SYMBOL_NAME       = "symbol"
	DECIMALS_NAME     = "decimals"
	TOTALSUPPLY_NAME  = "totalSupply"
	BALANCEOF_NAME    = "balanceOf"
	ALLOWANCE_NAME    = "allowance"
)

const (
	TRANSFER_FLAG byte = 1
	APPROVE_FLAG  byte = 2
)

func GenTotalSupplyKey(contract common.Address) []byte {
	return append(contract[:], TOTAL_SUPPLY_NAME...)
}

func GenBalanceKey(contract, addr common.Address) []byte {
	return append(contract[:], addr[:]...)
}

func GenApproveKey(contract, from, to common.Address) []byte {
	temp := append(contract[:], from[:]...)
	return append(temp, to[:]...)
}
func AddTransferNotifications(native *native.NativeService, contract common.Address, state *State) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.AddNotify(
		&event.NotifyEventInfo{
			ContractAddress: contract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}

func Transfer(native *native.NativeService, contract common.Address, state *State) (uint64, uint64, error) {
	if !native.CheckWitness(state.From) {
		return 0, 0, fmt.Errorf("authentication failed!")
	}

	fromBalance, err := fromTransfer(native, contract, state.From, state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, contract, state.To, state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func TransferedFrom(native *native.NativeService, contract common.Address, state *TransferFrom) (uint64, uint64, error) {
	if native.CheckWitness(state.Sender) == false {
		return 0, 0, fmt.Errorf("authentication failed!")
	}

	if err := fromApprove(native, genTransferFromKey(contract, state), state.Value); err != nil {
		return 0, 0, err
	}

	fromBalance, err := fromTransfer(native, contract, state.From, state.Value)
	if err != nil {
		return 0, 0, err
	}

	toBalance, err := toTransfer(native, contract, state.To, state.Value)
	if err != nil {
		return 0, 0, err
	}
	return fromBalance, toBalance, nil
}

func fromTransfer(native *native.NativeService, contract, from common.Address, value uint64) (uint64, error) {
	fromKey := GenBalanceKey(contract, from)
	fromBalance, err := utils.GetStorageUInt64(native, fromKey)
	if err != nil {
		return 0, err
	}
	if fromBalance < value {
		addr, _ := common.AddressParseFromBytes(fromKey[20:])
		return 0, fmt.Errorf("[Transfer] balance insufficient. contract:%s, account:%s,balance:%d, transfer amount:%d",
			contract.ToHexString(), addr.ToBase58(), fromBalance, value)
	} else if fromBalance == value {
		native.GetCacheDB().Delete(fromKey)
	} else {
		native.GetCacheDB().Put(fromKey, utils.GenUInt64StorageItem(fromBalance-value).ToArray())
	}
	return fromBalance, nil
}

func toTransfer(native *native.NativeService, contract, to common.Address, value uint64) (uint64, error) {
	toKey := GenBalanceKey(contract, to)
	toBalance, err := utils.GetStorageUInt64(native, toKey)
	if err != nil {
		return 0, err
	}
	native.GetCacheDB().Put(toKey, utils.GenUInt64StorageItem(toBalance+value).ToArray())
	return toBalance, nil
}

func fromApprove(native *native.NativeService, fromApproveKey []byte, value uint64) error {
	approveValue, err := utils.GetStorageUInt64(native, fromApproveKey)
	if err != nil {
		return err
	}
	if approveValue < value {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue, value)
	} else if approveValue == value {
		native.GetCacheDB().Delete(fromApproveKey)
	} else {
		native.GetCacheDB().Put(fromApproveKey, utils.GenUInt64StorageItem(approveValue-value).ToArray())
	}
	return nil
}

func genTransferFromKey(contract common.Address, state *TransferFrom) []byte {
	temp := append(contract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}
