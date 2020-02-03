package ont_lock_proxy

import (
	"bytes"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/errors"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	"io"
)

func RegisterOntLockContract(native *native.NativeService) {
	native.Register(LOCK_NAME, OntLock)
	native.Register(UNLOCK_NAME, OntUnlock)
	native.Register(BIND_NAME, OntBind)
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
				ContractAddress: utils.OntLockProxyContractAddress,
				States:          []interface{}{BIND_NAME, targetChainId, targetChainContractHash},
			})
	}

	return utils.BYTE_TRUE, nil
}

func OntLock(native *native.NativeService) ([]byte, error) {
	lockContract := utils.OntLockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())

	var lockParam LockParam
	err := lockParam.Deserialization(source)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] contract params deserialization error:%v", err)
	}

	if lockParam.Args.Value == 0 {
		return utils.BYTE_FALSE, nil
	}

	state := &ont.State{
		From:  lockParam.FromAddress,
		To:    lockContract,
		Value: lockParam.Args.Value,
	}
	_, _, err = ont.Transfer(native, utils.OntContractAddress, state)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindKey(lockContract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(contractHashBytes) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}

	AddLockNotifications(native, lockContract, contractHashBytes, &lockParam)

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
	ontContract := utils.OntContractAddress
	lockContract := utils.OntLockProxyContractAddress
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
	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindKey(ontContract, fromChainId))
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
	_, _, err = ont.Transfer(native, ontContract, &ont.State{lockContract, toAddress, args.Value})

	if err != nil {
		return utils.BYTE_FALSE, err
	}

	AddUnLockNotifications(native, lockContract, fromChainId, fromContractHashBytes, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}
