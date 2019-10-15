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

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.Chainid)
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	this.Chainid = chainid
	return nil
}

type CrossChainContract struct {
	ChainId         uint64
	ContractAddress string
}

func (this *CrossChainContract) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.ChainId)
	sink.WriteVarBytes([]byte(this.ContractAddress))
	return nil
}

func (this *CrossChainContract) Deserialization(source *common.ZeroCopySource) error {
	chainid, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error")
	}
	contractAddress, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize contractAddress error")
	}
	this.ChainId = chainid
	this.ContractAddress = contractAddress
	return nil
}

type CrossChainContractMappingParam struct {
	Address                string
	CrossChainContractName string
	CrossChainContractList []*CrossChainContract
}

func (this *CrossChainContractMappingParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes([]byte(this.Address))
	sink.WriteVarBytes([]byte(this.CrossChainContractName))
	sink.WriteVarUint(uint64(len(this.CrossChainContractList)))
	for _, v := range this.CrossChainContractList {
		err := v.Serialization(sink)
		if err != nil {
			return fmt.Errorf("v.Serialization, serialize asset map error: %v", err)
		}
	}
	return nil
}

func (this *CrossChainContractMappingParam) Deserialization(source *common.ZeroCopySource) error {
	address, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error")
	}
	crossChainContractName, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize crossChainContractName error")
	}
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize lenght error")
	}
	assetList := make([]*CrossChainContract, 0)
	for i := 0; uint64(i) < n; i++ {
		asset := new(CrossChainContract)
		err := asset.Deserialization(source)
		if err != nil {
			return fmt.Errorf("assetMap.Deserialization, deserialize asset map error: %v", err)
		}
		assetList = append(assetList, asset)
	}
	this.Address = address
	this.CrossChainContractName = crossChainContractName
	this.CrossChainContractList = assetList
	return nil
}

type ApproveCrossChainContractMappingParam struct {
	CrossChainContractName string
}

func (this *ApproveCrossChainContractMappingParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteVarBytes([]byte(this.CrossChainContractName))
	return nil
}

func (this *ApproveCrossChainContractMappingParam) Deserialization(source *common.ZeroCopySource) error {
	assetName, eof := source.NextString()
	if eof {
		return fmt.Errorf("utils.DecodeString, deserialize assetName error")
	}
	this.CrossChainContractName = assetName
	return nil
}
