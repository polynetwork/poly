package lock_proxy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/common/config"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/event"
	"github.com/ontio/multi-chain/native/service/cross_chain_manager"
	"github.com/ontio/multi-chain/native/service/ont"
	"github.com/ontio/multi-chain/native/service/utils"
	"io"
	"math/big"
)

func RegisterLockContract(native *native.NativeService) {
	native.Register(LOCK_NAME, Lock)
	native.Register(UNLOCK_NAME, Unlock)
	native.Register(BIND_PROXY_NAME, BindProxyHash)
	native.Register(BIND_ASSET_NAME, BindAssetHash)
	native.Register(GET_PROXY_HASH_NAME, GetProxyHash)
	native.Register(GET_ASSET_HASH_NAME, GetAssetHash)
	native.Register(GET_CROSSED_AMOUNT_NAME, GetCrossedAmount)
	native.Register(GET_CROSSED_LIMIT_NAME, GetCrossedLimit)
}

func BindProxyHash(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	var bindParam BindProxyParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] Deserialize BindProxyParam error:%s", io.ErrUnexpectedEOF)
	}
	// check witness
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] get operator error:%s", err)
	}
	if err = utils.ValidateOwner(native, operatorAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindProxyHash] checkWitness operator error:%s", err)
	}
	native.GetCacheDB().Put(GenBindProxyKey(contract, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetHash).ToArray())
	if config.DefConfig.Common.EnableEventLog {
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: contract,
				States:          []interface{}{BIND_PROXY_NAME, bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetHash)},
			})
	}

	return utils.BYTE_TRUE, nil
}

func BindAssetHash(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	var bindParam BindAssetParam
	if err := bindParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] Deserialization BindParam error:%s", io.ErrUnexpectedEOF)
	}
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] get operator error:%s", err)
	}
	// check witness
	if err = utils.ValidateOwner(native, operatorAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] checkWitness operator error:%s", err)
	}
	// store the target asset hash
	native.GetCacheDB().Put(GenBindAssetHashKey(contract, bindParam.SourceAssetHash, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.TargetAssetHash).ToArray())

	// make sure the new limit is greater than the stored limit
	limitKey := GenCrossedLimitKey(contract, bindParam.SourceAssetHash, bindParam.TargetChainId)
	limit, err := getAmount(native, limitKey)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] getCrossedLimit error:%s", err)
	}
	if bindParam.Limit.Cmp(limit) != 1 {
		return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] new Limit:%s should be greater than stored Limit:%s", bindParam.Limit.String(), limit.String())
	}
	// if the source asset hash is the target chain asset, increase the crossedAmount value by the limit increment
	if bindParam.IsTargetChainAsset {
		increment := big.NewInt(0).Sub(bindParam.Limit, limit)
		crossedAmountKey := GenCrossedAmountKey(contract, bindParam.SourceAssetHash, bindParam.TargetChainId)
		crossedAmount, err := getAmount(native, crossedAmountKey)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] getCrossedAmount error:%s", err)
		}
		newCrossedAmount := big.NewInt(0).Add(crossedAmount, increment)
		if newCrossedAmount.Cmp(crossedAmount) != 1 {
			return utils.BYTE_FALSE, fmt.Errorf("[BindAssetHash] new crossedAmount:%s is not greater than stored crossed amount:%s", newCrossedAmount.String(), crossedAmount.String())
		}
		native.GetCacheDB().Put(crossedAmountKey, utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())
	}
	// update the new limit
	native.GetCacheDB().Put(GenCrossedLimitKey(contract, bindParam.SourceAssetHash, bindParam.TargetChainId), utils.GenVarBytesStorageItem(bindParam.Limit.Bytes()).ToArray())

	if config.DefConfig.Common.EnableEventLog {
		native.AddNotify(
			&event.NotifyEventInfo{
				ContractAddress: contract,
				States:          []interface{}{BIND_PROXY_NAME, hex.EncodeToString(bindParam.SourceAssetHash[:]), bindParam.TargetChainId, hex.EncodeToString(bindParam.TargetAssetHash), bindParam.Limit.String()},
			})
	}

	return utils.BYTE_TRUE, nil
}

