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
	"github.com/ontio/multi-chain/native/service/utils"
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
	name, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize name error: %v", err)
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

type AssetMap struct {
	AssetMap map[uint64]*Asset
}

func (this *AssetMap) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteUint64(uint64(len(this.AssetMap)))
	var assetList []*Asset
	for _, v := range this.AssetMap {
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

func (this *AssetMap) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize length error")
	}
	assetMap := make(map[uint64]*Asset)
	for i := 0; uint64(i) < n; i++ {
		asset := new(Asset)
		if err := asset.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize asset error: %v", err)
		}
		assetMap[asset.ChainId] = asset
	}
	this.AssetMap = assetMap
	return nil
}
