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

package chain_manager

import (
	"fmt"
	"io"

	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type Status uint8

func (this *Status) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint8(uint8(*this))
}

func (this *Status) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	*this = Status(status)
	return nil
}

type SideChain struct {
	ChainID            uint64 //side chain id
	Ratio              uint64 //side chain ong ratio(ong:ongx)
	Deposit            uint64 //side chain deposit
	OngNum             uint64 //side chain ong num
	OngPool            uint64 //side chain ong pool limit
	Status             Status //side chain status
	GenesisBlockHeader []byte //side chain genesis block
}

func (this *SideChain) Serialize(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ChainID)
	sink.WriteUint64(this.Ratio)
	sink.WriteUint64(this.Deposit)
	sink.WriteUint64(this.OngNum)
	sink.WriteUint64(this.OngPool)
	this.Status.Serialization(sink)
	sink.WriteVarBytes(this.GenesisBlockHeader)
}

func (this *SideChain) Deserialize(source *common.ZeroCopySource) error {
	chainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextString, deserialize chainID error: %v", io.ErrUnexpectedEOF)
	}
	ratio, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ratio error: %v", io.ErrUnexpectedEOF)
	}
	deposit, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize deposit error: %v", io.ErrUnexpectedEOF)
	}
	ongNum, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ongNum error: %v", io.ErrUnexpectedEOF)
	}
	ongPool, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("source.NextUint64, deserialize ongPool error: %v", io.ErrUnexpectedEOF)
	}
	status := new(Status)
	err := status.Deserialization(source)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}
	genesisBlockHeader, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return fmt.Errorf("source.NextVarBytes, deserialize genesisBlockHeader error: %v", common.ErrIrregularData)
	}
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize genesisBlockHeader error: %v", io.ErrUnexpectedEOF)
	}
	this.ChainID = chainID
	this.Ratio = ratio
	this.Deposit = deposit
	this.OngNum = ongNum
	this.OngPool = ongPool
	this.Status = *status
	this.GenesisBlockHeader = genesisBlockHeader
	return nil
}

type SideChainNodeInfo struct {
	ChainID     uint64
	NodeInfoMap map[string]*NodeToSideChainParams
}

func (this *SideChainNodeInfo) Serialization(sink *common.ZeroCopySink) error {
	utils.EncodeVarUint(sink, this.ChainID)
	utils.EncodeVarUint(sink, uint64(len(this.NodeInfoMap)))
	var nodeInfoMapList []*NodeToSideChainParams
	for _, v := range this.NodeInfoMap {
		nodeInfoMapList = append(nodeInfoMapList, v)
	}
	sort.SliceStable(nodeInfoMapList, func(i, j int) bool {
		return nodeInfoMapList[i].PeerPubkey > nodeInfoMapList[j].PeerPubkey
	})
	for _, v := range nodeInfoMapList {
		if err := v.Serialization(sink); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *SideChainNodeInfo) Deserialization(source *common.ZeroCopySource) error {
	chainID, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize chainID error: %v", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeVarUint, deserialize PeerPoolMap length error: %v", err)
	}
	nodeInfoMap := make(map[string]*NodeToSideChainParams)
	for i := 0; uint64(i) < n; i++ {
		nodeInfo := new(NodeToSideChainParams)
		if err := nodeInfo.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		nodeInfoMap[nodeInfo.PeerPubkey] = nodeInfo
	}
	this.ChainID = chainID
	this.NodeInfoMap = nodeInfoMap
	return nil
}
