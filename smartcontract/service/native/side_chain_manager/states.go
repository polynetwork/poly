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
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type SideChain struct {
	ChainId      uint64
	Name         string
	BlocksToWait uint64
}

func (this *SideChain) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainId)
	utils.EncodeString(sink, this.Name)
	utils.EncodeVarUint(sink, this.BlocksToWait)
	return nil
}

func (this *SideChain) Deserialization(source *common.ZeroCopySource) error {
	chainId, err := utils.DecodeVarUint(source)
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

	this.ChainId = chainId
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type AssetMap struct {
	AssetMap map[uint64]*Asset
}

func (this *AssetMap) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, uint64(len(this.AssetMap)))
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
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize length error: %v", err)
	}
	assetMap := make(map[uint64]*Asset)
	for i := 0; uint64(i) < n; i++ {
		asset := new(Asset)
		if err := asset.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize asset error: %v", err)
		}
		assetMap[asset.Chainid] = asset
	}
	this.AssetMap = assetMap
	return nil
}
