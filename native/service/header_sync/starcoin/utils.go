/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package starcoin

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/polynetwork/poly/core/states"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/service/utils"
	stc "github.com/starcoinorg/starcoin-go/client"
	"github.com/starcoinorg/starcoin-go/types"
	"time"
)

const allowedFutureBlockTime = 30 * time.Second

func putBlockHeader(native *native.NativeService, blockHeader types.BlockHeader, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	storeBytes, err := blockHeader.BcsSerialize()
	if err != nil {
		return errors.WithStack(err)
	}
	headerHash, err := blockHeader.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), *headerHash),
		states.GenRawStorageItem(storeBytes))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number, stc.BytesToHexString(*headerHash))
	return nil
}

func putGenesisBlockHeader(native *native.NativeService, blockHeader types.BlockHeader, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress

	storeBytes, err := blockHeader.BcsSerialize()
	if err != nil {
		return errors.WithStack(err)
	}

	headerHash, err := blockHeader.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.GENESIS_HEADER), utils.GetUint64Bytes(chainID)),
		states.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), *headerHash),
		states.GenRawStorageItem(storeBytes))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(blockHeader.Number)),
		states.GenRawStorageItem(*headerHash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), states.GenRawStorageItem(utils.GetUint64Bytes(blockHeader.Number)))
	scom.NotifyPutHeader(native, chainID, blockHeader.Number, stc.BytesToHexString(*headerHash))
	return nil
}

func GetCurrentHeader(native *native.NativeService, chainID uint64) (*types.BlockHeader, error) {
	height, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	header, err := GetHeaderByHeight(native, height, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return header, nil
}

func GetCurrentHeaderHeight(native *native.NativeService, chainID uint64) (uint64, error) {
	heightStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.CURRENT_HEADER_HEIGHT), utils.GetUint64Bytes(chainID)))
	if err != nil {
		return 0, errors.Errorf("getPrevHeaderHeight error: %v", err)
	}
	if heightStore == nil {
		return 0, errors.Errorf("getPrevHeaderHeight, heightStore is nil")
	}
	heightBytes, err := states.GetValueFromRawStorageItem(heightStore)
	if err != nil {
		return 0, errors.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return utils.GetBytesUint64(heightBytes), err
}

func GetHeaderByHeight(native *native.NativeService, height, chainID uint64) (*types.BlockHeader, error) {
	latestHeight, err := GetCurrentHeaderHeight(native, chainID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if height > latestHeight {
		return nil, errors.Errorf("GetHeaderByHeight, height is too big")
	}
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)))
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, errors.Errorf("GetHeaderByHeight, can not find any header records")
	}
	hashBytes, err := states.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	return GetHeaderByHash(native, hashBytes, chainID)
}

func IsHeaderExist(native *native.NativeService, hash []byte, chainID uint64) (bool, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return false, errors.Errorf("IsHeaderExist, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func GetHeaderByHash(native *native.NativeService, hash []byte, chainID uint64) (*types.BlockHeader, error) {
	headerStore, err := native.GetCacheDB().Get(utils.ConcatKey(utils.HeaderSyncContractAddress,
		[]byte(scom.HEADER_INDEX), utils.GetUint64Bytes(chainID), hash))
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHash, get blockHashStore error: %v", err)
	}
	if headerStore == nil {
		return nil, errors.Errorf("GetHeaderByHash, can not find any header records")
	}
	storeBytes, err := states.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHash, deserialize headerBytes from raw storage item err:%v", err)
	}
	header, err := types.BcsDeserializeBlockHeader(storeBytes)
	if err != nil {
		return nil, errors.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return &header, nil
}

func appendHeader2Main(native *native.NativeService, height uint64, txhash types.HashValue, chainID uint64) error {
	contract := utils.HeaderSyncContractAddress
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.MAIN_CHAIN), utils.GetUint64Bytes(chainID), utils.GetUint64Bytes(height)),
		states.GenRawStorageItem(txhash))
	native.GetCacheDB().Put(utils.ConcatKey(contract, []byte(scom.CURRENT_HEADER_HEIGHT),
		utils.GetUint64Bytes(chainID)), states.GenRawStorageItem(utils.GetUint64Bytes(height)))
	scom.NotifyPutHeader(native, chainID, height, stc.BytesToHexString(txhash))
	return nil
}

func ReStructChain(native *native.NativeService, current, new *types.BlockHeader, chainID uint64) error {
	si, ti := current.Number, new.Number
	var err error
	if si > ti {
		current, err = GetHeaderByHeight(native, ti, chainID)
		if err != nil {
			return errors.Errorf("ReStructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
		si = ti
	}
	newHashes := make([]types.HashValue, 0)
	for ti > si {
		newHash, err := new.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}
		newHashes = append(newHashes, *newHash)
		new, err = GetHeaderByHash(native, new.ParentHash, chainID)
		if err != nil {
			return errors.Errorf("ReStructChain GetHeaderByHash hash:%x error:%s", new.ParentHash, err)
		}
		ti--
	}
	for !bytes.Equal(current.ParentHash, new.ParentHash) {
		newHash, err := new.GetHash()
		if err != nil {
			return errors.WithStack(err)
		}

		newHashes = append(newHashes, *newHash)
		new, err = GetHeaderByHash(native, new.ParentHash, chainID)
		if err != nil {
			return errors.Errorf("ReStructChain GetHeaderByHash hash:%x  error:%s", new.ParentHash, err)
		}
		ti--
		si--
		current, err = GetHeaderByHeight(native, si, chainID)
		if err != nil {
			return errors.Errorf("ReStructChain GetHeaderByHeight height:%d error:%s", ti, err)
		}
	}
	newHash, err := new.GetHash()
	if err != nil {
		return errors.WithStack(err)
	}
	newHashes = append(newHashes, *newHash)
	for i := len(newHashes) - 1; i >= 0; i-- {
		appendHeader2Main(native, ti, newHashes[i], chainID)
		ti++
	}
	return nil
}