func Lock(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	ontContract := utils.OntContractAddress
	ongContract := utils.OngContractAddress
	source := common.NewZeroCopySource(native.GetInput())

	var lockParam LockParam
	if err := lockParam.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] contract params deserialization error:%v", err)
	}

	if lockParam.Value == 0 {
		return utils.BYTE_FALSE, nil
	}

	// currently, only support ont and ong lock operation
	if !bytes.Equal(lockParam.SourceAssetHash[:], ontContract[:]) && !bytes.Equal(lockParam.SourceAssetHash[:], ongContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] only support ont/ong lock, expect:%s or %s, but got:%s", hex.EncodeToString(ontContract[:]), hex.EncodeToString(ongContract[:]), hex.EncodeToString(lockParam.SourceAssetHash[:]))
	}

	state := ont.State{
		From:  lockParam.FromAddress,
		To:    contract,
		Value: lockParam.Value,
	}
	transferInput := getTransferInput(state)
	if res, err := native.NativeCall(lockParam.SourceAssetHash, ont.TRANSFER_NAME, transferInput); !bytes.Equal(res.([]byte), utils.BYTE_TRUE) || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] NativeCall contract:%s 'transfer(%s, %s, %d)' error:%s", hex.EncodeToString(lockParam.SourceAssetHash[:]), lockParam.FromAddress.ToBase58(), hex.EncodeToString(contract[:]), lockParam.Value, err)
	}

	// make sure new crossed amount is strictly greater than old crossed amount and no less than the limit
	crossedAmount, err := getAmount(native, GenCrossedAmountKey(contract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] getCrossedAmount error:%s", err)
	}
	limit, err := getAmount(native, GenCrossedLimitKey(contract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] getCrossedLimit error:%s", err)
	}
	newCrossedAmount := big.NewInt(0).Add(crossedAmount, big.NewInt(0).SetUint64(lockParam.Value))
	if newCrossedAmount.Cmp(crossedAmount) != 1 || newCrossedAmount.Cmp(limit) == 1 {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] Is new crossedAmount GREATER than old crossedAmount?:%t, Is new crossedAmount SMALLER than limit?:%t", newCrossedAmount.Cmp(crossedAmount) == 1, newCrossedAmount.Cmp(limit) != 1)
	}
	// increase the new crossed amount by Value
	native.GetCacheDB().Put(GenCrossedAmountKey(contract, lockParam.SourceAssetHash, lockParam.ToChainID), utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())

	// get target chain proxy hash from storage
	targetProxyHashBs, err := utils.GetStorageVarBytes(native, GenBindProxyKey(contract, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind proxy contract hash with chainID:%d error:%s", lockParam.ToChainID, err)
	}
	if len(targetProxyHashBs) == 0 {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind proxy contract hash with chainID:%d contractHash empty", lockParam.ToChainID)
	}

	// get target asset hash from storage
	targetAssetHash, err := utils.GetStorageVarBytes(native, GenBindAssetHashKey(contract, lockParam.SourceAssetHash, lockParam.ToChainID))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] get bind asset contract hash:%s with chainID:%d error:%s", hex.EncodeToString(lockParam.SourceAssetHash[:]), lockParam.ToChainID, err)
	}

	args := Args{
		TargetAssetHash: targetAssetHash,
		ToAddress:       lockParam.ToAddress,
		Value:           lockParam.Value,
	}
	sink := common.NewZeroCopySink(nil)
	args.Serialization(sink)

	input := getCreateTxArgs(lockParam.ToChainID, targetProxyHashBs, UNLOCK_NAME, sink.Bytes())
	if _, err = native.NativeCall(utils.CrossChainManagerContractAddress, cross_chain_manager.CREATE_TX, input); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Lock] NativeCall %s createCrossChainTx 'createTx', error:%s", hex.EncodeToString(utils.CrossChainManagerContractAddress[:]), err)
	}

	AddLockNotifications(native, contract, targetProxyHashBs, targetAssetHash, &lockParam)
	return utils.BYTE_TRUE, nil
}

