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

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/core/genesis"
	"github.com/ontio/multi-chain/core/types"
	"github.com/ontio/multi-chain/native"
	"github.com/ontio/multi-chain/native/service/utils"
)

const (
	//function name
	REGISTER_SIDE_CHAIN                  = "registerSideChain"
	APPROVE_REGISTER_SIDE_CHAIN          = "approveRegisterSideChain"
	UPDATE_SIDE_CHAIN                    = "updateSideChain"
	APPROVE_UPDATE_SIDE_CHAIN            = "approveUpdateSideChain"
	REMOVE_SIDE_CHAIN                    = "removeSideChain"
	CROSS_CHAIN_CONTRACT_MAPPING         = "crossChainContractMapping"
	APPROVE_CROSS_CHAIN_CONTRACT_MAPPING = "approveCrossChainContractMapping"

	//key prefix
	REGISTER_SIDE_CHAIN_REQUEST      = "registerSideChainRequest"
	UPDATE_SIDE_CHAIN_REQUEST        = "updateSideChainRequest"
	SIDE_CHAIN                       = "sideChain"
	CROSS_CHAIN_CONTRACT_MAP         = "crossChainContractMap"
	CROSS_CHAIN_CONTRACT_MAP_REQUEST = "crossChainContractMapRequest"
)

//Register methods of governance contract
func RegisterSideChainManagerContract(native *native.NativeService) {
	native.Register(REGISTER_SIDE_CHAIN, RegisterSideChain)
	native.Register(APPROVE_REGISTER_SIDE_CHAIN, ApproveRegisterSideChain)
	native.Register(UPDATE_SIDE_CHAIN, UpdateSideChain)
	native.Register(APPROVE_UPDATE_SIDE_CHAIN, ApproveUpdateSideChain)
	native.Register(REMOVE_SIDE_CHAIN, RemoveSideChain)

	native.Register(CROSS_CHAIN_CONTRACT_MAPPING, CrossChainContractMapping)
	native.Register(APPROVE_CROSS_CHAIN_CONTRACT_MAPPING, ApproveCrossChainContractMapping)
}

func RegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RegisterSideChain, contract params deserialize error: %v", err)
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

	return utils.BYTE_TRUE, nil
}

func ApproveRegisterSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveRegisterSideChain, checkWitness error: %v", err)
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
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(REGISTER_SIDE_CHAIN_REQUEST), utils.GetUint64Bytes(params.Chainid)))

	return utils.BYTE_TRUE, nil
}

func UpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(RegisterSideChainParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("UpdateSideChain, contract params deserialize error: %v", err)
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

	return utils.BYTE_TRUE, nil
}

func ApproveUpdateSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveUpdateSideChain, checkWitness error: %v", err)
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
	chainidByte := utils.GetUint64Bytes(params.Chainid)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(UPDATE_SIDE_CHAIN_REQUEST), chainidByte))

	return utils.BYTE_TRUE, nil
}

func RemoveSideChain(native *native.NativeService) ([]byte, error) {
	params := new(ChainidParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, checkWitness error: %v", err)
	}

	sideChain, err := GetSideChain(native, params.Chainid)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, getUpdateSideChain error: %v", err)
	}
	if sideChain.ChainId == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("RemoveSideChain, chainid is not registered")
	}
	chainidByte := utils.GetUint64Bytes(params.Chainid)
	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(SIDE_CHAIN), chainidByte))

	return utils.BYTE_TRUE, nil
}

func CrossChainContractMapping(native *native.NativeService) ([]byte, error) {
	params := new(CrossChainContractMappingParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, contract params deserialize error: %v", err)
	}
	crossChainContractMapRequest, err := getCrossChainContractMapRequest(native, params.CrossChainContractName)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, getAssetMapRequest error: %v", err)
	}
	if crossChainContractMapRequest.CrossChainContractName != "" {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, asset name is already used")
	}
	err = putCrossChainContractMapRequest(native, params)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, putAssetMapRequest error: %v", err)
	}

	return utils.BYTE_TRUE, nil
}

func ApproveCrossChainContractMapping(native *native.NativeService) ([]byte, error) {
	params := new(ApproveCrossChainContractMappingParam)
	if err := params.Deserialization(common.NewZeroCopySource(native.GetInput())); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, contract params deserialize error: %v", err)
	}

	// get operator from database
	operatorAddress, err := types.AddressFromBookkeepers(genesis.GenesisBookkeepers)
	if err != nil {
		return utils.BYTE_FALSE, err
	}

	//check witness
	err = utils.ValidateOwner(native, operatorAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, checkWitness error: %v", err)
	}

	assetMapRequest, err := getCrossChainContractMapRequest(native, params.CrossChainContractName)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, getAssetMapRequest error: %v", err)
	}
	if assetMapRequest.CrossChainContractName == "" {
		return utils.BYTE_FALSE, fmt.Errorf("ApproveAssetMapping, asset name is not requested")
	}
	assetMap := make(map[uint64]*CrossChainContract)
	for _, v := range assetMapRequest.CrossChainContractList {
		assetMap[v.ChainId] = v
	}
	value := &CrossChainContractMap{
		CrossChainContractMap: assetMap,
	}
	sink := common.NewZeroCopySink(nil)
	err = value.Serialization(sink)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, value.Serialization error: %v", err)
	}
	err = putCrossChainContractMap(native, value)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("AssetMapping, putAssetMap error: %v", err)
	}

	native.GetCacheDB().Delete(utils.ConcatKey(utils.SideChainManagerContractAddress, []byte(CROSS_CHAIN_CONTRACT_MAP_REQUEST), []byte(params.CrossChainContractName)))
	return utils.BYTE_TRUE, nil
}
