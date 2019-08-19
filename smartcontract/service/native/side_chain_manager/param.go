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
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type RegisterSideChainParam struct {
	Address      string
	Chainid      uint64
	Name         string
	BlocksToWait uint64
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeString(sink, this.Address)
	utils.EncodeVarUint(sink, this.Chainid)
	utils.EncodeString(sink, this.Name)
	utils.EncodeVarUint(sink, this.BlocksToWait)
	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize address error: %v", err)
	}
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	name, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize name error: %v", err)
	}
	blocksToWait, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize blocksToWait error: %v", err)
	}
	this.Address = address
	this.Chainid = chainid
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type ChainidParam struct {
	Chainid uint64
}

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.Chainid)
	return nil
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	this.Chainid = chainid
	return nil
}

type Asset struct {
	Chainid         uint64
	ContractAddress string
}

func (this *Asset) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.Chainid)
	utils.EncodeString(sink, this.ContractAddress)
	return nil
}

func (this *Asset) Deserialization(source *common.ZeroCopySource) error {
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	contractAddress, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize contractAddress error: %v", err)
	}
	this.Chainid = chainid
	this.ContractAddress = contractAddress
	return nil
}

type AssetMappingParam struct {
	Address   string
	AssetName string
	AssetList []*Asset
}

func (this *AssetMappingParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeString(sink, this.Address)
	utils.EncodeString(sink, this.AssetName)
	utils.EncodeVarUint(sink, uint64(len(this.AssetList)))
	for _, v := range this.AssetList {
		err := v.Serialization(sink)
		if err != nil {
			return fmt.Errorf("v.Serialization, serialize asset map error: %v", err)
		}
	}
	return nil
}

func (this *AssetMappingParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error: %v", err)
	}
	assetName, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize assetName error: %v", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize lenght error: %v", err)
	}
	assetList := make([]*Asset, 0)
	for i := 0; uint64(i) < n; i++ {
		asset := new(Asset)
		err := asset.Deserialization(source)
		if err != nil {
			return fmt.Errorf("assetMap.Deserialization, deserialize asset map error: %v", err)
		}
		assetList = append(assetList, asset)
	}
	this.Address = address
	this.AssetName = assetName
	this.AssetList = assetList
	return nil
}

type ApproveAssetMappingParam struct {
	AssetName string
}

func (this *ApproveAssetMappingParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeString(sink, this.AssetName)
	return nil
}

func (this *ApproveAssetMappingParam) Deserialization(source *common.ZeroCopySource) error {
	assetName, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize assetName error: %v", err)
	}
	this.AssetName = assetName
	return nil
}
