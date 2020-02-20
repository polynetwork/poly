package ong

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/constants"
	"github.com/ontio/multi-chain/errors"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	"math/big"
)

func RegisterOngContract(native *native.NativeService) {
	native.Register(ont.NAME_NAME, OngName)
	native.Register(ont.SYMBOL_NAME, OngSymbol)
	native.Register(ont.DECIMALS_NAME, OngDecimals)
	native.Register(ont.TOTALSUPPLY_NAME, OngTotalSupply)
	native.Register(ont.BALANCEOF_NAME, OngBalanceOf)
	native.Register(ont.ALLOWANCE_NAME, OngAllowance)
	native.Register(ont.INIT_NAME, OngInit)
	native.Register(ont.TRANSFER_NAME, OngTransfer)
	native.Register(ont.APPROVE_NAME, OngApprove)
	native.Register(ont.TRANSFERFROM_NAME, OngTransferFrom)
}

func OngName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_NAME), nil
}

func OngSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONG_SYMBOL), nil
}

func OngDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONG_DECIMALS)).Bytes(), nil
}

func OngTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := utils.OngContractAddress
	amount, err := utils.GetStorageUInt64(native, ont.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngTotalSupply] get totalSupply error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OngBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, ont.TRANSFER_FLAG)
}

func OngAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, ont.APPROVE_FLAG)
}

func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	contract := utils.OngContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	from, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] get from address error!")
	}
	var key []byte
	if flag == ont.APPROVE_FLAG {

		to, eof := source.NextAddress()
		if eof {
			return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] get to address error!")
		}
		key = ont.GenApproveKey(contract, from, to)
	} else if flag == ont.TRANSFER_FLAG {
		key = ont.GenBalanceKey(contract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] address parse error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OngInit(native *native.NativeService) ([]byte, error) {
	contract := utils.OngContractAddress
	amount, err := utils.GetStorageUInt64(native, ont.GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, fmt.Errorf("Init ong has been completed!")
	}
	toAddress := utils.LockProxyContractAddress
	toAmount := constants.ONG_TOTAL_SUPPLY

	item := utils.GenUInt64StorageItem(toAmount)
	native.GetCacheDB().Put(ont.GenBalanceKey(contract, toAddress), item.ToArray())
	native.GetCacheDB().Put(ont.GenTotalSupplyKey(contract), item.ToArray())
	ont.AddTransferNotifications(native, contract, &ont.State{To: toAddress, Value: toAmount})
	return utils.BYTE_TRUE, nil
}

func OngTransfer(native *native.NativeService) ([]byte, error) {
	contract := utils.OngContractAddress
	var transfers ont.Transfers
	source := common.NewZeroCopySource(native.GetInput())
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Transfer] Transfers deserialize error!")
	}
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		if v.Value > constants.ONG_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ong amount:%d over totalSupply:%d", v.Value, constants.ONG_TOTAL_SUPPLY)
		}
		_, _, err := ont.Transfer(native, contract, &v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}

		ont.AddTransferNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}

func OngApprove(native *native.NativeService) ([]byte, error) {
	contract := utils.OngContractAddress

	var state ont.State
	source := common.NewZeroCopySource(native.GetInput())
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngApprove] state deserialize error!")
	}
	if state.Value > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("[OngApprove] approve ont amount:%d over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}
	if !native.CheckWitness(state.From) {
		return utils.BYTE_FALSE, errors.NewErr("[OngApprove] authentication failed!")
	}
	native.GetCacheDB().Put(ont.GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}

func OngTransferFrom(native *native.NativeService) ([]byte, error) {
	contract := utils.OngContractAddress
	var state ont.TransferFrom
	source := common.NewZeroCopySource(native.GetInput())
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OngTransferFrom] State deserialize error:%v", err)
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom ong amount:%d over totalSupply:%d", state.Value, constants.ONG_TOTAL_SUPPLY)
	}

	_, _, err := ont.TransferedFrom(native, contract, &state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	ont.AddTransferNotifications(native, contract, &ont.State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}
