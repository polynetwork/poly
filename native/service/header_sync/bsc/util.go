package bsc

import (
	"encoding/json"
	"errors"
	"fmt"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	cstates "github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
)

// ParseValidators ...
func ParseValidators(validatorsBytes []byte) ([]ecommon.Address, error) {
	if len(validatorsBytes)%ecommon.AddressLength != 0 {
		return nil, errors.New("invalid validators bytes")
	}
	n := len(validatorsBytes) / ecommon.AddressLength
	result := make([]ecommon.Address, n)
	for i := 0; i < n; i++ {
		address := make([]byte, ecommon.AddressLength)
		copy(address, validatorsBytes[i*ecommon.AddressLength:(i+1)*ecommon.AddressLength])
		result[i] = ecommon.BytesToAddress(address)
	}
	return result, nil
}

func putGenesisBlockHeader(native *native.NativeService, blockHeader types.Header, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	storeBytes, _ := json.Marshal(&blockHeader)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), blockHeader.Hash().Bytes()),
		cstates.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockHeader.Number.Uint64())),
		cstates.GenRawStorageItem(blockHeader.Hash().Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(blockHeader.Number.Uint64())))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number.Uint64(), blockHeader.Hash().String())
	return nil
}

// IsHeaderExist ...
func IsHeaderExist(native *native.NativeService, hash []byte, chainID uint64) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return false, fmt.Errorf("IsHeaderExist, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return false, nil
	} else {
		return true, nil
	}
}

// GetHeaderByHash ...
func GetHeaderByHash(native *native.NativeService, hash []byte, chainID uint64) (*types.Header, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any header records")
	}
	storeBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize headerBytes from raw storage item err:%v", err)
	}
	var header types.Header
	if err := json.Unmarshal(storeBytes, &header); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return &header, nil
}

func putBlockHeader(native *native.NativeService, blockHeader types.Header, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	storeBytes, _ := json.Marshal(&blockHeader)
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), blockHeader.Hash().Bytes()),
		cstates.GenRawStorageItem(storeBytes))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number.Uint64(), blockHeader.Hash().String())
	return nil
}

func GetCurrentHeader(native *native.NativeService, chainID uint64) (*types.Header, error) {
	height, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, err
	}
	header, err := GetHeaderByHeight(native, height, chainID)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func GetCurrentHeaderHeight(native *native.NativeService, chainID uint64) (uint64, error) {
	heightStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return 0, fmt.Errorf("getPrevHeaderHeight error: %v", err)
	}
	if heightStore == nil {
		return 0, fmt.Errorf("getPrevHeaderHeight, heightStore is nil")
	}
	heightBytes, err := cstates.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		return 0, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return utils.GetBytesUint64(heightBytes), err
}

func GetHeaderByHeight(native *native.NativeService, height, chainID uint64) (*types.Header, error) {
	latestHeight, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, err
	}
	if height > latestHeight {
		return nil, fmt.Errorf("GetHeaderByHeight, height is too big")
	}
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHeight, can not find any header records")
	}
	hashBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return GetHeaderByHash(native, hashBytes, chainID)
}

func appendHeader2Main(native *native.NativeService, height uint64, txhash ecommon.Hash, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		cstates.GenRawStorageItem(txhash.Bytes()))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), cstates.GenRawStorageItem(utils.GetUint64Bytes(height)))
	scom.NotifyPutHeader(native, chainID, height, txhash.String())
	return nil
}

func UpdateConsensusPeer(native *native.NativeService, chainID uint64, header *types.Header) error {
	return nil
}

func RestructChain(native *native.NativeService, current, new *types.Header, chainID uint64) error {
	si, ti := current.Number.Uint64(), new.Number.Uint64()
	var err error
	if si > ti {
		current, err = GetHeaderByHeight(native, ti, chainID)
		if err != nil {
			return fmt.Errorf("RestructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
		si = ti
	}
	newHashs := make([]ecommon.Hash, 0)
	for ti > si {
		newHashs = append(newHashs, new.Hash())
		new, err = GetHeaderByHash(native, new.ParentHash.Bytes(), chainID)
		if err != nil {
			return fmt.Errorf("RestructChain GetHeaderByHash hash:%x error:%s", new.ParentHash.Bytes(), err)
		}
		ti--
	}
	for current.ParentHash != new.ParentHash {
		newHashs = append(newHashs, new.Hash())
		new, err = GetHeaderByHash(native, new.ParentHash.Bytes(), chainID)
		if err != nil {
			return fmt.Errorf("RestructChain GetHeaderByHash hash:%x  error:%s", new.ParentHash.Bytes(), err)
		}
		ti--
		si--
		current, err = GetHeaderByHeight(native, si, chainID)
		if err != nil {
			return fmt.Errorf("RestructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
	}
	newHashs = append(newHashs, new.Hash())
	for i := len(newHashs) - 1; i >= 0; i-- {
		appendHeader2Main(native, ti, newHashs[i], chainID)
		ti++
	}
	return nil
}
