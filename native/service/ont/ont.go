package ont

import (
	"bytes"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/common/constants"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/errors"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/utils"
	"io"
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
	native.Register(BIND_NAME, OntBind)
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

	item := utils.GenUInt64StorageItem(toAmount)
	native.GetCacheDB().Put(GenBalanceKey(contract, toAddress), item.ToArray())
	AddTransferNotifications(native, utils.OntContractAddress, &State{To: toAddress, Value: toAmount})
	native.GetCacheDB().Put(GenTotalSupplyKey(contract), item.ToArray())

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

	AddTransferNotifications(native, contract, &State{From: state.From, To: state.To, Value: state.Value})
	return utils.BYTE_TRUE, nil
}

func OntBind(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.GetInput())
	targetChainId, eof := source.NextVarUint()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainId error")
	}
	targetChainContractHash, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] decode targetChainContractHasherror:%s", io.ErrUnexpectedEOF)
	}
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBind] getAdmin, get admin error: %v", err)
	}
	if native.CheckWitness(operatorAddress) == false {
		return utils.BYTE_FALSE, errors.NewErr("[OntBind] authentication failed!")
	}
	native.GetCacheDB().Put(GenBindKey(utils.OntContractAddress, targetChainId), utils.GenVarBytesStorageItem(targetChainContractHash).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.OntContractAddress,
				States:          []interface{}{BIND_NAME, targetChainId, targetChainContractHash},
			})
	}

	return utils.BYTE_TRUE, nil
}

func OntLock(native *native.NativeService) ([]byte, error) {
	contract := utils.OntContractAddress
	source := common.NewZeroCopySource(native.GetInput())

	var lockParam LockParam
	err := lockParam.Deserialization(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] contract params deserialization error:%v", err)
	}

	if lockParam.Args.Value == 0 {
		return utils.BYTE_FALSE, nil
	}
	if lockParam.Args.Value > constants.ONT_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] ont amount:%d over totalSupply:%d", lockParam.Args.Value, constants.ONT_TOTAL_SUPPLY)
	}

	state := &State{
		From:  lockParam.FromAddress,
		To:    contract,
		Value: lockParam.Args.Value,
	}
	_, _, err = Transfer(native, contract, state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindKey(contract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(contractHashBytes) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}

	AddLockNotifications(native, contract, contractHashBytes, &lockParam)

	sink := common.NewZeroCopySink(nil)
	lockParam.Args.Serialization(sink)
	input := getCreateTxArgs(lockParam.ToChainID, contractHashBytes, lockParam.Fee, "unlock", sink.Bytes())
	_, err = native.NativeCall(utils.CrossChainManagerContractAddress, "createTx", input)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] createTx, error:%s", err)
	}

	return utils.BYTE_TRUE, nil
}

func OntUnlock(native *native.NativeService) ([]byte, error) {

	//  this method cannot be invoked by anybody except verifyTxManagerContract
	if !native.CheckWitness(utils.CrossChainManagerContractAddress) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] should be invoked by CrossChainManager Contract, checkwitness failed!")
	}
	contract := utils.OntContractAddress
	source := common.NewZeroCopySource(native.GetInput())

	paramsBytes, eof := source.NextVarBytes()
	var args Args
	err := args.Deserialization(common.NewZeroCopySource(paramsBytes))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] deserialize args error:%s", err)
	}
	fromContractHashBytes, eof := source.NextVarBytes()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input deseriaize from contract hash error!")
	}
	fromChainId, eof := source.NextUint64()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] input deseriaize from chainID error!")
	}
	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindKey(contract, fromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] get bind contract hash with chainID:%d error:%s", fromChainId, err)
	}
	if !bytes.Equal(contractHashBytes, fromContractHashBytes) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] passed in contractHash NOT equal stored contractHash with chainID:%d, expect:%s, got:%s", fromChainId, contractHashBytes, fromContractHashBytes)
	}
	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] parse from bytes to address error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}
	_, _, err = Transfer(native, contract, &State{contract, toAddress, args.Value})
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	AddUnLockNotifications(native, contract, fromChainId, fromContractHashBytes, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}
