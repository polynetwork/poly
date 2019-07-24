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
	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/ont"
	"strconv"

	cstates "github.com/ontio/multi-chain/core/states"
	"github.com/ontio/multi-chain/smartcontract/service/native"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

func appCallTransferOng(native *native.NativeService, from common.Address, to common.Address, amount uint64) error {
	err := appCallTransfer(native, utils.OngContractAddress, from, to, amount)
	if err != nil {
		return fmt.Errorf("appCallTransferOng, appCallTransfer error: %v", err)
	}
	return nil
}

func appCallTransfer(native *native.NativeService, contract common.Address, from common.Address, to common.Address, amount uint64) error {
	var sts []ont.State
	sts = append(sts, ont.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := ont.Transfers{
		States: sts,
	}
	sink := common.NewZeroCopySink(nil)
	transfers.Serialization(sink)

	if _, err := native.NativeCall(contract, "transfer", sink.Bytes()); err != nil {
		return fmt.Errorf("appCallTransfer, appCall error: %v", err)
	}
	return nil
}

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
	sideChain := new(SideChain)
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
	chainidByte, err := utils.GetUint64Bytes(sideChain.Chainid)
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

func getSideChain(native *native.NativeService, chanid uint64) (*SideChain, error) {
	contract := utils.SideChainManagerContractAddress
	chainidByte, err := utils.GetUint64Bytes(chanid)
	if err != nil {
		return nil, fmt.Errorf("getSideChain, utils.GetUint64Bytes error: %v", err)
	}
	sideChainStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(SIDE_CHAIN),
		chainidByte))
	if err != nil {
		return nil, fmt.Errorf("getSideChain,get registerSideChainRequestStore error: %v", err)
	}
	sideChain := new(SideChain)
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
	chainidByte, err := utils.GetUint64Bytes(sideChain.Chainid)
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
	chainidByte, err := utils.GetUint64Bytes(sideChain.Chainid)
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
		prefix := strconv.Itoa(int(v.Chainid)) + v.ContractAddress
		native.CacheDB.Put(utils.ConcatKey(contract, []byte(ASSET_MAP), []byte(prefix)),
			cstates.GenRawStorageItem(sink.Bytes()))
	}
	return nil
}

func getAssetContractAddress(native *native.NativeService, fromChainid, toChainid uint64, contractAddress string) (string, error) {
	contract := utils.SideChainManagerContractAddress
	prefix := strconv.Itoa(int(fromChainid)) + contractAddress
	assetMapStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(ASSET_MAP), []byte(prefix)))
	if err != nil {
		return "", fmt.Errorf("getAssetMap,get assetMapStore error: %v", err)
	}
	if assetMapStore == nil {
		return "", fmt.Errorf("getAssetMap, can't find any record")
	}
	assetMapBytes, err := cstates.GetValueFromRawStorageItem(assetMapStore)
	if err != nil {
		return "", fmt.Errorf("getAssetMap, deserialize from raw storage item err:%v", err)
	}
	assetMap := new(AssetMap)
	if err := assetMap.Deserialization(common.NewZeroCopySource(assetMapBytes)); err != nil {
		return "", fmt.Errorf("getAssetMap, deserialize assetMap error: %v", err)
	}
	return assetMap.AssetMap[toChainid].ContractAddress, nil
}