func Unlock(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	// this method cannot be invoked by anybody except CrossChainManager
	if err := utils.ValidateOwner(native, utils.CrossChainManagerContractAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] can ONLY be invoked by CrossChainManagerContractAddress:%s Contract, checkwitness failed!", hex.EncodeToString(utils.CrossChainManagerContractAddress[:]))
	}
	ontContract := utils.OntContractAddress
	ongContract := utils.OngContractAddress

	var unlockParam UnlockParam

	if err := unlockParam.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] contract params deserialization error:%s", err)
	}

	var args Args
	if err := args.Deserialization(common.NewZeroCopySource(unlockParam.ArgsBs)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] deserialize args error:%s", err)
	}
	// only recognize the params from proxy contract already bound with chainId in current proxy contract
	proxyContractHash, err := utils.GetStorageVarBytes(native, GenBindProxyKey(contract, unlockParam.FromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] get bind proxy contract hash with chainID:%d error:%s", unlockParam.FromChainId, err)
	}
	if !bytes.Equal(proxyContractHash, unlockParam.FromContractHashBs) {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] passed in proxy contractHash NOT equal stored contractHash with chainID:%d, expect:%s, got:%s", unlockParam.FromChainId, hex.EncodeToString(proxyContractHash), hex.EncodeToString(unlockParam.FromContractHashBs))
	}
	// currently, only support ont and ong unlock operation
	if !bytes.Equal(args.TargetAssetHash, ontContract[:]) && !bytes.Equal(args.TargetAssetHash, ongContract[:]) {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] target asset hash, Is ONT contract?:%t, Is ONG contract?:%t, Args.TargetAssetHash:%s", bytes.Equal(args.TargetAssetHash, ontContract[:]), bytes.Equal(args.ToAddress, ongContract[:]), hex.EncodeToString(args.TargetAssetHash))
	}

	assetAddress, err := common.AddressParseFromBytes(args.TargetAssetHash)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] parse from Args.TargetAssetHash to contract address format error:%s", err)
	}
	toAddress, err := common.AddressParseFromBytes(args.ToAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] parse from Args.ToAddress to ORChain address format error:%s", err)
	}
	if args.Value == 0 {
		return utils.BYTE_TRUE, nil
	}
	// unlock ont or ong from current proxy contract into toAddress
	transferInput := getTransferInput(ont.State{contract, toAddress, args.Value})
	if res, err := native.NativeCall(assetAddress, ont.TRANSFER_NAME, transferInput); !bytes.Equal(res.([]byte), utils.BYTE_TRUE) || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] NativeCall contract:%s 'transfer(%s, %s, %d)' error:%s", hex.EncodeToString(assetAddress[:]), hex.EncodeToString(contract[:]), toAddress.ToBase58(), args.Value, err)
	}

	// make sure new crossed amount is strictly less than old crossed amount and no less than the limit
	crossedAmount, err := getAmount(native, GenCrossedAmountKey(contract, assetAddress, unlockParam.FromChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] getCrossedAmount error:%s", err)
	}
	newCrossedAmount := big.NewInt(0).Sub(crossedAmount, big.NewInt(0).SetUint64(args.Value))
	if newCrossedAmount.Cmp(crossedAmount) != -1 {
		return utils.BYTE_FALSE, fmt.Errorf("[Unlock] new crossedAmount:%s should be less than old crossedAmount:%s", newCrossedAmount.String(), crossedAmount.String())
	}
	// decrease the new crossed amount by Value
	native.GetCacheDB().Put(GenCrossedAmountKey(contract, assetAddress, unlockParam.FromChainId), utils.GenVarBytesStorageItem(newCrossedAmount.Bytes()).ToArray())

	AddUnLockNotifications(native, contract, unlockParam.FromChainId, unlockParam.FromContractHashBs, assetAddress, toAddress, args.Value)

	return utils.BYTE_TRUE, nil
}

func GetProxyHash(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	toChainId, eof := common.NewZeroCopySource(native.GetInput()).NextVarUint()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetProxyHash] input DecodeVarUint toChainId error:%s", io.ErrUnexpectedEOF)
	}
	proxyHash, err := utils.GetStorageVarBytes(native, GenBindProxyKey(contract, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetProxyHash] get proxy hash with toChainId:%d error:%s", toChainId, err)
	}
	return proxyHash, nil
}

func GetAssetHash(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	sourceAssetAddress, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] input NextAddress sourceAssetAddress error:%s", io.ErrUnexpectedEOF)
	}
	toChainId, eof := source.NextVarUint()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] input NextVarUint toChainId error:%s", io.ErrUnexpectedEOF)
	}
	toAssetHash, err := utils.GetStorageVarBytes(native, GenBindAssetHashKey(contract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetAssetHash] get asset hash with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, hex.EncodeToString(sourceAssetAddress[:]), err)
	}
	return toAssetHash, nil
}

func GetCrossedAmount(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	sourceAssetAddress, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] input NextAddress sourceAssetAddress error:%s", io.ErrUnexpectedEOF)
	}
	toChainId, eof := source.NextVarUint()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] input NextVarUint toChainId error:%s", io.ErrUnexpectedEOF)
	}

	crossedAmountBs, err := utils.GetStorageVarBytes(native, GenCrossedAmountKey(contract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedAmount] get crossed amount in big.Int bytes format with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, sourceAssetAddress.ToHexString(), err)
	}
	return crossedAmountBs, nil
}

func GetCrossedLimit(native *native.NativeService) ([]byte, error) {
	contract := utils.LockProxyContractAddress
	source := common.NewZeroCopySource(native.GetInput())
	sourceAssetAddress, eof := source.NextAddress()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] input NextAddress sourceAssetAddress error:%s", io.ErrUnexpectedEOF)
	}
	toChainId, eof := source.NextVarUint()
	if eof {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] input NextVarUint toChainId error:%s", io.ErrUnexpectedEOF)
	}

	crossedLimitBs, err := utils.GetStorageVarBytes(native, GenCrossedLimitKey(contract, sourceAssetAddress, toChainId))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("[GetCrossedLimit] get crossed limit in big.Int bytes format with toChainId:%d for sourceAssetAddress:%s error:%s", toChainId, sourceAssetAddress.ToHexString(), err)
	}
	return crossedLimitBs, nil
}
