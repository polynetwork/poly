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
	"sort"

	"github.com/ontio/multi-chain/common"
	"github.com/ontio/multi-chain/smartcontract/service/native/utils"
)

type RegisterSideChainParam struct {
	Address      common.Address
	Chainid      uint32
	Name         string
	BlocksToWait uint64
}

func (this *RegisterSideChainParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeAddress(sink, this.Address)
	if this.Chainid > math.MaxUint32 {
		return fmt.Errorf("chainid larger than max of uint32")
	}
	utils.EncodeVarUint(sink, uint64(this.Chainid))
	utils.EncodeString(sink, this.Name)
	utils.EncodeVarUint(sink, this.BlocksToWait)
	return nil
}

func (this *RegisterSideChainParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeString, deserialize address error: %v", err)
	}
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	if chainid > math.MaxUint32 {
		return fmt.Errorf("chainid larger than max of uint32")
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
	this.Chainid = uint32(chainid)
	this.Name = name
	this.BlocksToWait = blocksToWait
	return nil
}

type ChainidParam struct {
	Chainid uint32
}

func (this *ChainidParam) Serialization(sink *common.ZeroCopySink) error {
	if this.Chainid > math.MaxUint32 {
		return fmt.Errorf("chainid larger than max of uint32")
	}
	utils.EncodeVarUint(sink, uint64(this.Chainid))
	return nil
}

func (this *ChainidParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	if chainid > math.MaxUint32 {
		return fmt.Errorf("chainid larger than max of uint32")
	}
	this.Chainid = uint32(chainid)
	return nil
}



type AssetMappingParam struct {
	AssetList  []*AssetMap
}

func (this *AssetMappingParam) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, uint64(len(this.AssetMap)))

	for k, v := range this.AssetMap {
		assetList = append(assetList, v)
	}
	sort.SliceStable(peerPoolItemList, func(i, j int) bool {
		return peerPoolItemList[i].PeerPubkey > peerPoolItemList[j].PeerPubkey
	})
	for _, v := range peerPoolItemList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *AssetMappingParam) Deserialization(source *common.ZeroCopySource) error {
	chainid, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainid error: %v", err)
	}
	if chainid > math.MaxUint32 {
		return fmt.Errorf("chainid larger than max of uint32")
	}
	this.Chainid = uint32(chainid)
	return nil
}