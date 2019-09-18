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
)

type RegisterSideChainParam struct {
	Address      string
	ChainId      uint64
	Name         string
	BlocksToWait uint64
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteUint64(this.ChainId)
	sink.WriteVarBytes([]byte(this.Name))
	sink.WriteUint64(this.BlocksToWait)
	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize address error")
	}
	chainId, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	name, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize name error")
	}
	blocksToWait, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize blocksToWait error")
	}
	this.Address = address
	this.ChainId = chainId
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type ChainidParam struct {
	Chainid uint64
}

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.Chainid)
	return nil
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	this.Chainid = chainid
	return nil
}

type Asset struct {
	ChainId         uint64
	ContractAddress string
	Decimal         uint64
}

func (this *Asset) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.ChainId)
	sink.WriteVarBytes([]byte(this.ContractAddress))
	sink.WriteUint64(this.Decimal)
	return nil
}

func (this *Asset) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	contractAddress, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize contractAddress error")
	}
	this.ChainId = chainid
	decimal, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize decimal error")
	}
	this.ChainId = chainid
	this.ContractAddress = contractAddress
	this.Decimal = decimal
	return nil
}

type AssetMappingParam struct {
	Address   string
	AssetName string
	AssetList []*Asset
}

func (this *AssetMappingParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteVarBytes([]byte(this.AssetName))
	sink.WriteVarUint(uint64(len(this.AssetList)))
	for _, v := range this.AssetList {
		err := v.Serialization(sink)
		if err != nil {
			return fmt.Errorf("v.Serialization, serialize asset map error: %v", err)
		}
	}
	return nil
}

func (this *AssetMappingParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	assetName, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize assetName error")
	}
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize lenght error")
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
	sink.WriteVarBytes([]byte(this.AssetName))
	return nil
}

func (this *ApproveAssetMappingParam) Deserialization(source *common.ZeroCopySource) error {
	assetName, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize assetName error")
	}
	this.AssetName = assetName
	return nil
}
