package ont

import (
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/native/service/utils"
	"fmt"
	"github.com/ontio/multi-chain/common/constants"
	"github.com/ontio/multi-chain/errors"
	"math/big"
)

func RegisterOntContract(native *native.NativeService) {
	native.Register(NAME_NAME, OntName)
	native.Register(SYMBOL_NAME, OntSymbol)
	native.Register(DECIMALS_NAME, OntDecimals)
	native.Register(TOTALSUPPLY_NAME, OntTotalSupply)
	native.Register(BALANCEOF_NAME, OntBalanceOf)
	native.Register(ALLOWANCE_NAME, OntAllowance)

	native.Register(INIT_NAME, OntInit)
	native.Register(TRANSFER_NAME, OntTransfer)
	native.Register(APPROVE_NAME, OntApprove)
	native.Register(TRANSFERFROM_NAME, OntTransferFrom)

	native.Register(LOCK_NAME, OntLock)
	native.Register(UNLOCK_NAME, OntUnlock)


}



func OntName(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONT_NAME), nil
}

func OntSymbol(native *native.NativeService) ([]byte, error) {
	return []byte(constants.ONT_SYMBOL), nil
}

func OntDecimals(native *native.NativeService) ([]byte, error) {
	return big.NewInt(int64(constants.ONT_DECIMALS)).Bytes(), nil
}


func OntTotalSupply(native *native.NativeService) ([]byte, error) {
	contract := utils.OntContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntTotalSupply] get totalSupply error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OntBalanceOf(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, TRANSFER_FLAG)
}

func OntAllowance(native *native.NativeService) ([]byte, error) {
	return GetBalanceValue(native, APPROVE_FLAG)
}


func GetBalanceValue(native *native.NativeService, flag byte) ([]byte, error) {
	source := common.NewZeroCopySource(native.GetInput())

	from, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] get from address error!")
	}
	contract := utils.OntContractAddress
	var key []byte
	if flag == APPROVE_FLAG {

		to, eof := source.NextAddress()
		if eof {
			return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] get to address error!")
		}
		key = GenApproveKey(contract, from, to)
	} else if flag == TRANSFER_FLAG {
		key = GenBalanceKey(contract, from)
	}
	amount, err := utils.GetStorageUInt64(native, key)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetBalanceValue] address parse error!")
	}
	return big.NewInt(int64(amount)).Bytes(), nil
}

func OntInit(native *native.NativeService) ([]byte, error) {
	contract := utils.OntContractAddress
	amount, err := utils.GetStorageUInt64(native, GenTotalSupplyKey(contract))
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	if amount > 0 {
		return utils.BYTE_FALSE, fmt.Errorf("Init ont has been completed!")
	}
	toAddress := contract
	toAmount := constants.ONT_TOTAL_SUPPLY

	balanceKey := GenBalanceKey(contract, toAddress)
	item := utils.GenUInt64StorageItem(toAmount)
	native.GetCacheDB().Put(balanceKey, item.ToArray())
	AddTransferNotifications(native, utils.OntContractAddress, &State{To: toAddress, Value: toAmount})
	native.GetCacheDB().Put(GenTotalSupplyKey(contract), utils.GenUInt64StorageItem(toAmount).ToArray())

	return utils.BYTE_TRUE, nil
}


func OntTransfer(native *native.NativeService) ([]byte, error) {
	var transfers Transfers
	source := common.NewZeroCopySource(native.GetInput())
	if err := transfers.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Transfer] Transfers deserialize error!")
	}
	contract := utils.OntContractAddress
	for _, v := range transfers.States {
		if v.Value == 0 {
			continue
		}
		if v.Value > constants.ONT_TOTAL_SUPPLY {
			return utils.BYTE_FALSE, fmt.Errorf("transfer ont amount:%d over totalSupply:%d", v.Value, constants.ONT_TOTAL_SUPPLY)
		}
		_, _, err := Transfer(native, contract, &v)
		if err != nil {
			return utils.BYTE_FALSE, err
		}
		//// TODO: check how to deal with grantOng method
		//fromBalance, toBalance, err := Transfer(native, contract, &v)
		//if err != nil {
		//	return utils.BYTE_FALSE, err
		//}
		//if err := grantOng(native, contract, v.From, fromBalance); err != nil {
		//	return utils.BYTE_FALSE, err
		//}
		//
		//if err := grantOng(native, contract, v.To, toBalance); err != nil {
		//	return utils.BYTE_FALSE, err
		//}

		AddTransferNotifications(native, contract, &v)
	}
	return utils.BYTE_TRUE, nil
}



func OntApprove(native *native.NativeService) ([]byte, error) {
	var state State
	source := common.NewZeroCopySource(native.GetInput())
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntApprove] state deserialize error!")
	}
	if state.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("approve ont amount:%d over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	if native.CheckWitness(state.From) == false {
		return utils.BYTE_FALSE, errors.NewErr("authentication failed!")
	}
	contract := utils.OntContractAddress
	native.GetCacheDB().Put(GenApproveKey(contract, state.From, state.To), utils.GenUInt64StorageItem(state.Value).ToArray())
	return utils.BYTE_TRUE, nil
}



func OntTransferFrom(native *native.NativeService) ([]byte, error) {
	var state TransferFrom
	source := common.NewZeroCopySource(native.GetInput())
	if err := state.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntTransferFrom] State deserialize error:%v", err)
	}
	if state.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if state.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("transferFrom ont amount:%d over totalSupply:%d", state.Value, constants.ONT_TOTAL_SUPPLY)
	}
	contract := utils.OntContractAddress

	_, _, err := TransferedFrom(native, contract, &state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//// TODO: check how to deal with grantOng method
	//fromBalance, toBalance, err := TransferedFrom(native, contract, &state)
	//if err != nil {
	//	return utils.BYTE_FALSE, err
	//}
	//if err := grantOng(native, contract, state.From, fromBalance); err != nil {
	//	return utils.BYTE_FALSE, err
	//}
	//if err := grantOng(native, contract, state.To, toBalance); err != nil {
	//	return utils.BYTE_FALSE, err
	//}
	AddTransferNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}


func OntLock(native *native.NativeService) ([]byte, error) {
	contract := utils.OntContractAddress
	source := common.NewZeroCopySource(native.GetInput())

	toChainID, eof := source.NextUint64()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] input deseriaize toChainID error!")
	}

	from, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] input deseriaize from address error!")
	}

	var args Args
	argsBytes, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] input deseriaize Args{ToAddress, Value} bytes error!")
	}
	if err := args.Deserialization(common.NewZeroCopySource(argsBytes)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] Args deserialize error:%v", err)
	}
	if args.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if args.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] ont amount:%d over totalSupply:%d", args.Value, constants.ONT_TOTAL_SUPPLY)
	}

	state := &State{
		From: from,
		To: contract,
		Value: args.Value,
	}
	_, _, err := Transfer(native, utils.OntContractAddress, state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//targetChainContract := getContractAddressWithChainID(native, contract, toChainID)


	AddLockNotifications(native, contract, toChainID, from, args.ToAddress, args.Value)

	return utils.BYTE_TRUE, nil
}


func OntUnlock(native *native.NativeService) ([]byte, error) {

	return utils.BYTE_TRUE, nil
}