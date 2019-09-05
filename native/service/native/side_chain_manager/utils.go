/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package side_chain_manager

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ontio/multi-chain/common"
	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/native/service/native"
	"github.com/ontio/multi-chain/native/service/native/utils"
)

func getRegisterSideChain(native *native.NativeService, chanid uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(chanid)
	if err != nil {
		return nil, fmt.Errorf("getRegisterSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sideChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(REGISTER_SIDE_CHAIN_REQUEST),
		chainidByte))
	if err != nil {
		return nil, fmt.Errorf("getRegisterSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := &SideChain{
		ChainId: math.MaxUint64,
	}
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getRegisterSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getRegisterSideChain, deserialize sideChain error: %v", err)
		}
	}
	return sideChain, nil
}

func putRegisterSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(sideChain.ChainId)
	if err != nil {
		return fmt.Errorf("putRegisterSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	err = sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putRegisterSideChain, sideChain.Serialization error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(REGISTER_SIDE_CHAIN_REQUEST), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func GetSideChain(native *native.NativeService, chainID uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainIDByte, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("getSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sideChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN),
		chainIDByte))
	if err != nil {
		return nil, fmt.Errorf("getSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := &SideChain{
		ChainId: math.MaxUint64,
	}
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getSideChain, deserialize sideChain error: %v", err)
		}
	}
	return sideChain, nil
}

func putSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(sideChain.ChainId)
	if err != nil {
		return fmt.Errorf("putSideChain, utils.GetUint32Bytes error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	err = sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putSideChain, sideChain.Serialization error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(SIDE_CHAIN), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getUpdateSideChain(native *native.NativeService, chanid uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(chanid)
	if err != nil {
		return nil, fmt.Errorf("getUpdateSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sideChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(UPDATE_SIDE_CHAIN_REQUEST),
		chainidByte))
	if err != nil {
		return nil, fmt.Errorf("getUpdateSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := new(SideChain)
	if sideChainStore != nil {
		sideChainBytes, err := cstates.GetValueFromRawStorageItem(sideChainStore)
		if err != nil {
			return nil, fmt.Errorf("getUpdateSideChain, deserialize from raw storage item err:%v", err)
		}
		if err := sideChain.Deserialization(common.NewZeroCopySource(sideChainBytes)); err != nil {
			return nil, fmt.Errorf("getUpdateSideChain, deserialize sideChain error: %v", err)
		}
	}
	return sideChain, nil
}

func putUpdateSideChain(native *native.NativeService, sideChain *SideChain) error {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(sideChain.ChainId)
	if err != nil {
		return fmt.Errorf("putUpdateSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sink := common.NewZeroCopySink(nil)
	err = sideChain.Serialization(sink)
	if err != nil {
		return fmt.Errorf("putUpdateSideChain, sideChain.Serialization error: %v", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(UPDATE_SIDE_CHAIN_REQUEST), chainidByte),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getAssetMapRequest(native *native.NativeService, assetName string) (*AssetMappingParam, error) {
	contract := utils.SideChainManagerContractAddress
	assetMapRequestStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(ASSET_MAP_REQUEST),
		[]byte(assetName)))
	if err != nil {
		return nil, fmt.Errorf("getAssetMapRequest, get assetMapRequestStore error: %v", err)
	}
	assetMappingParam := new(AssetMappingParam)
	if assetMapRequestStore != nil {
		assetMapRequestBytes, err := cstates.GetValueFromRawStorageItem(assetMapRequestStore)
		if err != nil {
			return nil, fmt.Errorf("getAssetMapRequest, deserialize from raw storage item err:%v", err)
		}
		if err := assetMappingParam.Deserialization(common.NewZeroCopySource(assetMapRequestBytes)); err != nil {
			return nil, fmt.Errorf("getAssetMapRequest, deserialize sideChain error: %v", err)
		}
	}
	return assetMappingParam, nil
}

func putAssetMapRequest(native *native.NativeService, assetMappingParam *AssetMappingParam) error {
	contract := utils.SideChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	assetMappingParam.Serialization(sink)
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(ASSET_MAP_REQUEST), []byte(assetMappingParam.AssetName)),
		cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func putAssetMap(native *native.NativeService, assetMap *AssetMap) error {
	contract := utils.SideChainManagerContractAddress
	sink := common.NewZeroCopySink(nil)
	assetMap.Serialization(sink)
	for _, v := range assetMap.AssetMap {
		prefix := strconv.Itoa(int(v.ChainId)) + v.ContractAddress
		native.CacheDB.Put(utils.ConcatKey(contract, []byte(ASSET_MAP), []byte(prefix)),
			cstates.GenRawStorageItem(sink.Bytes()))
	}
	return nil
}

func GetDestAsset(native *native.NativeService, fromChainid, toChainid uint64, contractAddress string) (*Asset, error) {
	contract := utils.SideChainManagerContractAddress
	prefix := strconv.Itoa(int(fromChainid)) + contractAddress
	assetMapStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(ASSET_MAP), []byte(prefix)))
	if err != nil {
		return nil, fmt.Errorf("getAssetMap,get assetMapStore error: %v", err)
	}
	if assetMapStore == nil {
		return nil, fmt.Errorf("getAssetMap, can't find any record with from chainid %d and contract address %s", fromChainid, contractAddress)
	}
	assetMapBytes, err := cstates.GetValueFromRawStorageItem(assetMapStore)
	if err != nil {
		return nil, fmt.Errorf("getAssetMap, deserialize from raw storage item err:%v", err)
	}
	assetMap := new(AssetMap)
	if err := assetMap.Deserialization(common.NewZeroCopySource(assetMapBytes)); err != nil {
		return nil, fmt.Errorf("getAssetMap, deserialize assetMap error: %v", err)
	}
	return assetMap.AssetMap[toChainid], nil
}
