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
	"github.com/ontio/multi-chain/common"
)

type RegisterPeerParam struct {
	PeerPubkey string
	Address    []byte
}

func (this *RegisterPeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address)
}

func (this *RegisterPeerParam) Deserialization(source *common.ZeroCopySource) error {
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

type PeerParam struct {
	PeerPubkey string
}

func (this *PeerParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
}

func (this *PeerParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, eof := source.NextString()
	if eof {
		return fmt.Errorf("source.NextString, deserialize peerPubkey error")
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type PeerParam2 struct {
	PeerPubkey string
	Address    []byte
}

func (this *PeerParam2) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address)
}

func (this *PeerParam2) Deserialization(source *common.ZeroCopySource) error {
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

type PeerListParam struct {
	PeerPubkeyList []string
}

func (this *PeerListParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarUint(uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
}

func (this *PeerListParam) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextVarUint()
	if eof {
		return fmt.Errorf("source.NextVarUint, deserialize PeerPubkeyList length error")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("source.NextString, deserialize peerPubkey error")
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	this.PeerPubkeyList = peerPubkeyList
	return nil
}

type Configuration struct {
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
}

func (this *Configuration) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.BlockMsgDelay)
	sink.WriteUint32(this.HashMsgDelay)
	sink.WriteUint32(this.PeerHandshakeTimeout)
	sink.WriteUint32(this.MaxBlockChangeView)
}

func (this *Configuration) Deserialization(source *common.ZeroCopySource) error {
	blockMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize blockMsgDelay error")
	}
	hashMsgDelay, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize hashMsgDelay error")
	}
	peerHandshakeTimeout, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize peerHandshakeTimeout error")
	}
	maxBlockChangeView, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("source.NextUint32, deserialize maxBlockChangeView error")
	}

	this.BlockMsgDelay = blockMsgDelay
	this.HashMsgDelay = hashMsgDelay
	this.PeerHandshakeTimeout = peerHandshakeTimeout
	this.MaxBlockChangeView = maxBlockChangeView
	return nil
}
