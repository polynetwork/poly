package ont_lock_proxy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/common/constants"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/errors"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	"io"
)

func RegisterOntLockContract(native *native.NativeService) {
	native.Register(LOCK_NAME, OntLock)
	native.Register(UNLOCK_NAME, OntUnlock)
	native.Register(BIND_PROXY_NAME, OntBindProxyHash)
	native.Register(BIND_ASSET_NAME, OntBindAssetHash)
}

func OntBindProxyHash(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.GetInput())
	var bindParam BindProxyParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindProxyHash] deserialize BindParam error:%s", io.ErrUnexpectedEOF)
	}

	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindProxyHash] get operator error: %v", err)
	}
	if !native.CheckWitness(operatorAddress) {
		return utils.BYTE_FALSE, errors.NewErr("[OntBindProxyHash] authentication failed!")
	}
	native.GetCacheDB().Put(GenBindProxyKey(utils.OntLockProxyContractAddress, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetHash).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.OntLockProxyContractAddress,
				States:          []interface{}{BIND_PROXY_NAME, bindParam.TargetChainId, bindParam.TargetHash},
			})
	}

	return utils.BYTE_TRUE, nil
}

func OntBindAssetHash(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.GetInput())
	var bindParam BindAssetParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindAssetHash] deserialize BindParam error:%s", io.ErrUnexpectedEOF)
	}

	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntBindAssetHash] get operator error: %v", err)
	}
	if !native.CheckWitness(operatorAddress) {
		return utils.BYTE_FALSE, errors.NewErr("[OntBindAssetHash] authentication failed!")
	}
	native.GetCacheDB().Put(GenBindAssetKey(utils.OntLockProxyContractAddress, bindParam.SourceAssetHash[:], bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetAssetHash).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: utils.OntLockProxyContractAddress,
				States:          []interface{}{BIND_PROXY_NAME, hex.EncodeToString(bindParam.SourceAssetHash[:]), bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetAssetHash)},
			})
	}

	return utils.BYTE_TRUE, nil
}

func OntLock(native *native.NativeService) ([]byte, error) {
	lockContract := utils.OntLockProxyContractAddress
	ontContract := utils.OntContractAddress
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
	// currently, only support ont
	if !bytes.Equal(lockParam.Args.AssetHash, ontContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] only support ont lock, expect:%s, but got:%s", hex.EncodeToString(ontContract[:]), hex.EncodeToString(lockParam.Args.AssetHash))
	}

	state := ont.State{
		From:  lockParam.FromAddress,
		To:    lockContract,
		Value: lockParam.Args.Value,
	}
	transferInput := getTransferInput(state)
	res, err := native.NativeCall(utils.OntContractAddress, ont.TRANSFER_NAME, transferInput)
	if !bytes.Equal(res.([]byte), utils.BYTE_TRUE) || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] Transfer Ont, error:%s", err)
	}

	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindProxyKey(lockContract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(contractHashBytes) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[OntLock] get bind contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}

	AddLockNotifications(native, lockContract, contractHashBytes, &lockParam)

	sink := common.NewZeroCopySink(nil)
	lockParam.Args.Serialization(sink)
	input := getCreateTxArgs(lockParam.ToChainID, contractHashBytes, lockParam.Fee, UNLOCK_NAME, sink.Bytes())
	_, err = native.NativeCall(utils.CrossChainManagerContractAddress, cross_chain_manager.CREATE_TX, input)
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
	contractHashBytes, err := utils.GetStorageVarBytes(native, GenBindProxyKey(lockContract, fromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] get bind contract hash with chainID:%d error:%s", fromChainId, err)
	}
	if !bytes.Equal(contractHashBytes, fromContractHashBytes) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] passed in proxy contractHash NOT equal stored contractHash with chainID:%d, expect:%s, got:%s", fromChainId, contractHashBytes, fromContractHashBytes)
	}
	// currently, only support ont
	if !bytes.Equal(args.AssetHash, ontContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] target asset hash, expect:%s, got:%s", hex.EncodeToString(ontContract[:]), hex.EncodeToString(args.AssetHash))
	}
	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnlock] parse from bytes to address error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}

	transferInput := getTransferInput(ont.State{lockContract, toAddress, args.Value})
	res, err := native.NativeCall(utils.OntContractAddress, ont.TRANSFER_NAME, transferInput)
	if !bytes.Equal(res.([]byte), utils.BYTE_TRUE) || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[OntUnLock] Transfer Ont, error:%s", err)
	}

	AddUnLockNotifications(native, lockContract, fromChainId, fromContractHashBytes, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}
