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
	"sort"

	"github.com/ontio/multi-chain/common"
)

type SideChain struct {
	ChainId      uint64
	Name         string
	BlocksToWait uint64
}

func (this *SideChain) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(this.ChainId)
	sink.WriteString(this.Name)
	sink.WriteUint64(this.BlocksToWait)
	return nil
}

func (this *SideChain) Deserialization(source *common.ZeroCopySource) error {
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

	this.ChainId = chainId
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type CrossChainContractMap struct {
	CrossChainContractMap map[uint64]*CrossChainContract
}

func (this *CrossChainContractMap) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(uint64(len(this.CrossChainContractMap)))
	var assetList []*CrossChainContract
	for _, v := range this.CrossChainContractMap {
		assetList = append(assetList, v)
	}
	sort.SliceStable(assetList, func(i, j int) bool {
		return assetList[i].ContractAddress > assetList[j].ContractAddress
	})
	for _, v := range assetList {
		if err := v.Serialization(sink); err != nil {
			return fmt.Errorf("serialize asset error: %v", err)
		}
	}
	return nil
}

func (this *CrossChainContractMap) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize length error")
	}
	crossChainContractMap := make(map[uint64]*CrossChainContract)
	for i := 0; uint64(i) < n; i++ {
		asset := new(CrossChainContract)
		if err := asset.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize asset error: %v", err)
		}
		crossChainContractMap[asset.ChainId] = asset
	}
	this.CrossChainContractMap = crossChainContractMap
	return nil
}
