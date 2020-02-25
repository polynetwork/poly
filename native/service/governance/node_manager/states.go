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

package node_manager

import (
	"fmt"
	"io"
	"sort"

	"github.com/ontio/multi-chain/common"
)

type Status uint8

func (this *Status) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint8(uint8(*this))
}

func (this *Status) Deserialization(source *common.ZeroCopySource) error {
	status, eof := source.NextUint8()
	if eof {
		return fmt.Errorf("serialization.ReadUint8, deserialize status error: %v", io.ErrUnexpectedEOF)
	}
	*this = Status(status)
	return nil
}

type BlackListItem struct {
	PeerPubkey string //peerPubkey in black list
	Address    []byte //the owner of this peer
}

func (this *BlackListItem) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address)
}

func (this *BlackListItem) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPoolItem
}

func (this *PeerPoolMap) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.PeerPoolMap)))
	var peerPoolItemList []*PeerPoolItem
	for _, v := range this.PeerPoolMap {
		peerPoolItemList = append(peerPoolItemList, v)
	}
	sort.SliceStable(peerPoolItemList, func(i, j int) bool {
		return peerPoolItemList[i].PeerPubkey > peerPoolItemList[j].PeerPubkey
	})
	for _, v := range peerPoolItemList {
		v.Serialization(sink)
	}
}

func (this *PeerPoolMap) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize PeerPoolMap length error")
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint64(i) < n; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}
	this.PeerPoolMap = peerPoolMap
	return nil
}

type PeerPoolItem struct {
	Index      uint32 //peer index
	PeerPubkey string //peer pubkey
	Address    []byte //peer owner
	Status     Status
}

func (this *PeerPoolItem) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Index)
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address)
	this.Status.Serialization(sink)
}

func (this *PeerPoolItem) Deserialization(source *common.ZeroCopySource) error {
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize index error")
	}
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("source.NextVarBytes, deserialize address error")
	}
	status := new(Status)
	err := status.Deserialization(source)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}

	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.Status = *status
	return nil
}

type GovernanceView struct {
	View   uint32
	Height uint32
	TxHash common.Uint256
}

func (this *GovernanceView) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.View)
	sink.WriteUint32(this.Height)
	sink.WriteHash(this.TxHash)
}

func (this *GovernanceView) Deserialization(source *common.ZeroCopySource) error {
	view, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize view error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize height error")
	}
	txHash, eof := source.NextHash()
	if eof {
		return fmt.Errorf("source.NextHash, deserialize txHash error")
	}
	this.View = view
	this.Height = height
	this.TxHash = txHash
	return nil
}
