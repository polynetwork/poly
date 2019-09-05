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
	"github.com/ontio/multi-chain/native/service/native"
	"github.com/ontio/multi-chain/native/service/native/utils"
	"math"
)

const (
	//function name
	REGISTER_SIDE_CHAIN         = "registerSideChain"
	APPROVE_REGISTER_SIDE_CHAIN = "approveRegisterSideChain"
	UPDATE_SIDE_CHAIN           = "updateSideChain"
	APPROVE_UPDATE_SIDE_CHAIN   = "approveUpdateSideChain"
	REMOVE_SIDE_CHAIN           = "removeSideChain"
	ASSET_MAPPING               = "assetMapping"
	APPROVE_ASSET_MAPPING       = "approveAssetMapping"

	//key prefix
	REGISTER_SIDE_CHAIN_REQUEST = "registerSideChainRequest"
	UPDATE_SIDE_CHAIN_REQUEST   = "updateSideChainRequest"
	SIDE_CHAIN                  = "sideChain"
	ASSET_MAP                   = "assetMap"
	ASSET_MAP_REQUEST           = "assetMapRequest"

	//constant
	ONG = 500000000000
)

//Init contract address
func InitSideChainManager() {
	native.Contracts[utils.SideChainManagerContractAddress] = RegisterSideChainManagerContract
}

//Register methods of governance contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_REGISTER_SIDE_CHAIN, ApproveRegisterSideChain)
	native.Register(UPDATE_SIDE_CHAIN, UpdateSideChain)
	native.Register(APPROVE_UPDATE_SIDE_CHAIN, ApproveUpdateSideChain)
	native.Register(REMOVE_SIDE_CHAIN, RemoveSideChain)

	native.Register(ASSET_MAPPING, AssetMapping)
	native.Register(APPROVE_ASSET_MAPPING, ApproveAssetMapping)
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, common.AddressFromBase58 error: %v", err)
	}
	registerSideChain, err := getRegisterSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, getRegisterSideChain error: %v", err)
	}
	if registerSideChain.ChainId != math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, chainid already requested")
	}
	sideChain, err := GetSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, getSideChain error: %v", err)
	}
	if sideChain.ChainId != math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, chainid already registered")
	}
	sideChain = &SideChain{
		ChainId:      params.ChainId,
		Name:         params.Name,
		BlocksToWait: params.BlocksToWait,
	}
	err = putRegisterSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, putRegisterSideChain error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, address, utils.GovernanceContractAddress, ONG)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOng, ong transfer error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func ApproveRegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, contract params deserialize error: %v", err)
	}

	registerSideChain, err := getRegisterSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, getRegisterSideChain error: %v", err)
	}
	if registerSideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, chainid is not requested")
	}
	err = putSideChain(native, registerSideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, putSideChain error: %v", err)
	}
	chainidByte, err := utils.GetUint64Bytes(params.Chainid)
	if err != nil {
		return nil, fmt.Errorf("ApproveRegisterSideChain, utils.GetUint64Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(REGISTER_SIDE_CHAIN_REQUEST), chainidByte))

	return utils.BYTE_TRUE, nil
}

func UpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, contract params deserialize error: %v", err)
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, common.AddressFromBase58 error: %v", err)
	}
	sideChain, err := GetSideChain(native, params.ChainId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, getSideChain error: %v", err)
	}
	if sideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, chainid is not registered")
	}
	updateSideChain := &SideChain{
		ChainId:      params.ChainId,
		Name:         params.Name,
		BlocksToWait: params.BlocksToWait,
	}
	err = putUpdateSideChain(native, updateSideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, putUpdateSideChain error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, address, utils.GovernanceContractAddress, ONG)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOng, ong transfer error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func ApproveUpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, contract params deserialize error: %v", err)
	}

	sideChain, err := getUpdateSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, chainid is not requested update")
	}
	err = putSideChain(native, sideChain)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, putSideChain error: %v", err)
	}
	chainidByte, err := utils.GetUint64Bytes(params.Chainid)
	if err != nil {
		return nil, fmt.Errorf("ApproveUpdateSideChain, utils.GetUint64Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(UPDATE_SIDE_CHAIN_REQUEST), chainidByte))

	return utils.BYTE_TRUE, nil
}

func RemoveSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, contract params deserialize error: %v", err)
	}

	sideChain, err := GetSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, chainid is not registered")
	}
	chainidByte, err := utils.GetUint64Bytes(params.Chainid)
	if err != nil {
		return nil, fmt.Errorf("RemoveSideChain, utils.GetUint64Bytes error: %v", err)
	}
	native.CacheDB.Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(SIDE_CHAIN), chainidByte))

	return utils.BYTE_TRUE, nil
}

func AssetMapping(native *native.NativeService) ([]byte, error) {
	params := new(AssetMappingParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, contract params deserialize error: %v", err)
	}
	assetMapRequest, err := getAssetMapRequest(native, params.AssetName)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, getAssetMapRequest error: %v", err)
	}
	if assetMapRequest.AssetName != "" {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, asset name is already used")
	}
	err = putAssetMapRequest(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, putAssetMapRequest error: %v", err)
	}
	address, err := common.AddressFromBase58(params.Address)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, common.AddressFromBase58 error: %v", err)
	}

	//ong transfer
	err = appCallTransferOng(native, address, utils.GovernanceContractAddress, ONG)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("appCallTransferOng, ong transfer error: %v", err)
	}
	return utils.BYTE_TRUE, nil
}

func ApproveAssetMapping(native *native.NativeService) ([]byte, error) {
	params := new(ApproveAssetMappingParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, contract params deserialize error: %v", err)
	}
	assetMapRequest, err := getAssetMapRequest(native, params.AssetName)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, getAssetMapRequest error: %v", err)
	}
	if assetMapRequest.AssetName == "" {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, asset name is not requested")
	}
	assetMap := make(map[uint64]*Asset)
	for _, v := range assetMapRequest.AssetList {
		assetMap[v.ChainId] = v
	}
	value := &AssetMap{
		AssetMap: assetMap,
	}
	sink := common.NewZeroCopySink(nil)
	err = value.Serialization(sink)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, value.Serialization error: %v", err)
	}
	err = putAssetMap(native, value)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, putAssetMap error: %v", err)
	}

	native.CacheDB.Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(ASSET_MAP_REQUEST), []byte(params.AssetName)))
	return utils.BYTE_TRUE, nil
}
